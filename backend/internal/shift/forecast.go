package shift

import (
	"math"
	"time"
)

type ForecastStatus string

const (
	StatusOnTrack          ForecastStatus = "ON_TRACK"
	StatusAtRisk           ForecastStatus = "AT_RISK"
	StatusBehind           ForecastStatus = "BEHIND"
	StatusInsufficientData ForecastStatus = "INSUFFICIENT_DATA"
)

type ShiftForecastResult struct {
	ShiftTarget        int            `json:"shift_target"`
	UnitsProduced      int            `json:"units_produced"`
	UnitsRemaining     int            `json:"units_remaining"`
	HoursLeft          float64        `json:"hours_left"`
	CurrentHourlyRate  float64        `json:"current_hourly_rate"`
	RequiredHourlyRate float64        `json:"required_hourly_rate"`
	ProjectedTotal     int            `json:"projected_total"`
	Status             ForecastStatus `json:"status"`
	AssumptionWindow   string         `json:"assumption_window"`
}

// ComputeShiftForecast projects shift completion based on current pace vs target.
func ComputeShiftForecast(
	targetUnits int,
	goodProducedSoFar int,
	shiftTotalDuration time.Duration,
	shiftElapsedDuration time.Duration,
	rollingHourGoodUnits int,
) ShiftForecastResult {
	res := ShiftForecastResult{
		ShiftTarget:      targetUnits,
		UnitsProduced:    goodProducedSoFar,
		UnitsRemaining:   int(math.Max(0, float64(targetUnits-goodProducedSoFar))),
		AssumptionWindow: "Rolling 30-60 min production rate",
	}

	totalHours := shiftTotalDuration.Hours()
	elapsedHours := shiftElapsedDuration.Hours()
	hoursLeft := math.Max(0, totalHours-elapsedHours)
	res.HoursLeft = hoursLeft

	// If shift just started (less than 15 min elapsed), mark insufficient data
	if shiftElapsedDuration < 15*time.Minute {
		res.Status = StatusInsufficientData
		res.ProjectedTotal = goodProducedSoFar
		return res
	}

	if hoursLeft <= 0 {
		res.ProjectedTotal = goodProducedSoFar
		if goodProducedSoFar >= targetUnits {
			res.Status = StatusOnTrack
		} else {
			res.Status = StatusBehind
		}
		return res
	}

	// Required hourly rate to hit remaining target
	res.RequiredHourlyRate = float64(res.UnitsRemaining) / hoursLeft

	// Current rolling hourly rate
	currentRate := float64(rollingHourGoodUnits)
	res.CurrentHourlyRate = currentRate

	// Projected shift total = produced + (current_rate * hours_left)
	projected := float64(goodProducedSoFar) + (currentRate * hoursLeft)
	res.ProjectedTotal = int(math.Round(projected))

	// Status determination:
	// ON_TRACK: projected >= target
	// AT_RISK: projected >= 90% of target and < target
	// BEHIND: projected < 90% of target
	targetFloat := float64(targetUnits)
	if projected >= targetFloat {
		res.Status = StatusOnTrack
	} else if projected >= (targetFloat * 0.90) {
		res.Status = StatusAtRisk
	} else {
		res.Status = StatusBehind
	}

	return res
}

