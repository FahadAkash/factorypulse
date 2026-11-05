package bottleneck_test

import (
	"testing"

	"github.com/FahadAkash/factorypulse/backend/internal/bottleneck"
)

func TestEvaluateLineBottlenecks_PinpointsConstraint(t *testing.T) {
	// Refrigerator line: Station 1 (Press), Station 2 (Foam), Station 3 (Assembly), Station 4 (QC)
	// Station 2 is artificially slow:
	// - Station 1 is BLOCKED (blocked_ratio = 0.40)
	// - Station 2 is running at maximum utilization (utilization = 0.95)
	// - Station 3 is STARVED (starved_ratio = 0.45)
	stations := []bottleneck.StationWindowMetrics{
		{
			MachineID:        "M-REF-01",
			StationOrder:     1,
			RunTimeSeconds:   400,
			WindowSeconds:    900,
			BlockedRatio:     0.40, // Backing up
			UnitsProduced:    15,
			DesignedCapacity: 18,
		},
		{
			MachineID:        "M-REF-02",
			StationOrder:     2,
			RunTimeSeconds:   850,
			WindowSeconds:    900,
			UnitsProduced:    14,
			DesignedCapacity: 18,
		},
		{
			MachineID:        "M-REF-03",
			StationOrder:     3,
			RunTimeSeconds:   420,
			WindowSeconds:    900,
			StarvedRatio:     0.45, // Waiting for foam
			UnitsProduced:    14,
			DesignedCapacity: 18,
		},
		{
			MachineID:        "M-REF-04",
			StationOrder:     4,
			RunTimeSeconds:   450,
			WindowSeconds:    900,
			StarvedRatio:     0.30,
			UnitsProduced:    14,
			DesignedCapacity: 18,
		},
	}

	ranked := bottleneck.EvaluateLineBottlenecks(stations)

	if len(ranked) != 4 {
		t.Fatalf("expected 4 candidates, got %d", len(ranked))
	}

	// M-REF-02 must be identified as top constraint
	top := ranked[0]
	if top.MachineID != "M-REF-02" {
		t.Errorf("expected top bottleneck M-REF-02, got %s (reason: %s)", top.MachineID, top.EvidenceReason)
	}
	if !top.UpstreamBlocked {
		t.Errorf("expected UpstreamBlocked = true for M-REF-02")
	}
	if !top.DownstreamStarved {
		t.Errorf("expected DownstreamStarved = true for M-REF-02")
	}
}

func TestPersistenceTracker_RequiresConsecutiveWindows(t *testing.T) {
	tracker := bottleneck.NewPersistenceTracker(3)

	if tracker.RecordWindowTop("M-REF-02") {
		t.Errorf("window 1 should not trigger persistence alert")
	}
	if tracker.RecordWindowTop("M-REF-02") {
		t.Errorf("window 2 should not trigger persistence alert")
	}
	if !tracker.RecordWindowTop("M-REF-02") {
		t.Errorf("window 3 should trigger persistence alert")
	}

	// Fluctuation resets counter
	if tracker.RecordWindowTop("M-REF-01") {
		t.Errorf("different machine should not trigger")
	}
	if tracker.RecordWindowTop("M-REF-02") {
		t.Errorf("reset machine should not trigger on first window")
	}
}

