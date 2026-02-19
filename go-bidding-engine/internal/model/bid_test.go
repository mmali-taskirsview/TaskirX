package model

import (
	"testing"
)

func TestIsMatch_GeoFencing(t *testing.T) {
	// Campaign with GeoFence: NYC (approx 40.7128, -74.0060) within 10km
	campaign := &Campaign{
		Creative: Creative{Type: "banner"},
		Targeting: Targeting{
			GeoFences: []GeoFence{
				{Lat: 40.7128, Lon: -74.0060, Radius: 10.0},
			},
		},
	}

	tests := []struct {
		name     string
		device   InternalDevice
		expected bool
	}{
		{
			name: "Match inside GeoFence",
			device: InternalDevice{
				Type: "mobile",
				Geo: InternalGeo{
					Lat: 40.7200, // Very close
					Lon: -74.0100,
				},
			},
			expected: true,
		},
		{
			name: "No Match outside GeoFence (London)",
			device: InternalDevice{
				Type: "mobile",
				Geo: InternalGeo{
					Lat: 51.5074,
					Lon: -0.1278,
				},
			},
			expected: false,
		},
		{
			name: "No Match border case (20km away)",
			device: InternalDevice{
				Type: "mobile",
				Geo: InternalGeo{
					Lat: 40.7128 + 0.2, // ~22km diff in lat approx
					Lon: -74.0060,
				},
			},
			expected: false,
		},
		{
			name: "No Geo in Request",
			device: InternalDevice{
				Type: "mobile",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &BidRequest{
				Device: tt.device,
				AdSlot: AdSlot{Formats: []string{"banner"}},
			}
			if got := campaign.IsMatch(req); got != tt.expected {
				t.Errorf("IsMatch() = %v, want %v", got, tt.expected)
			}
		})
	}
}
