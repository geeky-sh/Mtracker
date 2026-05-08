package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/aash/mtracker/apps/api/internal/middleware"
	"github.com/aash/mtracker/apps/api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Predefined palette — colours are assigned cyclically based on activity count.
var palette = []string{
	"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4",
	"#F7DC6F", "#DDA0DD", "#98D8C8", "#F0B27A",
	"#BB8FCE", "#85C1E9", "#82E0AA", "#F1948A",
}

type ActivitiesHandler struct {
	db *gorm.DB
}

func NewActivitiesHandler(db *gorm.DB) *ActivitiesHandler {
	return &ActivitiesHandler{db: db}
}

type createActivityRequest struct {
	Name        string `json:"name"        binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
}

// List returns all activities for the authenticated user, each with its last logged date.
func (h *ActivitiesHandler) List(c *gin.Context) {
	userID := middleware.UserIDFromCtx(c)
	var activities []models.Activity
	if err := h.db.Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch activities"})
		return
	}

	if len(activities) > 0 {
		type lastLog struct {
			ActivityID string
			MaxDate    *time.Time
		}
		var rows []lastLog
		ids := make([]string, len(activities))
		for i, a := range activities {
			ids[i] = a.ID.String()
		}
		h.db.Raw(`SELECT activity_id::text, MAX(logged_date) AS max_date FROM activity_logs WHERE activity_id::text IN ? AND user_id = ? GROUP BY activity_id`, ids, userID).Scan(&rows)

		dateMap := make(map[string]*time.Time, len(rows))
		for _, r := range rows {
			r := r
			dateMap[r.ActivityID] = r.MaxDate
		}
		for i := range activities {
			activities[i].LastLoggedDate = dateMap[activities[i].ID.String()]
		}
	}

	c.JSON(http.StatusOK, activities)
}

// Create adds a new activity and auto-assigns a colour.
func (h *ActivitiesHandler) Create(c *gin.Context) {
	userID := middleware.UserIDFromCtx(c)

	var req createActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var count int64
	h.db.Model(&models.Activity{}).Where("user_id = ?", userID).Count(&count)

	activity := models.Activity{
		UserID:      userID,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Color:       palette[int(count)%len(palette)],
	}

	if err := h.db.Create(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create activity"})
		return
	}
	c.JSON(http.StatusCreated, activity)
}

// Delete removes an activity and all its associated logs.
func (h *ActivitiesHandler) Delete(c *gin.Context) {
	userID := middleware.UserIDFromCtx(c)
	activityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid activity id"})
		return
	}

	var activity models.Activity
	if err := h.db.First(&activity, "id = ? AND user_id = ?", activityID, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "activity not found"})
		return
	}

	if err := h.db.Where("activity_id = ? AND user_id = ?", activityID, userID).Delete(&models.ActivityLog{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete logs"})
		return
	}

	if err := h.db.Delete(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete activity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// Search returns activities whose names contain the query string (case-insensitive).
// Used to surface similar activities while the user is typing.
func (h *ActivitiesHandler) Search(c *gin.Context) {
	userID := middleware.UserIDFromCtx(c)
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusOK, []models.Activity{})
		return
	}

	var activities []models.Activity
	if err := h.db.Where("user_id = ? AND name ILIKE ?", userID, "%"+q+"%").
		Order("name ASC").
		Limit(5).
		Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}
	c.JSON(http.StatusOK, activities)
}
