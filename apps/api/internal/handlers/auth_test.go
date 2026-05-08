package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func postJSON(body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

// ── Login ─────────────────────────────────────────────────────────────────────

func TestLogin_MissingFields(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	c, w := postJSON(`{}`)
	h.Login(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_WrongCredentials(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	c, w := postJSON(`{"username":"bad","password":"wrong"}`)
	h.Login(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogin_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	mock.ExpectQuery(`SELECT .* FROM "users"`).WillReturnError(errors.New("db down"))
	c, w := postJSON(`{"username":"admin","password":"admin"}`)
	h.Login(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestLogin_UserNotFound_CreateSuccess(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	mock.ExpectQuery(`SELECT .* FROM "users"`).WillReturnRows(sqlmock.NewRows(nil))
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	c, w := postJSON(`{"username":"admin","password":"admin"}`)
	h.Login(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "token")
}

func TestLogin_UserNotFound_CreateError(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	mock.ExpectQuery(`SELECT .* FROM "users"`).WillReturnRows(sqlmock.NewRows(nil))
	mock.ExpectQuery(`INSERT INTO "users"`).WillReturnError(errors.New("insert failed"))
	c, w := postJSON(`{"username":"admin","password":"admin"}`)
	h.Login(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestLogin_UserExists(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	uid := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "email", "google_id", "name", "avatar_url"}).
		AddRow(uid, "admin@mtracker.local", "", "Admin", "")
	mock.ExpectQuery(`SELECT .* FROM "users"`).WillReturnRows(rows)
	c, w := postJSON(`{"username":"admin","password":"admin"}`)
	h.Login(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "token")
}

// ── GetProfile ────────────────────────────────────────────────────────────────

func TestGetProfile_Found(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	uid := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "email", "name"}).
		AddRow(uid, "user@example.com", "User")
	mock.ExpectQuery(`SELECT .* FROM "users"`).WillReturnRows(rows)
	c, w := authedContext(uid)
	h.GetProfile(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetProfile_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	mock.ExpectQuery(`SELECT .* FROM "users"`).WillReturnRows(sqlmock.NewRows(nil))
	c, w := authedContext(uuid.New())
	h.GetProfile(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ── GoogleSignIn ──────────────────────────────────────────────────────────────

func mockGoogleServer(t *testing.T, status int, body string) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}))
	orig := googleUserInfoURL
	googleUserInfoURL = ts.URL
	t.Cleanup(func() {
		ts.Close()
		googleUserInfoURL = orig
	})
}

func withFailingSigner(t *testing.T) {
	t.Helper()
	orig := jwtSigner
	jwtSigner = func(_ *jwt.Token, _ interface{}) (string, error) {
		return "", errors.New("signing failed")
	}
	t.Cleanup(func() { jwtSigner = orig })
}

func TestLogin_JWTSignError(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	withFailingSigner(t)
	uid := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "email", "google_id", "name", "avatar_url"}).
		AddRow(uid, "admin@mtracker.local", "", "Admin", "")
	mock.ExpectQuery(`SELECT .* FROM "users"`).WillReturnRows(rows)
	c, w := postJSON(`{"username":"admin","password":"admin"}`)
	h.Login(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGoogleSignIn_JWTSignError(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	withFailingSigner(t)
	body, _ := json.Marshal(map[string]string{"sub": "gid1", "email": "g@ex.com", "name": "G"})
	mockGoogleServer(t, http.StatusOK, string(body))
	uid := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "email", "google_id", "name", "avatar_url"}).
		AddRow(uid, "g@ex.com", "gid1", "G", "")
	mock.ExpectQuery(`SELECT .* FROM "users"`).WillReturnRows(rows)
	c, w := postJSON(`{"access_token":"tok"}`)
	h.GoogleSignIn(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGoogleSignIn_HTTPGetError(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	orig := googleUserInfoURL
	googleUserInfoURL = "http://127.0.0.1:1" // nothing listening — connection refused
	t.Cleanup(func() { googleUserInfoURL = orig })
	c, w := postJSON(`{"access_token":"tok"}`)
	h.GoogleSignIn(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGoogleSignIn_ReadBodyError(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	// Hijack the connection: send 200 headers with Content-Length but close before body.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "no hijack", 500)
			return
		}
		conn, buf, _ := hj.Hijack()
		buf.WriteString("HTTP/1.1 200 OK\r\nContent-Length: 100\r\n\r\n")
		buf.Flush()
		conn.Close() // close before sending the promised 100 bytes
	}))
	t.Cleanup(ts.Close)
	orig := googleUserInfoURL
	googleUserInfoURL = ts.URL
	t.Cleanup(func() { googleUserInfoURL = orig })
	c, w := postJSON(`{"access_token":"tok"}`)
	h.GoogleSignIn(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGoogleSignIn_MissingToken(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	c, w := postJSON(`{}`)
	h.GoogleSignIn(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGoogleSignIn_GoogleNon200(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	mockGoogleServer(t, http.StatusUnauthorized, `{}`)
	c, w := postJSON(`{"access_token":"bad"}`)
	h.GoogleSignIn(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGoogleSignIn_BadJSON(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	mockGoogleServer(t, http.StatusOK, `not-json`)
	c, w := postJSON(`{"access_token":"tok"}`)
	h.GoogleSignIn(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGoogleSignIn_EmptySub(t *testing.T) {
	db, _ := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	body, _ := json.Marshal(map[string]string{"email": "a@b.com"})
	mockGoogleServer(t, http.StatusOK, string(body))
	c, w := postJSON(`{"access_token":"tok"}`)
	h.GoogleSignIn(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGoogleSignIn_NewUser_CreateSuccess(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	body, _ := json.Marshal(map[string]string{"sub": "gid1", "email": "g@ex.com", "name": "G"})
	mockGoogleServer(t, http.StatusOK, string(body))
	mock.ExpectQuery(`SELECT .* FROM "users"`).WillReturnRows(sqlmock.NewRows(nil))
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	c, w := postJSON(`{"access_token":"tok"}`)
	h.GoogleSignIn(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGoogleSignIn_NewUser_CreateError(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	body, _ := json.Marshal(map[string]string{"sub": "gid1", "email": "g@ex.com", "name": "G"})
	mockGoogleServer(t, http.StatusOK, string(body))
	mock.ExpectQuery(`SELECT .* FROM "users"`).WillReturnRows(sqlmock.NewRows(nil))
	mock.ExpectQuery(`INSERT INTO "users"`).WillReturnError(errors.New("insert failed"))
	c, w := postJSON(`{"access_token":"tok"}`)
	h.GoogleSignIn(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGoogleSignIn_ExistingUser(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	body, _ := json.Marshal(map[string]string{"sub": "gid1", "email": "g@ex.com", "name": "G"})
	mockGoogleServer(t, http.StatusOK, string(body))
	uid := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "email", "google_id", "name", "avatar_url"}).
		AddRow(uid, "g@ex.com", "gid1", "G", "")
	mock.ExpectQuery(`SELECT .* FROM "users"`).WillReturnRows(rows)
	c, w := postJSON(`{"access_token":"tok"}`)
	h.GoogleSignIn(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGoogleSignIn_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	h := NewAuthHandler(db, "secret")
	body, _ := json.Marshal(map[string]string{"sub": "gid1", "email": "g@ex.com", "name": "G"})
	mockGoogleServer(t, http.StatusOK, string(body))
	mock.ExpectQuery(`SELECT .* FROM "users"`).WillReturnError(errors.New("db down"))
	c, w := postJSON(`{"access_token":"tok"}`)
	h.GoogleSignIn(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
