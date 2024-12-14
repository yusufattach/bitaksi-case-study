package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/yusufatac/bitaksi-case-study/internal/handler"
	"github.com/yusufatac/bitaksi-case-study/internal/middleware"
)

type Router struct {
	Engine         *gin.Engine
	authMiddleware *middleware.AuthMiddleware

	locationHandler *handler.LocationHandler
	matchingHandler *handler.MatchingHandler
	authHandler     *handler.AuthHandler
}

func NewRouter(
	authMiddleware *middleware.AuthMiddleware,
	locationHandler *handler.LocationHandler,
	matchingHandler *handler.MatchingHandler,
	authHandler *handler.AuthHandler,
) *Router {
	return &Router{
		Engine:          gin.Default(),
		authMiddleware:  authMiddleware,
		locationHandler: locationHandler,
		matchingHandler: matchingHandler,
		authHandler:     authHandler,
	}
}

func (r *Router) SetupDriverLocationRoutes() {
	// Enable CORS
	r.Engine.Use(corsMiddleware())

	// Swagger documentation
	r.Engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	r.Engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := r.Engine.Group("/api/v1")

	// Public routes
	auth := v1.Group("/auth")
	{
		auth.POST("/register", r.authHandler.Register)
		auth.POST("/login", r.authHandler.Login)
	}

	// Protected routes
	protected := v1.Group("")
	protected.Use(r.authMiddleware.RequireAuth())
	{
		// Location routes
		locations := protected.Group("/locations")
		{
			locations.POST("", r.locationHandler.UpdateLocation)
			locations.POST("/batch", r.locationHandler.UpdateLocations)
			locations.POST("/nearby", r.locationHandler.FindNearbyDrivers)
		}
	}
}

func (r *Router) SetupMatchingApiRoutes() {
	// Enable CORS
	r.Engine.Use(corsMiddleware())

	// Swagger documentation
	r.Engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	r.Engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := r.Engine.Group("/api/v1")

	// Public routes
	auth := v1.Group("/auth")
	{
		auth.POST("/register", r.authHandler.Register)
		auth.POST("/login", r.authHandler.Login)
	}

	// Protected routes
	protected := v1.Group("")
	protected.Use(r.authMiddleware.RequireAuth())
	{
		// Matching routes
		match := protected.Group("/match")
		{
			match.POST("", r.matchingHandler.FindNearestDriver)
		}
	}
}

func (r *Router) Run(addr string) error {
	return r.Engine.Run(addr)
}

// CORS middleware
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
