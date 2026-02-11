package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Counters
	BidRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "bid_requests_total",
		Help: "The total number of bid requests received",
	})

	BidsPlacedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "bids_placed_total",
		Help: "The total number of bids successfully placed",
	})

	FraudBlockedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "fraud_blocked_total",
		Help: "Total number of requests blocked by fraud detection",
	})

	AdMatchErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ad_match_errors_total",
		Help: "Total number of errors interacting with Ad Matching Service",
	})

	OptimizationErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "optimization_errors_total",
		Help: "Total number of errors interacting with Optimization Service",
	})

	// Histograms
	BidLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "bid_request_duration_seconds",
		Help:    "Duration of bid request processing",
		Buckets: prometheus.DefBuckets, // Default buckets: .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
	})
)
