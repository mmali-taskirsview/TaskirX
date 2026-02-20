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
	analyticsHandler := handler.NewAnalyticsHandler(biddingService)
	advancedHandler := handler.NewAdvancedHandler(biddingService)
	log.Printf("✓ Initialized handlers: bidHandler=%v, analyticsHandler=%v, advancedHandler=%v", bidHandler != nil, analyticsHandler != nil, advancedHandler != nil)

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
	router.POST("/openrtb", bidHandler.HandleOpenRTB) // New OpenRTB Endpoint
	router.GET("/health", bidHandler.HandleHealth)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.POST("/refresh", bidHandler.HandleRefresh)
	router.GET("/track", bidHandler.HandleTrack)

	// Analytics routes for Supply Path Optimization
	log.Println("Registering analytics routes...")
	router.GET("/api/analytics/supply-chain", analyticsHandler.GetSupplyChainMetrics)
	router.GET("/api/analytics/supply-path-optimization", analyticsHandler.GetSupplyPathOptimization)
	router.GET("/api/analytics/bid-path/:requestId", analyticsHandler.GetBidPathAnalytics)
	router.GET("/api/analytics/service-performance", analyticsHandler.GetServicePerformance)
	router.GET("/api/analytics/direct-publisher-analysis", analyticsHandler.GetDirectPublisherAnalysis)
	router.GET("/api/analytics/cost-benefit-analysis", analyticsHandler.GetCostBenefitAnalysis)
	router.GET("/api/analytics/bid-landscape", analyticsHandler.GetBidLandscape)
	router.GET("/api/analytics/auto-bid-recommendations", analyticsHandler.GetAutoBidRecommendations)
	router.GET("/api/analytics/segment-performance", analyticsHandler.GetSegmentPerformance)
	router.GET("/api/analytics/optimal-bid-floor", analyticsHandler.GetOptimalBidFloor)
	router.GET("/api/analytics/multi-touch-attribution", analyticsHandler.GetMultiTouchAttribution)

	// Cross-device targeting routes
	router.POST("/api/analytics/cross-device/link", analyticsHandler.LinkDevices)
	router.GET("/api/analytics/cross-device/graph", analyticsHandler.GetDeviceGraph)
	router.GET("/api/analytics/cross-device/frequency", analyticsHandler.GetCrossDeviceFrequency)

	// Tracking pixel routes (used by ImpressionURL / ClickURL in ad markup)
	router.GET("/api/analytics/track/click", analyticsHandler.TrackClick)
	router.GET("/api/analytics/track/impression", analyticsHandler.TrackImpression)

	// Conversion attribution tracking (postback endpoint)
	router.POST("/api/analytics/track/conversion", analyticsHandler.TrackConversion)
	log.Println("✓ Analytics routes registered")

	// Advanced Services Routes
	log.Println("Registering advanced service routes...")

	// Bid Landscape
	router.POST("/api/advanced/bid-landscape/analyze", advancedHandler.HandleBidLandscapeAnalysis)
	router.POST("/api/advanced/bid-landscape/record", advancedHandler.HandleRecordBid)

	// Creative Optimization
	router.POST("/api/advanced/creative/select", advancedHandler.HandleCreativeSelect)

	// Incrementality Testing
	router.POST("/api/advanced/incrementality/evaluate", advancedHandler.HandleIncrementalityEval)
	router.GET("/api/advanced/incrementality/results/:experiment_id", advancedHandler.HandleGetExperimentResults)
	router.POST("/api/advanced/incrementality/conversion", advancedHandler.HandleRecordConversion)

	// Privacy Sandbox
	router.POST("/api/advanced/privacy/topic", advancedHandler.HandleRegisterTopic)
	router.POST("/api/advanced/privacy/interest-group", advancedHandler.HandleAddToInterestGroup)
	router.GET("/api/advanced/privacy/interest-groups/:user_id", advancedHandler.HandleGetInterestGroups)

	// Contextual AI
	router.POST("/api/advanced/contextual/analyze", advancedHandler.HandleContextAnalysis)

	// Real-Time Alerts
	router.POST("/api/advanced/alerts/check", advancedHandler.HandleCheckAlerts)
	router.POST("/api/advanced/alerts/metrics", advancedHandler.HandleRecordMetrics)

	// Competitive Intelligence
	router.POST("/api/advanced/competitive/analyze", advancedHandler.HandleCompetitiveAnalysis)
	router.POST("/api/advanced/competitive/outcome", advancedHandler.HandleRecordAuctionOutcome)
	router.GET("/api/advanced/competitive/report", advancedHandler.HandleGetMarketReport)

	// Unified ID
	router.POST("/api/advanced/identity/resolve", advancedHandler.HandleResolveIdentity)
	router.POST("/api/advanced/identity/link", advancedHandler.HandleLinkIdentities)
	router.GET("/api/advanced/identity/report", advancedHandler.HandleGetIdentityReport)
	router.GET("/api/advanced/identity/cross-device-reach", advancedHandler.HandleGetCrossDeviceReach)

	// Advanced Services Status
	router.GET("/api/advanced/status", advancedHandler.HandleAdvancedServicesStatus)

	// S2S Bidding (Server-to-Server Header Bidding)
	router.POST("/api/advanced/s2s/partner", advancedHandler.HandleRegisterS2SPartner)
	router.GET("/api/advanced/s2s/partner/:id", advancedHandler.HandleGetS2SPartner)
	router.GET("/api/advanced/s2s/partners", advancedHandler.HandleListS2SPartners)
	router.DELETE("/api/advanced/s2s/partner/:id", advancedHandler.HandleRemoveS2SPartner)
	router.POST("/api/advanced/s2s/bid", advancedHandler.HandleS2SBidRequest)
	router.GET("/api/advanced/s2s/stats", advancedHandler.HandleGetS2SStats)

	// Bid Cache Management
	router.GET("/api/advanced/cache/stats", advancedHandler.HandleGetBidCacheStats)
	router.GET("/api/advanced/cache/hit-rate", advancedHandler.HandleGetBidCacheHitRate)
	router.DELETE("/api/advanced/cache", advancedHandler.HandleClearBidCache)
	router.DELETE("/api/advanced/cache/partner/:partner_id", advancedHandler.HandleInvalidateBidCachePartner)
	router.POST("/api/advanced/cache/clean", advancedHandler.HandleCleanExpiredBidCache)

	log.Println("✓ Advanced service routes registered")

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
