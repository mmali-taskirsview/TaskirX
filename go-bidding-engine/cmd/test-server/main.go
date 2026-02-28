package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/handler"
	"github.com/taskirx/go-bidding-engine/internal/model"
	"github.com/taskirx/go-bidding-engine/internal/service"
)

func main() {
	port := getEnv("PORT", "8080")
	backendAPIURL := getEnv("BACKEND_API_URL", "http://localhost:4000")
	env := getEnv("ENV", "development")

	log.Printf("Starting TaskirX Go Bidding Engine (Test Mode)...")
	log.Printf("Environment: %s", env)
	log.Printf("Port: %s", port)
	log.Printf("Cache: Mock (in-memory)")

	// Initialize Mock cache (no Redis required)
	mockCache := cache.NewMockCache()
	log.Println("✓ Using Mock Cache (in-memory)")

	// Initialize services
	biddingService := service.NewBiddingService(mockCache, backendAPIURL)

	// Load test campaigns into mock cache
	log.Println("Loading test campaigns into mock cache...")
	loadTestCampaigns(mockCache)

	// Initialize handlers
	bidHandler := handler.NewBidHandler(biddingService)

	// Setup Gin router
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Middleware - Add panic recovery first
	router.Use(panicRecoveryMiddleware())
	router.Use(corsMiddleware())
	router.Use(loggingMiddleware())

	// Routes
	router.POST("/bid", bidHandler.HandleBid)
	router.POST("/openrtb", bidHandler.HandleOpenRTB)
	router.GET("/health", bidHandler.HandleHealth)
	router.GET("/campaigns/refresh", bidHandler.HandleRefresh)

	// Metrics
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Start server
	log.Printf("✓ Server ready on port %s", port)
	log.Printf("✓ Health endpoint: http://localhost:%s/health", port)
	log.Printf("✓ Bid endpoint: http://localhost:%s/bid", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func loadTestCampaigns(cache cache.Cache) {
	// Create test campaigns with correct structure
	campaigns := []model.Campaign{
		{
			ID:       "test-campaign-001",
			Name:     "Test Mobile Campaign",
			Status:   "active",
			Type:     "cpm",
			BidPrice: 2.50,
			Budget:   10000.0,
			Spent:    1000.0,
			Targeting: model.Targeting{
				Countries: []string{"US", "CA"},
				Devices:   []string{"mobile", "tablet"},
			},
			Creative: model.Creative{
				Type:   "banner",
				URL:    "https://example.com/ad.jpg",
				Width:  300,
				Height: 250,
			},
		},
		{
			ID:       "test-campaign-002",
			Name:     "Test Video Campaign",
			Status:   "active",
			Type:     "cpm",
			BidPrice: 5.00,
			Budget:   20000.0,
			Spent:    5000.0,
			Targeting: model.Targeting{
				Countries: []string{"US"},
				Devices:   []string{"mobile", "desktop"},
			},
			Creative: model.Creative{
				Type:     "video",
				URL:      "https://example.com/video.mp4",
				Width:    1280,
				Height:   720,
				Duration: 30,
				MimeType: "video/mp4",
			},
		},
		{
			ID:       "test-campaign-003",
			Name:     "Test Desktop Campaign",
			Status:   "active",
			Type:     "cpc",
			BidPrice: 0.50,
			Budget:   15000.0,
			Spent:    2000.0,
			Targeting: model.Targeting{
				Countries: []string{"US", "UK", "CA"},
				Devices:   []string{"desktop"},
			},
			Creative: model.Creative{
				Type:   "banner",
				URL:    "https://example.com/banner.jpg",
				Width:  728,
				Height: 90,
			},
		},
	}

	// Store campaigns in cache using proper method
	campaignJSON, err := json.Marshal(campaigns)
	if err != nil {
		log.Printf("Warning: Failed to marshal campaigns: %v", err)
		return
	}

	// Store as active campaigns
	cache.Set("active_campaigns", string(campaignJSON), 0)

	// Also store individually by ID
	for _, c := range campaigns {
		cJSON, _ := json.Marshal(c)
		cache.Set(fmt.Sprintf("campaign:%s", c.ID), string(cJSON), 0)
	}

	log.Printf("✓ Test campaigns loaded (%d active campaigns)", len(campaigns))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for health checks and metrics
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		log.Printf("[%s] %s %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())
		c.Next()
	}
}

func panicRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC recovered: %v", err)
				c.JSON(500, gin.H{
					"error":   "Internal server error",
					"message": fmt.Sprintf("%v", err),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
