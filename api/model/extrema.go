package model

import (
	"errors"
	"fmt"
	"math"
)

const (
	// maxNumBuckets is the maximum number of buckets to use for histograms
	maxNumBuckets = 50
)

// Extrema represents the extrema for a single variable.
type Extrema struct {
	Key  string  `json:"-"`
	Type string  `json:"-"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
}

// NewExtrema instantiates a new extrema struct.
func NewExtrema(min float64, max float64) (*Extrema, error) {
	if min >= max {
		return nil, fmt.Errorf("extrema min cannot be equal to or greater than max")
	}
	if math.IsNaN(min) || math.IsNaN(max) {
		return nil, errors.New("extrema cannot contain NaN values")
	}
	return &Extrema{
		Min: min,
		Max: max,
	}, nil
}

// GetBucketInterval calculates the size of the buckets given the extrema.
func (e *Extrema) GetBucketInterval() float64 {
	if IsFloatingPoint(e.Type) {
		return e.getFloatingPointInterval()
	}
	return e.getIntegerInterval()
}

// GetBucketCount calculates the number of buckets for the extrema.
func (e *Extrema) GetBucketCount() int {
	interval := e.GetBucketInterval()
	rounded := e.GetBucketMinMax()
	rang := rounded.Max - rounded.Min
	return int(round(rang / interval))
}

// GetBucketMinMax calculates the bucket min and max for the extrema.
func (e *Extrema) GetBucketMinMax() *Extrema {
	interval := e.GetBucketInterval()

	roundedMin := floorByUnit(e.Min, interval)
	roundedMax := ceilByUnit(e.Max, interval)

	// if interval does not straddle 0, return itf
	if roundedMin > 0 || roundedMin < 0 {
		return &Extrema{
			Min: roundedMin,
			Max: roundedMax,
		}
	}

	// if the interval boundary falls on 0, return it
	if math.Mod(-roundedMin/interval, 1) == 0 {
		return &Extrema{
			Min: roundedMin,
			Max: roundedMax,
		}
	}

	// NOTE: prevent infinite loop, simply return unrounded extrema. This
	// shouldn't ever actually happen, but we know how that usually turns out...
	if math.IsNaN(interval) ||
		math.IsNaN(roundedMin) ||
		math.IsNaN(roundedMax) ||
		interval <= 0 {
		return e
	}

	// build new min from zero
	newMin := 0.0
	for {
		newMin = newMin - interval
		if newMin <= roundedMin {
			break
		}
	}

	// build new max from zero
	newMax := 0.0
	for {
		newMax = newMax + interval
		if newMax >= roundedMax {
			break
		}
	}

	return &Extrema{
		Min: newMin,
		Max: newMax,
	}
}

func (e *Extrema) getFloatingPointInterval() float64 {
	rang := e.Max - e.Min
	interval := math.Abs(rang / maxNumBuckets)
	return roundInterval(interval)
}

func (e *Extrema) getIntegerInterval() float64 {
	rang := e.Max - e.Min
	if int(rang) < maxNumBuckets {
		return 1
	}
	return math.Ceil(rang / maxNumBuckets)
}

func round(x float64) float64 {
	t := math.Trunc(x)
	if math.Abs(x-t) >= 0.5 {
		return t + math.Copysign(1, x)
	}
	return t
}

func floorByUnit(x float64, unit float64) float64 {
	return math.Floor(x/unit) * unit
}

func ceilByUnit(x float64, unit float64) float64 {
	return math.Ceil(x/unit) * unit
}

func roundInterval(interval float64) float64 {
	round := math.Pow(10, math.Floor(math.Log10(interval)))
	// round interval are considered 1, 2, or 5.
	interval /= round
	if interval <= 2 {
		interval = 2
	} else if interval <= 5 {
		interval = 5
	} else {
		interval = 10
	}
	return interval * round
}
