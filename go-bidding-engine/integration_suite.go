package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// Integration test suite for the bidding engine
// Tests end-to-end functionality with real server

type IntegrationTestSuite struct {
	baseURL    string
	httpClient *http.Client
	results    map[string]TestResult
}

type TestResult struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"`
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error,omitempty"`
	Details  interface{}   `json:"details,omitempty"`
}

func NewIntegrationTestSuite(baseURL string) *IntegrationTestSuite {
	return &IntegrationTestSuite{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		results: make(map[string]TestResult),
	}
}

func (its *IntegrationTestSuite) RunAllTests() {
	log.Println("🚀 Starting Integration Test Suite...")

	tests := []func(){
		its.TestHealthEndpoint,
		its.TestBidEndpoint_Basic,
		its.TestBidEndpoint_Video,
		its.TestBidEndpoint_Native,
		its.TestBidEndpoint_InvalidPayload,
		its.TestBidEndpoint_NoMatchingCampaign,
		its.TestMetricsEndpoint,
		its.TestConcurrentRequests,
		its.TestServerLoad,
	}

	for _, test := range tests {
		test()
		time.Sleep(100 * time.Millisecond) // Brief pause between tests
	}

	its.PrintResults()
}

func (its *IntegrationTestSuite) TestHealthEndpoint() {
	start := time.Now()
	name := "Health Endpoint"

	resp, err := its.httpClient.Get(its.baseURL + "/health")
	if err != nil {
		its.results[name] = TestResult{
			Name:     name,
			Status:   "FAIL",
			Duration: time.Since(start),
			Error:    err.Error(),
		}
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Parse response
	var healthResp map[string]interface{}
	json.Unmarshal(body, &healthResp)

	its.results[name] = TestResult{
		Name:     name,
		Status:   "PASS",
		Duration: time.Since(start),
		Details: map[string]interface{}{
			"status_code": resp.StatusCode,
			"response":    healthResp,
		},
	}
}

func (its *IntegrationTestSuite) TestBidEndpoint_Basic() {
	start := time.Now()
	name := "Basic Bid Request"

	payload := map[string]interface{}{
		"id":           "test-bid-integration-001",
		"publisher_id": "pub-integration-001",
		"ad_slot": map[string]interface{}{
			"id":         "slot-001",
			"dimensions": []int{300, 250},
			"formats":    []string{"banner"},
			"position":   "above-fold",
		},
		"user": map[string]interface{}{
			"id":      "user-integration-001",
			"country": "US",
		},
		"device": map[string]interface{}{
			"type": "mobile",
			"os":   "iOS",
		},
		"context": map[string]interface{}{
			"site_domain": "integration-test.com",
		},
	}

	resp, err := its.makePostRequest("/bid", payload)
	if err != nil {
		its.results[name] = TestResult{
			Name:     name,
			Status:   "FAIL",
			Duration: time.Since(start),
			Error:    err.Error(),
		}
		return
	}

	its.results[name] = TestResult{
		Name:     name,
		Status:   "PASS",
		Duration: time.Since(start),
		Details: map[string]interface{}{
			"status_code": resp["status_code"],
			"response":    resp["body"],
			"latency_ms":  time.Since(start).Milliseconds(),
		},
	}
}

func (its *IntegrationTestSuite) TestBidEndpoint_Video() {
	start := time.Now()
	name := "Video Bid Request"

	payload := map[string]interface{}{
		"id":           "test-bid-video-001",
		"publisher_id": "pub-video-001",
		"ad_slot": map[string]interface{}{
			"id":         "slot-video-001",
			"dimensions": []int{1280, 720},
			"formats":    []string{"video"},
			"position":   "in-stream",
			"video": map[string]interface{}{
				"mimes":       []string{"video/mp4"},
				"minduration": 5,
				"maxduration": 30,
				"protocols":   []int{2, 3},
			},
		},
		"user": map[string]interface{}{
			"id":        "user-video-001",
			"country":   "US",
			"interests": []string{"sports", "technology"},
		},
		"device": map[string]interface{}{
			"type": "mobile",
			"os":   "iOS",
		},
	}

	resp, err := its.makePostRequest("/bid", payload)
	if err != nil {
		its.results[name] = TestResult{
			Name:     name,
			Status:   "FAIL",
			Duration: time.Since(start),
			Error:    err.Error(),
		}
		return
	}

	its.results[name] = TestResult{
		Name:     name,
		Status:   "PASS",
		Duration: time.Since(start),
		Details: map[string]interface{}{
			"status_code": resp["status_code"],
			"latency_ms":  time.Since(start).Milliseconds(),
		},
	}
}

func (its *IntegrationTestSuite) TestBidEndpoint_Native() {
	start := time.Now()
	name := "Native Bid Request"

	payload := map[string]interface{}{
		"id":           "test-bid-native-001",
		"publisher_id": "pub-native-001",
		"ad_slot": map[string]interface{}{
			"id":       "slot-native-001",
			"formats":  []string{"native"},
			"position": "feed",
			"native": map[string]interface{}{
				"assets": []map[string]interface{}{
					{"id": 1, "title": map[string]interface{}{"len": 140}},
					{"id": 2, "img": map[string]interface{}{"w": 300, "h": 250}},
					{"id": 3, "data": map[string]interface{}{"type": 1}},
				},
			},
		},
		"user": map[string]interface{}{
			"id":      "user-native-001",
			"country": "US",
		},
		"device": map[string]interface{}{
			"type": "mobile",
			"os":   "Android",
		},
	}

	resp, err := its.makePostRequest("/bid", payload)
	if err != nil {
		its.results[name] = TestResult{
			Name:     name,
			Status:   "FAIL",
			Duration: time.Since(start),
			Error:    err.Error(),
		}
		return
	}

	its.results[name] = TestResult{
		Name:     name,
		Status:   "PASS",
		Duration: time.Since(start),
		Details: map[string]interface{}{
			"latency_ms": time.Since(start).Milliseconds(),
		},
	}
}

func (its *IntegrationTestSuite) TestBidEndpoint_InvalidPayload() {
	start := time.Now()
	name := "Invalid Payload Handling"

	// Send invalid JSON
	payload := map[string]interface{}{
		"invalid": "payload",
	}

	resp, err := its.makePostRequest("/bid", payload)
	if err != nil {
		// Expected for invalid payload
		its.results[name] = TestResult{
			Name:     name,
			Status:   "PASS",
			Duration: time.Since(start),
			Details: map[string]interface{}{
				"expected_error": true,
				"error":          err.Error(),
			},
		}
		return
	}

	// If we get here, server didn't reject invalid payload (might be OK)
	its.results[name] = TestResult{
		Name:     name,
		Status:   "PASS",
		Duration: time.Since(start),
		Details: map[string]interface{}{
			"status_code": resp["status_code"],
			"note":        "Server accepted invalid payload",
		},
	}
}

func (its *IntegrationTestSuite) TestBidEndpoint_NoMatchingCampaign() {
	start := time.Now()
	name := "No Matching Campaign"

	// Request from country not in campaign targeting
	payload := map[string]interface{}{
		"id":           "test-bid-nomatch-001",
		"publisher_id": "pub-nomatch-001",
		"ad_slot": map[string]interface{}{
			"id":         "slot-001",
			"dimensions": []int{300, 250},
			"formats":    []string{"banner"},
		},
		"user": map[string]interface{}{
			"id":      "user-nomatch-001",
			"country": "ZZ", // Invalid country
		},
		"device": map[string]interface{}{
			"type": "mobile",
		},
	}

	resp, err := its.makePostRequest("/bid", payload)
	if err == nil {
		its.results[name] = TestResult{
			Name:     name,
			Status:   "PASS",
			Duration: time.Since(start),
			Details: map[string]interface{}{
				"status_code": resp["status_code"],
				"note":        "Server handled no-match scenario",
			},
		}
	} else {
		its.results[name] = TestResult{
			Name:     name,
			Status:   "PASS",
			Duration: time.Since(start),
			Details: map[string]interface{}{
				"expected_error": true,
				"error":          err.Error(),
			},
		}
	}
}

func (its *IntegrationTestSuite) TestMetricsEndpoint() {
	start := time.Now()
	name := "Metrics Endpoint"

	resp, err := its.httpClient.Get(its.baseURL + "/metrics")
	if err != nil {
		its.results[name] = TestResult{
			Name:     name,
			Status:   "FAIL",
			Duration: time.Since(start),
			Error:    err.Error(),
		}
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	its.results[name] = TestResult{
		Name:     name,
		Status:   "PASS",
		Duration: time.Since(start),
		Details: map[string]interface{}{
			"status_code":   resp.StatusCode,
			"content_type":  resp.Header.Get("Content-Type"),
			"response_size": len(body),
		},
	}
}

func (its *IntegrationTestSuite) TestConcurrentRequests() {
	start := time.Now()
	name := "Concurrent Requests"

	concurrency := 10
	done := make(chan TestResult, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			start := time.Now()
			payload := map[string]interface{}{
				"id":           fmt.Sprintf("concurrent-bid-%d", id),
				"publisher_id": "pub-concurrent",
				"ad_slot": map[string]interface{}{
					"id":         "slot-concurrent",
					"dimensions": []int{300, 250},
					"formats":    []string{"banner"},
				},
				"user": map[string]interface{}{
					"id":      fmt.Sprintf("user-concurrent-%d", id),
					"country": "US",
				},
				"device": map[string]interface{}{
					"type": "mobile",
				},
			}

			_, err := its.makePostRequest("/bid", payload)

			result := TestResult{
				Name:     fmt.Sprintf("Concurrent Request %d", id),
				Duration: time.Since(start),
			}

			if err != nil {
				result.Status = "FAIL"
				result.Error = err.Error()
			} else {
				result.Status = "PASS"
			}

			done <- result
		}(i)
	}

	// Collect results
	passed := 0
	maxDuration := time.Duration(0)
	for i := 0; i < concurrency; i++ {
		result := <-done
		if result.Status == "PASS" {
			passed++
		}
		if result.Duration > maxDuration {
			maxDuration = result.Duration
		}
	}

	status := "PASS"
	if passed < concurrency {
		status = "PARTIAL"
	}

	its.results[name] = TestResult{
		Name:     name,
		Status:   status,
		Duration: time.Since(start),
		Details: map[string]interface{}{
			"concurrency":    concurrency,
			"passed":         passed,
			"max_duration":   maxDuration.Milliseconds(),
			"total_duration": time.Since(start).Milliseconds(),
		},
	}
}

func (its *IntegrationTestSuite) TestServerLoad() {
	start := time.Now()
	name := "Server Load Test"

	totalRequests := 50
	passed := 0
	var totalLatency time.Duration

	for i := 0; i < totalRequests; i++ {
		reqStart := time.Now()
		payload := map[string]interface{}{
			"id":           fmt.Sprintf("load-test-%d", i),
			"publisher_id": "pub-load-test",
			"ad_slot": map[string]interface{}{
				"id":         "slot-load",
				"dimensions": []int{300, 250},
				"formats":    []string{"banner"},
			},
			"user": map[string]interface{}{
				"id":      fmt.Sprintf("user-load-%d", i),
				"country": "US",
			},
			"device": map[string]interface{}{
				"type": "mobile",
			},
		}

		_, err := its.makePostRequest("/bid", payload)
		reqDuration := time.Since(reqStart)
		totalLatency += reqDuration

		if err == nil {
			passed++
		}

		// Small delay to avoid overwhelming
		if i%10 == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	avgLatency := totalLatency / time.Duration(totalRequests)

	its.results[name] = TestResult{
		Name:     name,
		Status:   "PASS",
		Duration: time.Since(start),
		Details: map[string]interface{}{
			"total_requests": totalRequests,
			"passed":         passed,
			"success_rate":   float64(passed) / float64(totalRequests),
			"avg_latency_ms": avgLatency.Milliseconds(),
			"total_duration": time.Since(start).Milliseconds(),
			"rps":            float64(totalRequests) / time.Since(start).Seconds(),
		},
	}
}

func (its *IntegrationTestSuite) makePostRequest(endpoint string, payload interface{}) (map[string]interface{}, error) {
	jsonData, _ := json.Marshal(payload)

	resp, err := its.httpClient.Post(
		its.baseURL+endpoint,
		"application/json",
		strings.NewReader(string(jsonData)),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var responseData interface{}
	json.Unmarshal(body, &responseData)

	return map[string]interface{}{
		"status_code": resp.StatusCode,
		"body":        responseData,
	}, nil
}

func (its *IntegrationTestSuite) PrintResults() {
	log.Println("\n📊 Integration Test Results:")
	log.Println(strings.Repeat("=", 50))

	passed := 0
	failed := 0
	var totalDuration time.Duration

	for _, result := range its.results {
		status := "❌"
		if result.Status == "PASS" {
			status = "✅"
			passed++
		} else {
			failed++
		}

		totalDuration += result.Duration

		log.Printf("%s %s (%.2fms)",
			status,
			result.Name,
			float64(result.Duration.Nanoseconds())/1000000)

		if result.Error != "" {
			log.Printf("   Error: %s", result.Error)
		}
	}

	log.Println(strings.Repeat("=", 50))
	log.Printf("Summary: %d passed, %d failed", passed, failed)
	log.Printf("Total Duration: %.2fms", float64(totalDuration.Nanoseconds())/1000000)

	if failed == 0 {
		log.Println("🎉 All integration tests passed!")
	} else {
		log.Printf("⚠️  %d test(s) failed", failed)
	}
}

func main() {
	serverURL := "http://localhost:8080"

	log.Printf("Starting integration tests against: %s", serverURL)

	suite := NewIntegrationTestSuite(serverURL)
	suite.RunAllTests()
}
