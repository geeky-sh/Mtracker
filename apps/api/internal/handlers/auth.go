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

	var user models.User
	result := h.db.Where("google_id = ?", info.Sub).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		user = models.User{
			Email:     info.Email,
			GoogleID:  info.Sub,
			Name:      info.Name,
			AvatarURL: info.Picture,
		}
		if err := h.db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
	} else if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	} else {
		h.db.Model(&user).Updates(map[string]interface{}{
			"name":       info.Name,
			"avatar_url": info.Picture,
		})
	}

	token, err := h.issueJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
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

	var user models.User
	result := h.db.Where("email = ?", "admin@mtracker.local").First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		user = models.User{
			Email: "admin@mtracker.local",
			Name:  "Admin",
		}
		if err := h.db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
	} else if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
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
