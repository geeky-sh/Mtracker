package handlers

import (
	"net/http"
	"time"

	"github.com/aash/mtracker/apps/api/internal/middleware"
	"github.com/aash/mtracker/apps/api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LogsHandler struct {
	db *gorm.DB
}

func NewLogsHandler(db *gorm.DB) *LogsHandler {
	return &LogsHandler{db: db}
}

type createLogRequest struct {
	ActivityID string `json:"activity_id" binding:"required"`
	LoggedDate string `json:"logged_date" binding:"required"` // YYYY-MM-DD
}

// Create records that an activity happened on a given date.
// Only today or the previous 2 days are accepted.
func (h *LogsHandler) Create(c *gin.Context) {
	userID := middleware.UserIDFromCtx(c)

	var req createLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	activityID, err := uuid.Parse(req.ActivityID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid activity_id"})
		return
	}

	loggedDate, err := time.Parse("2006-01-02", req.LoggedDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "logged_date must be YYYY-MM-DD"})
		return
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	earliest := today.AddDate(0, 0, -30)
	if loggedDate.Before(earliest) || loggedDate.After(today) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date must be today or within the past 30 days"})
		return
	}

	// Verify the activity belongs to this user
	var activity models.Activity
	if err := h.db.First(&activity, "id = ? AND user_id = ?", activityID, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "activity not found"})
		return
	}

	log := models.ActivityLog{
		ActivityID: activityID,
		UserID:     userID,
		LoggedDate: loggedDate,
	}

	if err := h.db.Create(&log).Error; err != nil {
		// Unique constraint violation means it was already logged
		c.JSON(http.StatusConflict, gin.H{"error": "activity already logged for this date"})
		return
	}
	c.JSON(http.StatusCreated, log)
}

// ListByActivity returns all log dates for a given activity (used by the analytics calendar).
func (h *LogsHandler) ListByActivity(c *gin.Context) {
	userID := middleware.UserIDFromCtx(c)
	activityIDStr := c.Query("activity_id")
	if activityIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "activity_id query param is required"})
		return
	}

	activityID, err := uuid.Parse(activityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid activity_id"})
		return
	}

	var logs []models.ActivityLog
	if err := h.db.Where("activity_id = ? AND user_id = ?", activityID, userID).
		Order("logged_date ASC").
		Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch logs"})
		return
	}
	c.JSON(http.StatusOK, logs)
}

// Delete removes a specific log entry.
func (h *LogsHandler) Delete(c *gin.Context) {
	userID := middleware.UserIDFromCtx(c)
	logID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid log id"})
		return
	}

	result := h.db.Where("id = ? AND user_id = ?", logID, userID).Delete(&models.ActivityLog{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete log"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "log not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
