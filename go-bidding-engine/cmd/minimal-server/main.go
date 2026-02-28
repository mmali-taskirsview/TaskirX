package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Starting minimal test server...")

	// Create minimal gin router
	gin.SetMode(gin.ReleaseMode) // Reduce output noise
	router := gin.New()

	// Add basic recovery
	router.Use(gin.Recovery())

	// Simple health endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	// Simple test endpoint
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "test successful",
		})
	})

	log.Println("✓ Minimal server ready on port 8081")
	log.Println("✓ Health: http://localhost:8081/health")
	log.Println("✓ Test: http://localhost:8081/test")

	if err := router.Run(":8081"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
