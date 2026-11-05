package bottleneck

import (
	"sort"
)

type StationWindowMetrics struct {
	MachineID         string  `json:"machine_id"`
	StationOrder      int     `json:"station_order"`
	RunTimeSeconds    float64 `json:"run_time_seconds"`
	WindowSeconds     float64 `json:"window_seconds"`
	StarvedRatio      float64 `json:"starved_ratio"` // 0.0 to 1.0
	BlockedRatio      float64 `json:"blocked_ratio"` // 0.0 to 1.0
	UnitsProduced     int     `json:"units_produced"`
	DesignedCapacity  int     `json:"designed_capacity"`
	HasActiveDowntime bool    `json:"has_active_downtime"`
}

type BottleneckCandidate struct {
	MachineID         string  `json:"machine_id"`
	StationOrder      int     `json:"station_order"`
	BottleneckScore   float64 `json:"bottleneck_score"` // Higher is more constrained
	Utilization       float64 `json:"utilization"`
	ThroughputRatio   float64 `json:"throughput_ratio"`
	UpstreamBlocked   bool    `json:"upstream_blocked"`
	DownstreamStarved bool    `json:"downstream_starved"`
	EvidenceReason    string  `json:"evidence_reason"`
}

// EvaluateLineBottlenecks ranks stations on a line based on utilization, buffer propagation, and throughput constraints.
func EvaluateLineBottlenecks(stations []StationWindowMetrics) []BottleneckCandidate {
	var candidates []BottleneckCandidate
	if len(stations) == 0 {
		return candidates
	}

	// Sort stations by station order (1 to N)
	sorted := make([]StationWindowMetrics, len(stations))
	copy(sorted, stations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].StationOrder < sorted[j].StationOrder
	})

	for i, st := range sorted {
		// Domain rule: Exclude active STOPPED downtime intervals (that is tracked as incident, not line constraint)
		if st.HasActiveDowntime {
			continue
		}

		util := 0.0
		if st.WindowSeconds > 0 {
			util = st.RunTimeSeconds / st.WindowSeconds
		}

		tpRatio := 1.0
		if st.DesignedCapacity > 0 {
			tpRatio = float64(st.UnitsProduced) / float64(st.DesignedCapacity)
		}

		// Check if upstream station is frequently BLOCKED
		upstreamBlocked := false
		if i > 0 && sorted[i-1].BlockedRatio > 0.15 {
			upstreamBlocked = true
		}

		// Check if downstream station is frequently STARVED
		downstreamStarved := false
		if i < len(sorted)-1 && sorted[i+1].StarvedRatio > 0.15 {
			downstreamStarved = true
		}

		// Bottleneck score formula: high utilization + low throughput + upstream blocked + downstream starved
		score := util * (1.0 - (tpRatio * 0.5))
		reason := "Normal operations"

		if upstreamBlocked && downstreamStarved {
			score += 1.0 // Strongest constraint indicator
			reason = "High utilization while upstream is BLOCKED and downstream is STARVED"
		} else if upstreamBlocked {
			score += 0.5
			reason = "Backing up upstream stations (upstream BLOCKED)"
		} else if downstreamStarved {
			score += 0.5
			reason = "Starving downstream stations (downstream STARVED)"
		}

		candidates = append(candidates, BottleneckCandidate{
			MachineID:         st.MachineID,
			StationOrder:      st.StationOrder,
			BottleneckScore:   score,
			Utilization:       util,
			ThroughputRatio:   tpRatio,
			UpstreamBlocked:   upstreamBlocked,
			DownstreamStarved: downstreamStarved,
			EvidenceReason:    reason,
		})
	}

	// Rank descending by constraint score
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].BottleneckScore > candidates[j].BottleneckScore
	})

	return candidates
}

type PersistenceTracker struct {
	ConsecutiveWindows map[string]int
	Threshold          int
}

func NewPersistenceTracker(threshold int) *PersistenceTracker {
	if threshold <= 0 {
		threshold = 3
	}
	return &PersistenceTracker{
		ConsecutiveWindows: make(map[string]int),
		Threshold:          threshold,
	}
}

func (t *PersistenceTracker) RecordWindowTop(candidateMachineID string) bool {
	for k := range t.ConsecutiveWindows {
		if k != candidateMachineID {
			t.ConsecutiveWindows[k] = 0
		}
	}

	t.ConsecutiveWindows[candidateMachineID]++
	return t.ConsecutiveWindows[candidateMachineID] >= t.Threshold
}
