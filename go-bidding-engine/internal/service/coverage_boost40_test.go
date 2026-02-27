package service

// Boost 40: Targeting remaining edge cases
// Focus: isEventActive full date format, seasonal edge cases, additional helper functions
// Target: isEventActive (92.0%), calculateSeasonalMultiplier (71.9%)

import (
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ── isEventActive - Full Date Format (YYYY-MM-DD) ────────────────────────────

// Test full date format parsing (non-recurring events)
func TestB40_isEventActive_FullDateFormat_ActiveEvent(t *testing.T) {
	mc := &MockCache{}
	svc := NewBiddingService(mc, "test-boost40-1")

	// Event from March 15, 2026 to March 20, 2026
	event := model.SeasonalEvent{
		Name:      "Spring Sale",
		StartDate: "2026-03-15",
		EndDate:   "2026-03-20",
		Boost:     1.5,
		Active:    true,
		Recurring: false,
	}

	// Test date: March 17, 2026 (within range)
	testDate := time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC)

	active := svc.isEventActive(event, testDate)
	if !active {
		t.Errorf("Expected event to be active on March 17, got inactive")
	}
}

func TestB40_isEventActive_FullDateFormat_BeforeStart(t *testing.T) {
	mc := &MockCache{}
	svc := NewBiddingService(mc, "test-boost40-2")

	event := model.SeasonalEvent{
		Name:      "Summer Campaign",
		StartDate: "2026-06-01",
		EndDate:   "2026-06-30",
		Boost:     1.3,
		Active:    true,
		Recurring: false,
	}

	// Test date: May 31, 2026 (before start)
	testDate := time.Date(2026, 5, 31, 23, 59, 59, 0, time.UTC)

	active := svc.isEventActive(event, testDate)
	if active {
		t.Errorf("Expected event to be inactive before start date, got active")
	}
}

func TestB40_isEventActive_FullDateFormat_AfterEnd(t *testing.T) {
	mc := &MockCache{}
	svc := NewBiddingService(mc, "test-boost40-3")

	event := model.SeasonalEvent{
		Name:      "Holiday Special",
		StartDate: "2025-12-20",
		EndDate:   "2025-12-31",
		Boost:     2.0,
		Active:    true,
		Recurring: false,
	}

	// Test date: January 1, 2026 (after end)
	testDate := time.Date(2026, 1, 1, 0, 0, 1, 0, time.UTC)

	active := svc.isEventActive(event, testDate)
	if active {
		t.Errorf("Expected event to be inactive after end date, got active")
	}
}

func TestB40_isEventActive_FullDateFormat_ExactStartDate(t *testing.T) {
	mc := &MockCache{}
	svc := NewBiddingService(mc, "test-boost40-4")

	event := model.SeasonalEvent{
		Name:      "Launch Day",
		StartDate: "2026-04-01",
		EndDate:   "2026-04-15",
		Boost:     1.8,
		Active:    true,
		Recurring: false,
	}

	// Test date: April 1, 2026 00:00:00 (exact start)
	testDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	active := svc.isEventActive(event, testDate)
	if !active {
		t.Errorf("Expected event to be active at exact start time, got inactive")
	}
}

func TestB40_isEventActive_FullDateFormat_ExactEndDate(t *testing.T) {
	mc := &MockCache{}
	svc := NewBiddingService(mc, "test-boost40-5")

	event := model.SeasonalEvent{
		Name:      "End of Season",
		StartDate: "2026-08-15",
		EndDate:   "2026-08-31",
		Boost:     1.4,
		Active:    true,
		Recurring: false,
	}

	// Test date: August 31, 2026 23:59:59 (end of day)
	testDate := time.Date(2026, 8, 31, 23, 59, 59, 0, time.UTC)

	active := svc.isEventActive(event, testDate)
	if !active {
		t.Errorf("Expected event to be active at end of day, got inactive")
	}
}

func TestB40_isEventActive_FullDateFormat_InvalidStartDate(t *testing.T) {
	mc := &MockCache{}
	svc := NewBiddingService(mc, "test-boost40-6")

	event := model.SeasonalEvent{
		Name:      "Invalid Event",
		StartDate: "2026-13-45", // Invalid month and day
		EndDate:   "2026-12-31",
		Boost:     1.5,
		Active:    true,
		Recurring: false,
	}

	testDate := time.Date(2026, 12, 25, 12, 0, 0, 0, time.UTC)

	active := svc.isEventActive(event, testDate)
	if active {
		t.Errorf("Expected invalid start date to return false, got true")
	}
}

func TestB40_isEventActive_FullDateFormat_InvalidEndDate(t *testing.T) {
	mc := &MockCache{}
	svc := NewBiddingService(mc, "test-boost40-7")

	event := model.SeasonalEvent{
		Name:      "Bad End Date",
		StartDate: "2026-01-01",
		EndDate:   "not-a-date",
		Boost:     1.3,
		Active:    true,
		Recurring: false,
	}

	testDate := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	active := svc.isEventActive(event, testDate)
	if active {
		t.Errorf("Expected invalid end date to return false, got true")
	}
}

// ── isEventActive - Recurring vs Non-Recurring Edge Cases ────────────────────

// Test recurring event with year wrap - different scenario
func TestB40_isEventActive_RecurringYearWrap_NotInRange(t *testing.T) {
	mc := &MockCache{}
	svc := NewBiddingService(mc, "test-boost40-8")

	// Recurring event: Dec 26 - Jan 2 (year wrap)
	event := model.SeasonalEvent{
		Name:      "Holiday Sale",
		StartDate: "12-26",
		EndDate:   "01-02",
		Boost:     1.8,
		Active:    true,
		Recurring: true,
	}

	// Test date: February 15 (completely outside range)
	testDate := time.Date(2026, 2, 15, 12, 0, 0, 0, time.UTC)

	active := svc.isEventActive(event, testDate)
	if active {
		t.Errorf("Expected event to be inactive in February, got active")
	}
}

// Test non-recurring full date event crossing months
func TestB40_isEventActive_FullDateCrossingMonths(t *testing.T) {
	mc := &MockCache{}
	svc := NewBiddingService(mc, "test-boost40-9")

	event := model.SeasonalEvent{
		Name:      "Month Crossing Event",
		StartDate: "2026-03-25",
		EndDate:   "2026-04-05",
		Boost:     1.5,
		Active:    true,
		Recurring: false,
	}

	// Test date: April 1, 2026 (in second month of range)
	testDate := time.Date(2026, 4, 1, 15, 0, 0, 0, time.UTC)

	active := svc.isEventActive(event, testDate)
	if !active {
		t.Errorf("Expected event to be active on April 1, got inactive")
	}
}

// Test short date format that could be mistaken for full date
func TestB40_isEventActive_ShortDateFormat_Length5(t *testing.T) {
	mc := &MockCache{}
	svc := NewBiddingService(mc, "test-boost40-10")

	// 5-character date: MM-DD format
	event := model.SeasonalEvent{
		Name:      "Summer Start",
		StartDate: "06-01",
		EndDate:   "06-30",
		Boost:     1.3,
		Active:    true,
		Recurring: true,
	}

	// Test date: June 15 (within range)
	testDate := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)

	active := svc.isEventActive(event, testDate)
	if !active {
		t.Errorf("Expected MM-DD format event to be active in June, got inactive")
	}
}

// Test edge case: event spanning entire year (long duration)
func TestB40_isEventActive_FullDate_LongDuration(t *testing.T) {
	mc := &MockCache{}
	svc := NewBiddingService(mc, "test-boost40-11")

	event := model.SeasonalEvent{
		Name:      "Yearly Campaign",
		StartDate: "2026-01-01",
		EndDate:   "2026-12-31",
		Boost:     1.2,
		Active:    true,
		Recurring: false,
	}

	// Test date: July 4, 2026 (middle of year)
	testDate := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)

	active := svc.isEventActive(event, testDate)
	if !active {
		t.Errorf("Expected long-duration event to be active mid-year, got inactive")
	}
}

// Test one-day event with full date format
func TestB40_isEventActive_FullDate_SingleDay(t *testing.T) {
	mc := &MockCache{}
	svc := NewBiddingService(mc, "test-boost40-12")

	event := model.SeasonalEvent{
		Name:      "Flash Sale",
		StartDate: "2026-05-15",
		EndDate:   "2026-05-15",
		Boost:     2.5,
		Active:    true,
		Recurring: false,
	}

	// Test date: May 15, 2026 (the single day)
	testDate := time.Date(2026, 5, 15, 18, 0, 0, 0, time.UTC)

	active := svc.isEventActive(event, testDate)
	if !active {
		t.Errorf("Expected single-day event to be active on that day, got inactive")
	}
}

// Test one-day event - day after
func TestB40_isEventActive_FullDate_SingleDay_DayAfter(t *testing.T) {
	mc := &MockCache{}
	svc := NewBiddingService(mc, "test-boost40-13")

	event := model.SeasonalEvent{
		Name:      "Flash Sale",
		StartDate: "2026-05-15",
		EndDate:   "2026-05-15",
		Boost:     2.5,
		Active:    true,
		Recurring: false,
	}

	// Test date: May 16, 2026 (day after)
	testDate := time.Date(2026, 5, 16, 0, 0, 1, 0, time.UTC)

	active := svc.isEventActive(event, testDate)
	if active {
		t.Errorf("Expected single-day event to be inactive day after, got active")
	}
}

// Test recurring event - not in short format (but Recurring flag set)
func TestB40_isEventActive_RecurringFlag_WithShortDate(t *testing.T) {
	mc := &MockCache{}
	svc := NewBiddingService(mc, "test-boost40-14")

	event := model.SeasonalEvent{
		Name:      "Weekly Promo",
		StartDate: "07-01",
		EndDate:   "07-07",
		Boost:     1.4,
		Active:    true,
		Recurring: true,
	}

	// Test date: July 5 (within range)
	testDate := time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC)

	active := svc.isEventActive(event, testDate)
	if !active {
		t.Errorf("Expected recurring short-date event to be active, got inactive")
	}
}
