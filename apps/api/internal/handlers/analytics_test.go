package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSummary_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewAnalyticsHandler(db)
	mock.ExpectQuery(`SELECT`).WillReturnError(errors.New("db error"))
	c, w := authedContext(uuid.New())
	c.Request = httptest.NewRequest(http.MethodGet, "/analytics?days=30", nil)
	h.Summary(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSummary_Success(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewAnalyticsHandler(db)
	rows := sqlmock.NewRows([]string{"activity_id", "name", "color", "count"}).
		AddRow(uuid.New().String(), "Run", "#FF0000", 5)
	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)
	c, w := authedContext(uuid.New())
	c.Request = httptest.NewRequest(http.MethodGet, "/analytics?days=30", nil)
	h.Summary(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Run")
}

func TestSummary_InvalidDays_FallsBackTo30(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewAnalyticsHandler(db)
	rows := sqlmock.NewRows([]string{"activity_id", "name", "color", "count"})
	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)
	c, w := authedContext(uuid.New())
	// "abc" is not a valid integer → falls back to 30
	c.Request = httptest.NewRequest(http.MethodGet, "/analytics?days=abc", nil)
	h.Summary(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSummary_NegativeDays_FallsBackTo30(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewAnalyticsHandler(db)
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"activity_id", "name", "color", "count"}))
	c, w := authedContext(uuid.New())
	c.Request = httptest.NewRequest(http.MethodGet, "/analytics?days=-5", nil)
	h.Summary(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSummary_TooManyDays_FallsBackTo30(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewAnalyticsHandler(db)
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"activity_id", "name", "color", "count"}))
	c, w := authedContext(uuid.New())
	c.Request = httptest.NewRequest(http.MethodGet, "/analytics?days=999", nil)
	h.Summary(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

// Cover the gin.Context path when c.Request is set with query params via gin params
func TestSummary_DefaultQuery(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewAnalyticsHandler(db)
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"activity_id", "name", "color", "count"}))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/analytics", nil) // no days param → default 30
	c.Set("user_id", uuid.New())
	h.Summary(c)
	assert.Equal(t, http.StatusOK, w.Code)
}
