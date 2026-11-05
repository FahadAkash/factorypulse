package oee_test

import (
	"testing"

	"github.com/FahadAkash/factorypulse/backend/internal/oee"
)

func TestComputeMachineOEE_NormalOperation(t *testing.T) {
	losses := oee.MachineLossBreakdown{
		PlannedTimeSeconds: 3600, // 1 hour
		RunTimeSeconds:     3000,
		StoppedSeconds:     300,
		StarvedSeconds:     200,
		BlockedSeconds:     100,
	}

	// 50 units produced in 3000s, ideal cycle = 50s => raw performance = (50*50)/3000 = 2500/3000 = 0.8333
	// Good = 48, Rejected = 2 => Quality = 48/50 = 0.96
	// Availability = 3000 / 3600 = 0.8333
	metrics := oee.ComputeMachineOEE(losses, 50, 48, 50000)

	if metrics.Availability < 0.83 || metrics.Availability > 0.84 {
		t.Errorf("expected availability ~0.833, got %f", metrics.Availability)
	}
	if metrics.Performance < 0.83 || metrics.Performance > 0.84 {
		t.Errorf("expected performance ~0.833, got %f", metrics.Performance)
	}
	if metrics.Quality == nil || *metrics.Quality < 0.959 || *metrics.Quality > 0.961 {
		t.Errorf("expected quality 0.96, got %v", metrics.Quality)
	}
	if metrics.OEE == nil {
		t.Fatalf("expected non-nil OEE")
	}
}

func TestComputeMachineOEE_PerformanceCappedAtOne(t *testing.T) {
	// Burst scenario: 100 units reported in 1000s, ideal cycle = 50s => raw = 5000 / 1000 = 5.0
	// Domain rule: Performance must be strictly capped at 1.0
	losses := oee.MachineLossBreakdown{
		PlannedTimeSeconds: 1000,
		RunTimeSeconds:     1000,
	}

	metrics := oee.ComputeMachineOEE(losses, 100, 100, 50000)
	if metrics.Performance > 1.0 {
		t.Errorf("expected performance capped at 1.0, got %f", metrics.Performance)
	}
	if metrics.Performance != 1.0 {
		t.Errorf("expected performance = 1.0, got %f", metrics.Performance)
	}
}

func TestComputeMachineOEE_DivisionByZeroGuarded(t *testing.T) {
	losses := oee.MachineLossBreakdown{
		PlannedTimeSeconds: 1000,
		RunTimeSeconds:     500,
		StoppedSeconds:     500,
	}

	// 0 units produced => Quality must be nil ("n/a"), never 0 or 100
	metrics := oee.ComputeMachineOEE(losses, 0, 0, 50000)
	if metrics.Quality != nil {
		t.Errorf("expected quality nil when total=0, got %v", *metrics.Quality)
	}
	if metrics.OEE != nil {
		t.Errorf("expected OEE nil when total=0, got %v", *metrics.OEE)
	}
}

