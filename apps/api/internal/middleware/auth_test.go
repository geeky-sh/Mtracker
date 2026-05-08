package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func init() { gin.SetMode(gin.TestMode) }

func makeValidToken(secret string, userID uuid.UUID, email string) string {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	return tok
}

func runAuth(secret, authHeader string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, engine := gin.CreateTestContext(w)
	engine.Use(Auth(secret))
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	c.Request = req
	engine.ServeHTTP(w, req)
	return w
}

func TestAuth_MissingHeader(t *testing.T) {
	w := runAuth("secret", "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_NonBearerHeader(t *testing.T) {
	w := runAuth("secret", "Basic dXNlcjpwYXNz")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_InvalidToken(t *testing.T) {
	w := runAuth("secret", "Bearer notavalidtoken")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_WrongSigningMethod(t *testing.T) {
	// "none" algorithm is not HMAC — our middleware rejects it
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodNone, &Claims{}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	w := runAuth("secret", "Bearer "+tok)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_ValidToken(t *testing.T) {
	uid := uuid.New()
	tok := makeValidToken("mysecret", uid, "user@example.com")
	w := runAuth("mysecret", "Bearer "+tok)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserIDFromCtx_Set(t *testing.T) {
	uid := uuid.New()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", uid)
	assert.Equal(t, uid, UserIDFromCtx(c))
}

func TestUserIDFromCtx_Missing(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	assert.Equal(t, uuid.Nil, UserIDFromCtx(c))
}
