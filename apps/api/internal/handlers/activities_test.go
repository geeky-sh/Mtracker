package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func activityRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "user_id", "name", "description", "color", "created_at", "updated_at"})
}

func addActivity(rows *sqlmock.Rows, id, userID uuid.UUID, name, color string) *sqlmock.Rows {
	return rows.AddRow(id, userID, name, "", color, time.Now(), time.Now())
}

func jsonPostCtx(userID uuid.UUID, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", userID)
	return c, w
}

// ── List ──────────────────────────────────────────────────────────────────────

func TestList_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewActivitiesHandler(db)
	mock.ExpectQuery(`SELECT .* FROM "activities"`).WillReturnError(errors.New("db error"))
	c, w := authedContext(uuid.New())
	h.List(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestList_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewActivitiesHandler(db)
	mock.ExpectQuery(`SELECT .* FROM "activities"`).WillReturnRows(activityRows())
	c, w := authedContext(uuid.New())
	h.List(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestList_WithActivities(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewActivitiesHandler(db)
	uid, aid := uuid.New(), uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "activities"`).
		WillReturnRows(addActivity(activityRows(), aid, uid, "Run", "#FF0000"))
	mock.ExpectQuery(`SELECT activity_id`).
		WillReturnRows(sqlmock.NewRows([]string{"activity_id", "max_date"}))
	c, w := authedContext(uid)
	h.List(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Run")
}

func TestList_WithActivities_DateEnriched(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewActivitiesHandler(db)
	uid, aid := uuid.New(), uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "activities"`).
		WillReturnRows(addActivity(activityRows(), aid, uid, "Run", "#FF0000"))
	lastDate := time.Now()
	mock.ExpectQuery(`SELECT activity_id`).
		WillReturnRows(sqlmock.NewRows([]string{"activity_id", "max_date"}).
			AddRow(aid.String(), lastDate))
	c, w := authedContext(uid)
	h.List(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Run")
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestCreate_InvalidBody(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewActivitiesHandler(db)
	c, w := jsonPostCtx(uuid.New(), `{}`)
	h.Create(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreate_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewActivitiesHandler(db)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "activities"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`INSERT INTO "activities"`).WillReturnError(errors.New("insert failed"))
	c, w := jsonPostCtx(uuid.New(), `{"name":"Run"}`)
	h.Create(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreate_Success(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewActivitiesHandler(db)
	uid := uuid.New()
	mock.ExpectQuery(`SELECT count\(\*\) FROM "activities"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(`INSERT INTO "activities"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	c, w := jsonPostCtx(uid, `{"name":"Run","description":"daily"}`)
	h.Create(c)
	assert.Equal(t, http.StatusCreated, w.Code)
}

// ── Search ────────────────────────────────────────────────────────────────────

func TestSearch_EmptyQuery(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewActivitiesHandler(db)
	c, w := authedContext(uuid.New())
	c.Request = httptest.NewRequest(http.MethodGet, "/activities/search?q=", nil)
	h.Search(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", w.Body.String())
}

func TestSearch_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewActivitiesHandler(db)
	mock.ExpectQuery(`SELECT .* FROM "activities"`).WillReturnError(errors.New("db error"))
	c, w := authedContext(uuid.New())
	c.Request = httptest.NewRequest(http.MethodGet, "/activities/search?q=run", nil)
	h.Search(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSearch_Success(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewActivitiesHandler(db)
	uid := uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "activities"`).
		WillReturnRows(addActivity(activityRows(), uuid.New(), uid, "Running", "#FF0000"))
	c, w := authedContext(uid)
	c.Request = httptest.NewRequest(http.MethodGet, "/activities/search?q=run", nil)
	h.Search(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Running")
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestDeleteActivity_InvalidUUID(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewActivitiesHandler(db)
	c, w := authedContext(uuid.New())
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}
	h.Delete(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteActivity_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewActivitiesHandler(db)
	mock.ExpectQuery(`SELECT .* FROM "activities"`).WillReturnRows(activityRows())
	c, w := authedContext(uuid.New())
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
	h.Delete(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteActivity_DeleteLogsError(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewActivitiesHandler(db)
	uid, aid := uuid.New(), uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "activities"`).
		WillReturnRows(addActivity(activityRows(), aid, uid, "Run", "#FF0000"))
	mock.ExpectExec(`DELETE FROM "activity_logs"`).WillReturnError(errors.New("del failed"))
	c, w := authedContext(uid)
	c.Params = gin.Params{{Key: "id", Value: aid.String()}}
	h.Delete(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteActivity_DeleteActivityError(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewActivitiesHandler(db)
	uid, aid := uuid.New(), uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "activities"`).
		WillReturnRows(addActivity(activityRows(), aid, uid, "Run", "#FF0000"))
	mock.ExpectExec(`DELETE FROM "activity_logs"`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`DELETE FROM "activities"`).WillReturnError(errors.New("del failed"))
	c, w := authedContext(uid)
	c.Params = gin.Params{{Key: "id", Value: aid.String()}}
	h.Delete(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteActivity_Success(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewActivitiesHandler(db)
	uid, aid := uuid.New(), uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "activities"`).
		WillReturnRows(addActivity(activityRows(), aid, uid, "Run", "#FF0000"))
	mock.ExpectExec(`DELETE FROM "activity_logs"`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`DELETE FROM "activities"`).WillReturnResult(sqlmock.NewResult(0, 1))
	c, w := authedContext(uid)
	c.Params = gin.Params{{Key: "id", Value: aid.String()}}
	h.Delete(c)
	assert.Equal(t, http.StatusOK, w.Code)
}
