package oee

import (
	"math"
	"time"
)

type MachineLossBreakdown struct {
	PlannedTimeSeconds float64 `json:"planned_time_seconds"`
	RunTimeSeconds     float64 `json:"run_time_seconds"`
	StoppedSeconds     float64 `json:"stopped_seconds"`
	StarvedSeconds     float64 `json:"starved_seconds"`
	BlockedSeconds     float64 `json:"blocked_seconds"`
	MaintenanceSeconds float64 `json:"maintenance_seconds"`
}

type MetricsResult struct {
	Availability float64              `json:"availability"` // 0.0 to 1.0
	Performance  float64              `json:"performance"`  // 0.0 to 1.0 (capped)
	Quality      *float64             `json:"quality"`      // nil if total = 0 ("n/a")
	OEE          *float64             `json:"oee"`          // nil if quality is nil
	Losses       MachineLossBreakdown `json:"losses"`
}

// ComputeMachineOEE calculates Availability, Performance, and Quality according to domain rules.
func ComputeMachineOEE(
	losses MachineLossBreakdown,
	totalUnits int,
	goodUnits int,
	idealCycleMs int,
) MetricsResult {
	res := MetricsResult{
		Losses: losses,
	}

	// 1. Availability = run_time / planned_time
	// Planned time excludes planned maintenance
	netPlanned := losses.PlannedTimeSeconds
	if netPlanned <= 0 {
		netPlanned = losses.RunTimeSeconds + losses.StoppedSeconds + losses.StarvedSeconds + losses.BlockedSeconds
	}

	if netPlanned > 0 {
		res.Availability = math.Min(1.0, math.Max(0.0, losses.RunTimeSeconds/netPlanned))
	} else {
		res.Availability = 0.0
	}

	// 2. Performance = (total_units * ideal_cycle_time) / run_time, CAPPED at 1.0
	idealCycleSeconds := float64(idealCycleMs) / 1000.0
	if losses.RunTimeSeconds > 0 && totalUnits > 0 {
		rawPerformance := (float64(totalUnits) * idealCycleSeconds) / losses.RunTimeSeconds
		res.Performance = math.Min(1.0, math.Max(0.0, rawPerformance))
	} else {
		res.Performance = 0.0
	}

	// 3. Quality = good / total. Guard division by zero: nil => "n/a"
	if totalUnits > 0 {
		q := math.Min(1.0, math.Max(0.0, float64(goodUnits)/float64(totalUnits)))
		res.Quality = &q

		// Overall OEE = A * P * Q
		oeeVal := res.Availability * res.Performance * q
		res.OEE = &oeeVal
	} else {
		res.Quality = nil
		res.OEE = nil
	}

	return res
}

// ComputeLineOEE computes line-level OEE strictly from final output vs planned line capacity.
// Invariant: Never compute Line OEE as an arithmetic mean of individual machine OEEs.
func ComputeLineOEE(
	finalGoodUnits int,
	plannedCapacityUnits int,
	linePlannedDuration time.Duration,
	lineRunDuration time.Duration,
) MetricsResult {
	res := MetricsResult{}

	// Availability for the line
	if linePlannedDuration > 0 {
		res.Availability = math.Min(1.0, float64(lineRunDuration)/float64(linePlannedDuration))
	}

	// Quality and output achievement
	if plannedCapacityUnits > 0 {
		p := math.Min(1.0, float64(finalGoodUnits)/float64(plannedCapacityUnits))
		res.Performance = p
		q := 1.0 // Finished goods entering pallet have passed final QC
		res.Quality = &q
		oeeVal := res.Availability * p * q
		res.OEE = &oeeVal
	}

	return res
}

