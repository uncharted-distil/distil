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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uncharted-distil/distil-compute/model"
)

const (
	epsilon = 0.000001
	zero    = 0.0
)

var (
	samples = []*Extrema{
		{
			Type: model.RealType,
			Min:  0.2667,
			Max:  1.4630,
		},
		{
			Type: model.IntegerType,
			Min:  1,
			Max:  20,
		},
		{
			Type: model.IntegerType,
			Min:  1,
			Max:  60,
		},
		{
			Type: model.RealType,
			Min:  -1.4630,
			Max:  -0.2667,
		},
		{
			Type: model.IntegerType,
			Min:  -20,
			Max:  -1,
		},
		{
			Type: model.IntegerType,
			Min:  -60,
			Max:  -1,
		},
		{
			Type: model.RealType,
			Min:  -1.4630,
			Max:  3.2667,
		},
		{
			Type: model.IntegerType,
			Min:  -21,
			Max:  57,
		},
		{
			Type: model.IntegerType,
			Min:  -3,
			Max:  5,
		},
		{
			Type: model.RealType,
			Min:  -512.4630,
			Max:  1097.2667,
		},
	}
)

func TestExtremaBucketInterval(t *testing.T) {
	assert.InDelta(t, samples[0].GetBucketInterval(), 0.05, epsilon)
	assert.InDelta(t, samples[1].GetBucketInterval(), 1.0, zero)
	assert.InDelta(t, samples[2].GetBucketInterval(), 2.0, zero)
	assert.InDelta(t, samples[3].GetBucketInterval(), 0.05, epsilon)
	assert.InDelta(t, samples[4].GetBucketInterval(), 1.0, zero)
	assert.InDelta(t, samples[5].GetBucketInterval(), 2.0, zero)
	assert.InDelta(t, samples[6].GetBucketInterval(), 0.1, zero)
	assert.InDelta(t, samples[7].GetBucketInterval(), 2.0, zero)
	assert.InDelta(t, samples[8].GetBucketInterval(), 1.0, zero)
	assert.InDelta(t, samples[9].GetBucketInterval(), 50.0, zero)
}

func assertRangeIsDivisibleByInterval(t *testing.T, extrema *Extrema) {
	minMax := extrema.GetBucketMinMax()
	interval := extrema.GetBucketInterval()
	div := (minMax.Max - minMax.Min) / interval
	assert.InDelta(t, math.Mod(div, 1.0), 0.0, epsilon)
}

func assertZeroIsBoundary(t *testing.T, extrema *Extrema) {
	if extrema.Min > 0 || extrema.Max < 0 {
		return
	}
	minMax := extrema.GetBucketMinMax()
	interval := extrema.GetBucketInterval()
	div := -minMax.Min / interval
	assert.InDelta(t, math.Mod(div, 1.0), 0.0, epsilon)
}

func TestExtremaBucketMinMax(t *testing.T) {

	for _, sample := range samples {
		assertRangeIsDivisibleByInterval(t, sample)
	}

	for _, sample := range samples {
		assertZeroIsBoundary(t, sample)
	}

	assert.InDelta(t, samples[0].GetBucketMinMax().Min, 0.25, epsilon)
	assert.InDelta(t, samples[0].GetBucketMinMax().Max, 1.5, epsilon)

	assert.InDelta(t, samples[1].GetBucketMinMax().Min, 1.0, epsilon)
	assert.InDelta(t, samples[1].GetBucketMinMax().Max, 20.0, epsilon)

	assert.InDelta(t, samples[2].GetBucketMinMax().Min, 0.0, epsilon)
	assert.InDelta(t, samples[2].GetBucketMinMax().Max, 60.0, epsilon)

	assert.InDelta(t, samples[3].GetBucketMinMax().Min, -1.5, epsilon)
	assert.InDelta(t, samples[3].GetBucketMinMax().Max, -0.25, epsilon)

	assert.InDelta(t, samples[4].GetBucketMinMax().Min, -20.0, epsilon)
	assert.InDelta(t, samples[4].GetBucketMinMax().Max, -1.0, epsilon)

	assert.InDelta(t, samples[5].GetBucketMinMax().Min, -60.0, epsilon)
	assert.InDelta(t, samples[5].GetBucketMinMax().Max, 0.0, epsilon)

	assert.InDelta(t, samples[6].GetBucketMinMax().Min, -1.5, epsilon)
	assert.InDelta(t, samples[6].GetBucketMinMax().Max, 3.3, epsilon)

	assert.InDelta(t, samples[7].GetBucketMinMax().Min, -22.0, epsilon)
	assert.InDelta(t, samples[7].GetBucketMinMax().Max, 58.0, epsilon)

	assert.InDelta(t, samples[8].GetBucketMinMax().Min, -3.0, epsilon)
	assert.InDelta(t, samples[8].GetBucketMinMax().Max, 5.0, epsilon)

	assert.InDelta(t, samples[9].GetBucketMinMax().Min, -550.0, epsilon)
	assert.InDelta(t, samples[9].GetBucketMinMax().Max, 1100.0, epsilon)
}

func TestExtremaBucketCount(t *testing.T) {
	assert.InDelta(t, samples[0].GetBucketCount(), 25, epsilon)
	assert.InDelta(t, samples[1].GetBucketCount(), 19, zero)
	assert.InDelta(t, samples[2].GetBucketCount(), 30, zero)
	assert.InDelta(t, samples[3].GetBucketCount(), 25, epsilon)
	assert.InDelta(t, samples[4].GetBucketCount(), 19, zero)
	assert.InDelta(t, samples[5].GetBucketCount(), 30, zero)
	assert.InDelta(t, samples[6].GetBucketCount(), 48, zero)
	assert.InDelta(t, samples[7].GetBucketCount(), 40, zero)
	assert.InDelta(t, samples[8].GetBucketCount(), 8, zero)
	assert.InDelta(t, samples[9].GetBucketCount(), 33, zero)
}
