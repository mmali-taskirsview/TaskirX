package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "ok",
		"service":   "stdlib-test",
		"timestamp": time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "stdlib server working",
		"method":  r.Method,
		"path":    r.URL.Path,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	log.Println("Starting pure stdlib HTTP server...")

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/test", testHandler)

	log.Println("✓ Stdlib server ready on port 8082")
	log.Println("✓ Health: http://localhost:8082/health")
	log.Println("✓ Test: http://localhost:8082/test")

	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
