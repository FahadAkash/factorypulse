package shift_test

import (
	"testing"
	"time"

	"github.com/FahadAkash/factorypulse/backend/internal/shift"
)

func TestComputeShiftForecast_ShiftStart(t *testing.T) {
	// Less than 15 minutes elapsed => INSUFFICIENT_DATA
	res := shift.ComputeShiftForecast(500, 10, 8*time.Hour, 10*time.Minute, 60)
	if res.Status != shift.StatusInsufficientData {
		t.Errorf("expected INSUFFICIENT_DATA at shift start, got %s", res.Status)
	}
}

func TestComputeShiftForecast_OnTrack(t *testing.T) {
	// 4 hours elapsed of 8 hours (4 hours left)
	// Target = 500, Produced = 260. Current rate = 65 units/hr.
	// Projected = 260 + (65 * 4) = 520 => ON_TRACK
	res := shift.ComputeShiftForecast(500, 260, 8*time.Hour, 4*time.Hour, 65)

	if res.Status != shift.StatusOnTrack {
		t.Errorf("expected ON_TRACK, got %s", res.Status)
	}
	if res.ProjectedTotal != 520 {
		t.Errorf("expected projected total 520, got %d", res.ProjectedTotal)
	}
	if res.UnitsRemaining != 240 {
		t.Errorf("expected 240 remaining, got %d", res.UnitsRemaining)
	}
}

func TestComputeShiftForecast_AtRisk(t *testing.T) {
	// 4 hours left. Produced = 220. Current rate = 60 units/hr.
	// Projected = 220 + (60 * 4) = 460 => 460 is 92% of 500 (between 90% and 100%) => AT_RISK
	res := shift.ComputeShiftForecast(500, 220, 8*time.Hour, 4*time.Hour, 60)

	if res.Status != shift.StatusAtRisk {
		t.Errorf("expected AT_RISK, got %s", res.Status)
	}
	if res.ProjectedTotal != 460 {
		t.Errorf("expected projected total 460, got %d", res.ProjectedTotal)
	}
}

func TestComputeShiftForecast_Behind(t *testing.T) {
	// 4 hours left. Produced = 150. Current rate = 50 units/hr.
	// Projected = 150 + (50 * 4) = 350 => 350 is 70% of 500 (< 90%) => BEHIND
	res := shift.ComputeShiftForecast(500, 150, 8*time.Hour, 4*time.Hour, 50)

	if res.Status != shift.StatusBehind {
		t.Errorf("expected BEHIND, got %s", res.Status)
	}
	if res.ProjectedTotal != 350 {
		t.Errorf("expected projected total 350, got %d", res.ProjectedTotal)
	}
}

