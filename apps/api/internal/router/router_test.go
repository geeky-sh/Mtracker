package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/aash/mtracker/apps/api/internal/config"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	sqlDB, _, err := sqlmock.New()
	require.NoError(t, err)
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	return db
}

func TestSetup_RegistersRoutes(t *testing.T) {
	db := newTestDB(t)
	cfg := &config.Config{JWTSecret: "test", GoogleClientID: "cid"}

	engine := Setup(db, cfg)
	require.NotNil(t, engine)

	routes := engine.Routes()
	routeMap := make(map[string]bool)
	for _, r := range routes {
		routeMap[r.Method+":"+r.Path] = true
	}

	assert.True(t, routeMap["GET:/health"])
	assert.True(t, routeMap["POST:/api/v1/auth/login"])
	assert.True(t, routeMap["GET:/api/v1/profile"])
	assert.True(t, routeMap["GET:/api/v1/activities"])
	assert.True(t, routeMap["POST:/api/v1/activities"])
	assert.True(t, routeMap["GET:/api/v1/activities/search"])
	assert.True(t, routeMap["DELETE:/api/v1/activities/:id"])
	assert.True(t, routeMap["GET:/api/v1/logs"])
	assert.True(t, routeMap["POST:/api/v1/logs"])
	assert.True(t, routeMap["DELETE:/api/v1/logs/:id"])
	assert.True(t, routeMap["GET:/api/v1/analytics"])
}

func TestSetup_HealthCheck(t *testing.T) {
	db := newTestDB(t)
	engine := Setup(db, &config.Config{JWTSecret: "test", GoogleClientID: "cid"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetup_OPTIONSReturns204(t *testing.T) {
	db := newTestDB(t)
	engine := Setup(db, &config.Config{JWTSecret: "test", GoogleClientID: "cid"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/activities", nil)
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}
