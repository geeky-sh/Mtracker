package router

import (
	"github.com/aash/mtracker/apps/api/internal/config"
	"github.com/aash/mtracker/apps/api/internal/handlers"
	"github.com/aash/mtracker/apps/api/internal/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.RequestLogger())

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	authH := handlers.NewAuthHandler(db, cfg.JWTSecret)
	activitiesH := handlers.NewActivitiesHandler(db)
	logsH := handlers.NewLogsHandler(db)
	analyticsH := handlers.NewAnalyticsHandler(db)

	v1 := r.Group("/api/v1")
	{
		// Public
		v1.POST("/auth/login", authH.Login)

		// Protected
		protected := v1.Group("/")
		protected.Use(middleware.Auth(cfg.JWTSecret))
		{
			protected.GET("/profile", authH.GetProfile)

			protected.GET("/activities", activitiesH.List)
			protected.POST("/activities", activitiesH.Create)
			protected.GET("/activities/search", activitiesH.Search)
			protected.DELETE("/activities/:id", activitiesH.Delete)

			protected.GET("/logs", logsH.ListByActivity)
			protected.POST("/logs", logsH.Create)
			protected.DELETE("/logs/:id", logsH.Delete)

			protected.GET("/analytics", analyticsH.Summary)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
