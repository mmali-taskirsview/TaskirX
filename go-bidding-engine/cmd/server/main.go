package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/handler"
	"github.com/taskirx/go-bidding-engine/internal/service"
)

func main() {
	// Load environment variables
	port := getEnv("PORT", "5000")
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	backendAPIURL := getEnv("BACKEND_API_URL", "http://localhost:4000")
	aiServiceURL := getEnv("AI_SERVICE_URL", "http://ad-matching:6002/api")
	fraudServiceURL := getEnv("FRAUD_SERVICE_URL", "http://fraud-detection:6001/api")
	optServiceURL := getEnv("OPTIMIZATION_SERVICE_URL", "http://bid-optimizer:6003/api")
	env := getEnv("ENV", "development")

	// Build Redis URL from components if individual vars are set
	redisURL := fmt.Sprintf("redis://%s:%s", redisHost, redisPort)

	log.Printf("Starting TaskirX Go Bidding Engine...")
	log.Printf("Environment: %s", env)
	log.Printf("Port: %s", port)
	log.Printf("Backend API: %s", backendAPIURL)
	log.Printf("AI Service: %s", aiServiceURL)
	log.Printf("Fraud Service: %s", fraudServiceURL)
	log.Printf("Optimization Service: %s", optServiceURL)

	// Initialize Redis cache
	redisCache, err := cache.NewRedisCache(redisURL, redisPassword, 0)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisCache.Close()

	log.Println("✓ Connected to Redis")

	// Initialize services
	biddingService := service.NewBiddingService(redisCache, backendAPIURL)
	biddingService.SetAIServiceURL(aiServiceURL)
	biddingService.SetFraudServiceURL(fraudServiceURL)
	biddingService.SetOptimizationServiceURL(optServiceURL)
	
	// Load initial campaigns
	log.Println("Loading initial campaigns...")
	if err := biddingService.RefreshCampaigns(backendAPIURL); err != nil {
		log.Printf("Warning: Failed to load campaigns: %v", err)
	}

	// Initialize handlers
	bidHandler := handler.NewBidHandler(biddingService)

	// Setup Gin router
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.Default()

	// Middleware
	router.Use(corsMiddleware())
	router.Use(loggingMiddleware())

	// Routes
	router.POST("/bid", bidHandler.HandleBid)
	router.GET("/health", bidHandler.HandleHealth)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.POST("/refresh", bidHandler.HandleRefresh)

	// Start background campaign refresh (every 5 minutes)
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		
		for range ticker.C {
			log.Println("Refreshing campaigns from backend...")
			if err := biddingService.RefreshCampaigns(backendAPIURL); err != nil {
				log.Printf("Failed to refresh campaigns: %v", err)
			}
		}
	}()

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("🚀 Server listening on %s", addr)
	log.Printf("📊 Metrics available at http://localhost:%s/metrics", port)
	
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

// loggingMiddleware logs requests
func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		
		c.Next()
		
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		
		log.Printf("[%s] %s %d %v", method, path, statusCode, latency)
	}
}
