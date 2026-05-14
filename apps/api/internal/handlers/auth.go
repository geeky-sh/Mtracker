package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aash/mtracker/apps/api/internal/middleware"
	"github.com/aash/mtracker/apps/api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db        *gorm.DB
	jwtSecret string
}

func NewAuthHandler(db *gorm.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{db: db, jwtSecret: jwtSecret}
}

type googleUserInfo struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type googleAuthRequest struct {
	AccessToken string `json:"access_token" binding:"required"`
}

// findOrCreateByIdentity looks up a user via user_identities. If no matching
// identity exists, it creates a new User (using newUser) and links it.
// Returns the user and whether it was newly created.
func (h *AuthHandler) findOrCreateByIdentity(provider models.Provider, providerUserID string, newUser func() models.User) (models.User, bool, error) {
	var identity models.UserIdentity
	err := h.db.Where("provider = ? AND provider_user_id = ?", provider, providerUserID).First(&identity).Error
	if err == gorm.ErrRecordNotFound {
		user := newUser()
		if err := h.db.Create(&user).Error; err != nil {
			return models.User{}, false, err
		}
		identity = models.UserIdentity{
			UserID:         user.ID,
			Provider:       provider,
			ProviderUserID: providerUserID,
		}
		if err := h.db.Create(&identity).Error; err != nil {
			return models.User{}, false, err
		}
		return user, true, nil
	} else if err != nil {
		return models.User{}, false, err
	}
	var user models.User
	if err := h.db.First(&user, "id = ?", identity.UserID).Error; err != nil {
		return models.User{}, false, err
	}
	return user, false, nil
}

// GoogleSignIn verifies a Google OAuth access token, upserts the user, and returns a JWT.
func (h *AuthHandler) GoogleSignIn(c *gin.Context) {
	var req googleAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "access_token is required"})
		return
	}

	info, err := fetchGoogleUserInfo(req.AccessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid Google access token"})
		return
	}

	user, created, err := h.findOrCreateByIdentity(models.ProviderGoogle, info.Sub, func() models.User {
		return models.User{Email: info.Email, Name: info.Name, AvatarURL: info.Picture}
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upsert user"})
		return
	}
	if !created {
		h.db.Model(&user).Updates(map[string]interface{}{"name": info.Name, "avatar_url": info.Picture})
	}

	token, err := h.issueJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

// jwtSigner is the function that signs a JWT token; overridable in tests.
var jwtSigner = func(token *jwt.Token, key interface{}) (string, error) {
	return token.SignedString(key)
}

func (h *AuthHandler) issueJWT(user models.User) (string, error) {
	claims := middleware.Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwtSigner(jwt.NewWithClaims(jwt.SigningMethodHS256, claims), []byte(h.jwtSecret))
}

var googleUserInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"

func fetchGoogleUserInfo(accessToken string) (*googleUserInfo, error) {
	resp, err := http.Get(fmt.Sprintf("%s?access_token=%s", googleUserInfoURL, accessToken))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var info googleUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	if info.Sub == "" {
		return nil, fmt.Errorf("invalid user info response")
	}
	return &info, nil
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login authenticates with username/password and returns a JWT.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	if req.Username != "admin" || req.Password != "admin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	user, _, err := h.findOrCreateByIdentity(models.ProviderPassword, "admin", func() models.User {
		return models.User{Email: "admin@mtracker.local", Name: "Admin"}
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	token, err := h.issueJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

// GetProfile returns the authenticated user's profile.
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := middleware.UserIDFromCtx(c)
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}
