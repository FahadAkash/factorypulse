package quality_test

import (
	"testing"

	"github.com/FahadAkash/factorypulse/backend/internal/quality"
)

func TestAnalyzeDefects_ParetoAndAlert(t *testing.T) {
	counts := map[string]int{
		"SURFACE_SCRATCH":       15,
		"DIMENSION_OUT_OF_SPEC": 30,
		"FOAM_VOIDS":            5,
	}

	// 50 rejected, 950 good => total 1000 => defect rate = 50/1000 = 0.05 (5%)
	// Threshold = 0.03 (3%) => Alert should trigger
	res := quality.AnalyzeDefects(950, 50, counts, 0.03)

	if !res.AlertTriggered {
		t.Errorf("expected alert to trigger for 5%% defect rate vs 3%% baseline")
	}
	if len(res.ParetoBreakdown) != 3 {
		t.Fatalf("expected 3 defect categories, got %d", len(res.ParetoBreakdown))
	}

	// First Pareto item must be DIMENSION_OUT_OF_SPEC with count 30 (60%)
	if res.ParetoBreakdown[0].DefectCode != "DIMENSION_OUT_OF_SPEC" {
		t.Errorf("expected top defect DIMENSION_OUT_OF_SPEC, got %s", res.ParetoBreakdown[0].DefectCode)
	}
	if res.ParetoBreakdown[0].Percentage != 60.0 {
		t.Errorf("expected 60%%, got %f", res.ParetoBreakdown[0].Percentage)
	}
}
