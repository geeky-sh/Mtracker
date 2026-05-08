package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/aash/mtracker/apps/api/internal/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AnalyticsHandler struct {
	db *gorm.DB
}

func NewAnalyticsHandler(db *gorm.DB) *AnalyticsHandler {
	return &AnalyticsHandler{db: db}
}

type ActivitySummary struct {
	ActivityID string `json:"activity_id"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	Count      int    `json:"count"`
}

func (h *AnalyticsHandler) Summary(c *gin.Context) {
	userID := middleware.UserIDFromCtx(c)

	days, err := strconv.Atoi(c.DefaultQuery("days", "30"))
	if err != nil || days <= 0 || days > 365 {
		days = 30
	}

	since := time.Now().UTC().Truncate(24 * time.Hour).AddDate(0, 0, -days)

	var results []ActivitySummary
	if err := h.db.Raw(`
		SELECT a.id::text AS activity_id, a.name, a.color, COUNT(l.id) AS count
		FROM activities a
		LEFT JOIN activity_logs l ON l.activity_id = a.id
			AND l.logged_date >= ?
			AND l.user_id = ?
		WHERE a.user_id = ?
		GROUP BY a.id, a.name, a.color
		ORDER BY count DESC
	`, since, userID, userID).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch analytics"})
		return
	}

	c.JSON(http.StatusOK, results)
}
