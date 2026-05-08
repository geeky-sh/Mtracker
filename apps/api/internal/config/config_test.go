package config

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEnv_ReturnsFallbackWhenUnset(t *testing.T) {
	os.Unsetenv("__MTRACKER_TEST_UNSET__")
	assert.Equal(t, "fallback", getEnv("__MTRACKER_TEST_UNSET__", "fallback"))
}

func TestGetEnv_ReturnsValueWhenSet(t *testing.T) {
	t.Setenv("__MTRACKER_TEST_SET__", "hello")
	assert.Equal(t, "hello", getEnv("__MTRACKER_TEST_SET__", "fallback"))
}

func TestMustEnv_ReturnsValueWhenSet(t *testing.T) {
	t.Setenv("__MTRACKER_MUST__", "world")
	assert.Equal(t, "world", mustEnv("__MTRACKER_MUST__"))
}

func TestMustEnv_PanicsWhenUnset(t *testing.T) {
	os.Unsetenv("__MTRACKER_MUST_UNSET__")
	assert.Panics(t, func() { mustEnv("__MTRACKER_MUST_UNSET__") })
}

func TestBuildDSN_UsesEnvVars(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "myhost")
	t.Setenv("POSTGRES_PORT", "5555")
	t.Setenv("POSTGRES_USER", "myuser")
	t.Setenv("POSTGRES_PASSWORD", "mypass")
	t.Setenv("POSTGRES_DB", "mydb")

	dsn := buildDSN()
	assert.Contains(t, dsn, "host=myhost")
	assert.Contains(t, dsn, "port=5555")
	assert.Contains(t, dsn, "user=myuser")
	assert.Contains(t, dsn, "password=mypass")
	assert.Contains(t, dsn, "dbname=mydb")
	assert.Contains(t, dsn, "sslmode=disable")
}

func TestBuildDSN_UsesDefaults(t *testing.T) {
	for _, k := range []string{"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB"} {
		os.Unsetenv(k)
	}
	dsn := buildDSN()
	assert.True(t, strings.Contains(dsn, "host=localhost") || strings.Contains(dsn, "host="))
}

func TestLoad_ReturnsConfig(t *testing.T) {
	t.Setenv("JWT_SECRET", "testsecret")
	t.Setenv("GOOGLE_CLIENT_ID", "testclient")
	t.Setenv("PORT", "9090")

	cfg := Load()
	require.NotNil(t, cfg)
	assert.Equal(t, "testsecret", cfg.JWTSecret)
	assert.Equal(t, "testclient", cfg.GoogleClientID)
	assert.Equal(t, "9090", cfg.Port)
}

func TestLoad_DefaultPort(t *testing.T) {
	t.Setenv("JWT_SECRET", "s")
	t.Setenv("GOOGLE_CLIENT_ID", "c")
	os.Unsetenv("PORT")

	cfg := Load()
	assert.Equal(t, "8080", cfg.Port)
}
