//
//   Copyright © 2019 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package model

import (
	"errors"
	"fmt"
	"math"

	"github.com/uncharted-distil/distil-compute/model"
)

const (
	// maxNumBuckets is the maximum number of buckets to use for histograms
	maxNumBuckets = 50

	// HourInterval represents an hour time bucketing interval
	HourInterval = 60 * 60
	// DayInterval represents an day time bucketing interval
	DayInterval = HourInterval * 24
	// WeekInterval represents a week time bucketing interval
	WeekInterval = DayInterval * 7
	// MonthInterval represents a month time bucketing interval
	MonthInterval = WeekInterval * 4
)

// Extrema represents the extrema for a single variable.
type Extrema struct {
	Key  string  `json:"-"`
	Type string  `json:"-"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
}

// BinningArgs represent timeseries binning args.
type BinningArgs struct {
	Rounded  *Extrema
	Count    int
	Interval float64
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

// GetTimeseriesBinningArgs returns the histogram binning args
func (e *Extrema) GetTimeseriesBinningArgs(intervalInSeconds int) BinningArgs {
	if model.IsDateTime(e.Type) {
		return BinningArgs{
			Rounded:  e.GetTimeBucketMinMax(intervalInSeconds),
			Interval: float64(intervalInSeconds),
			Count:    e.GetTimeBucketCount(intervalInSeconds),
		}
	}
	return BinningArgs{
		Rounded:  e.GetBucketMinMax(),
		Interval: e.GetBucketInterval(),
		Count:    e.GetBucketCount(),
	}
}

// GetTimeBucketCount calculates the number of buckets for the extrema.
func (e *Extrema) GetTimeBucketCount(intervalInSeconds int) int {
	interval := float64(intervalInSeconds)
	rounded := e.GetTimeBucketMinMax(intervalInSeconds)
	rang := rounded.Max - rounded.Min

	// rounding issues could lead to negative numbers
	count := int(round(rang / interval))
	if count <= 0 {
		count = 1
	}
	return count
}

// GetTimeBucketMinMax calculates the bucket min and max for the extrema.
func (e *Extrema) GetTimeBucketMinMax(intervalInSeconds int) *Extrema {
	interval := float64(intervalInSeconds)

	roundedMin := floorByUnit(e.Min, interval)
	roundedMax := ceilByUnit(e.Max, interval)

	// if interval does not straddle 0, return it
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

// GetBucketInterval calculates the size of the buckets given the extrema.
func (e *Extrema) GetBucketInterval() float64 {
	if model.IsFloatingPoint(e.Type) {
		return e.getFloatingPointInterval()
	}
	return e.getIntegerInterval()
}

// GetBucketCount calculates the number of buckets for the extrema.
func (e *Extrema) GetBucketCount() int {
	interval := e.GetBucketInterval()
	rounded := e.GetBucketMinMax()
	rang := rounded.Max - rounded.Min

	// rounding issues could lead to negative numbers
	count := int(round(rang / interval))
	if count <= 0 {
		count = 1
	} else if count > maxNumBuckets {
		count = maxNumBuckets
	}
	return count
}

// GetBucketMinMax calculates the bucket min and max for the extrema.
func (e *Extrema) GetBucketMinMax() *Extrema {
	interval := e.GetBucketInterval()

	roundedMin := floorByUnit(e.Min, interval)
	roundedMax := ceilByUnit(e.Max, interval)

	// if interval does not straddle 0, return it
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
