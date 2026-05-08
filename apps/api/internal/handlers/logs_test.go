package handlers

import (
	"errors"
	"fmt"
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

func logRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "activity_id", "user_id", "logged_date", "created_at"})
}

func logPostCtx(userID uuid.UUID, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", userID)
	return c, w
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestCreateLog_MissingFields(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewLogsHandler(db)
	c, w := logPostCtx(uuid.New(), `{}`)
	h.Create(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateLog_InvalidActivityID(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewLogsHandler(db)
	c, w := logPostCtx(uuid.New(), `{"activity_id":"bad","logged_date":"2026-04-01"}`)
	h.Create(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateLog_InvalidDate(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewLogsHandler(db)
	body := fmt.Sprintf(`{"activity_id":%q,"logged_date":"not-a-date"}`, uuid.New())
	c, w := logPostCtx(uuid.New(), body)
	h.Create(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateLog_DateTooOld(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewLogsHandler(db)
	old := time.Now().AddDate(0, 0, -60).Format("2006-01-02")
	body := fmt.Sprintf(`{"activity_id":%q,"logged_date":%q}`, uuid.New(), old)
	c, w := logPostCtx(uuid.New(), body)
	h.Create(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateLog_DateFuture(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewLogsHandler(db)
	future := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	body := fmt.Sprintf(`{"activity_id":%q,"logged_date":%q}`, uuid.New(), future)
	c, w := logPostCtx(uuid.New(), body)
	h.Create(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateLog_ActivityNotFound(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewLogsHandler(db)
	mock.ExpectQuery(`SELECT .* FROM "activities"`).WillReturnRows(activityRows())
	today := time.Now().Format("2006-01-02")
	body := fmt.Sprintf(`{"activity_id":%q,"logged_date":%q}`, uuid.New(), today)
	c, w := logPostCtx(uuid.New(), body)
	h.Create(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateLog_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewLogsHandler(db)
	uid, aid := uuid.New(), uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "activities"`).
		WillReturnRows(activityRows().AddRow(aid, uid, "Run", "", "#FF0000", time.Now(), time.Now()))
	mock.ExpectQuery(`INSERT INTO "activity_logs"`).WillReturnError(errors.New("unique"))
	today := time.Now().Format("2006-01-02")
	body := fmt.Sprintf(`{"activity_id":%q,"logged_date":%q}`, aid, today)
	c, w := logPostCtx(uid, body)
	h.Create(c)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCreateLog_Success(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewLogsHandler(db)
	uid, aid := uuid.New(), uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "activities"`).
		WillReturnRows(activityRows().AddRow(aid, uid, "Run", "", "#FF0000", time.Now(), time.Now()))
	mock.ExpectQuery(`INSERT INTO "activity_logs"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	today := time.Now().Format("2006-01-02")
	body := fmt.Sprintf(`{"activity_id":%q,"logged_date":%q}`, aid, today)
	c, w := logPostCtx(uid, body)
	h.Create(c)
	assert.Equal(t, http.StatusCreated, w.Code)
}

// ── ListByActivity ────────────────────────────────────────────────────────────

func TestListLogs_MissingParam(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewLogsHandler(db)
	c, w := authedContext(uuid.New())
	c.Request = httptest.NewRequest(http.MethodGet, "/logs", nil)
	h.ListByActivity(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListLogs_InvalidUUID(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewLogsHandler(db)
	c, w := authedContext(uuid.New())
	c.Request = httptest.NewRequest(http.MethodGet, "/logs?activity_id=bad", nil)
	h.ListByActivity(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListLogs_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewLogsHandler(db)
	mock.ExpectQuery(`SELECT .* FROM "activity_logs"`).WillReturnError(errors.New("db error"))
	c, w := authedContext(uuid.New())
	c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/logs?activity_id=%s", uuid.New()), nil)
	h.ListByActivity(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListLogs_Success(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewLogsHandler(db)
	uid, aid, lid := uuid.New(), uuid.New(), uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "activity_logs"`).
		WillReturnRows(logRows().AddRow(lid, aid, uid, time.Now(), time.Now()))
	c, w := authedContext(uid)
	c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/logs?activity_id=%s", aid), nil)
	h.ListByActivity(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestDeleteLog_InvalidUUID(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewLogsHandler(db)
	c, w := authedContext(uuid.New())
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}
	h.Delete(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteLog_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewLogsHandler(db)
	mock.ExpectExec(`DELETE FROM "activity_logs"`).WillReturnError(errors.New("db error"))
	c, w := authedContext(uuid.New())
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
	h.Delete(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteLog_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewLogsHandler(db)
	mock.ExpectExec(`DELETE FROM "activity_logs"`).WillReturnResult(sqlmock.NewResult(0, 0))
	c, w := authedContext(uuid.New())
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
	h.Delete(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteLog_Success(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewLogsHandler(db)
	lid := uuid.New()
	mock.ExpectExec(`DELETE FROM "activity_logs"`).WillReturnResult(sqlmock.NewResult(0, 1))
	c, w := authedContext(uuid.New())
	c.Params = gin.Params{{Key: "id", Value: lid.String()}}
	h.Delete(c)
	assert.Equal(t, http.StatusOK, w.Code)
}
