package quality

import (
	"sort"
)

type DefectCount struct {
	DefectCode string  `json:"defect_code"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type DefectAnalysisResult struct {
	TotalProduced   int           `json:"total_produced"`
	TotalRejected   int           `json:"total_rejected"`
	DefectRate      float64       `json:"defect_rate"`
	ParetoBreakdown []DefectCount `json:"pareto_breakdown"`
	AlertTriggered  bool          `json:"alert_triggered"`
}

// AnalyzeDefects computes defect rate, Pareto distribution, and evaluates against alarm thresholds.
func AnalyzeDefects(
	goodUnits int,
	rejectedUnits int,
	defectCounts map[string]int,
	baselineThreshold float64,
) DefectAnalysisResult {
	total := goodUnits + rejectedUnits
	res := DefectAnalysisResult{
		TotalProduced: total,
		TotalRejected: rejectedUnits,
	}

	if total <= 0 {
		return res
	}

	res.DefectRate = float64(rejectedUnits) / float64(total)
	if baselineThreshold > 0 && res.DefectRate > baselineThreshold {
		res.AlertTriggered = true
	}

	// Build Pareto distribution
	var pareto []DefectCount
	for code, count := range defectCounts {
		pct := 0.0
		if rejectedUnits > 0 {
			pct = (float64(count) / float64(rejectedUnits)) * 100.0
		}
		pareto = append(pareto, DefectCount{
			DefectCode: code,
			Count:      count,
			Percentage: pct,
		})
	}

	// Sort Pareto descending by count
	sort.Slice(pareto, func(i, j int) bool {
		return pareto[i].Count > pareto[j].Count
	})

	res.ParetoBreakdown = pareto
	return res
}

