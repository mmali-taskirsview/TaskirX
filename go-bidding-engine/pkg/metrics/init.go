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

	BidRequestsByFormat = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "bid_requests_by_format_total",
		Help: "The total number of bid requests received by format type",
	}, []string{"format"})

	BidsPlacedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "bids_placed_total",
		Help: "The total number of bids successfully placed by format",
	}, []string{"format"})

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

	NoBidTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "no_bid_total",
		Help: "Total number of no-bid responses by reason",
	}, []string{"reason"})

	// Tracking Metrics
	EventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tracking_events_total",
		Help: "Total tracking events by type",
	}, []string{"type", "campaign_id"})

	VideoEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "video_events_total",
		Help: "Video playback events",
	}, []string{"event", "campaign_id"})

	RichMediaEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rich_media_events_total",
		Help: "Rich Media Interaction events",
	}, []string{"action", "campaign_id"})

	// Histograms
	BidLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "bid_request_duration_seconds",
		Help:    "Duration of bid request processing",
		Buckets: prometheus.DefBuckets, // Default buckets: .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
	})
)
