package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/yusufatac/bitaksi-case-study/internal/handler"
	"github.com/yusufatac/bitaksi-case-study/internal/middleware"
	"github.com/yusufatac/bitaksi-case-study/internal/repository/mongodb"
	"github.com/yusufatac/bitaksi-case-study/internal/router"
	"github.com/yusufatac/bitaksi-case-study/internal/service"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := getEnv("MONGODB_URI", "mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Ping MongoDB
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	db := client.Database(getEnv("MONGODB_DATABASE", "bitaksi"))

	// Initialize repositories
	locationRepo := mongodb.NewLocationRepository(db)
	userRepo := mongodb.NewUserRepository(db)

	// Initialize services
	locationService := service.NewLocationService(locationRepo)
	matchingService := service.NewMatchingService(locationService)
	authService := service.NewAuthService(userRepo, getEnv("JWT_SECRET", "your-secret-key"))

	// Initialize handlers
	locationHandler := handler.NewLocationHandler(locationService)
	matchingHandler := handler.NewMatchingHandler(matchingService)
	authHandler := handler.NewAuthHandler(authService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(getEnv("JWT_SECRET", "your-secret-key"))
	circuitBreaker := middleware.NewCircuitBreaker(
		middleware.WithFailureThreshold(5),
		middleware.WithResetTimeout(10*time.Second),
	)

	// Initialize router
	r := router.NewRouter(
		authMiddleware,
		locationHandler,
		matchingHandler,
		authHandler,
	)

	// Setup routes with circuit breaker
	r.Engine.Use(circuitBreaker.Middleware())
	r.SetupMatchingApiRoutes()

	// Start server with graceful shutdown
	port := getEnv("PORT", "8081")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r.Engine,
	}

	// Channel to listen for OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	log.Printf("Matching API Service starting on port %s", port)

	<-quit
	log.Println("Shutting down server...")

	// Create a deadline to wait for.
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
