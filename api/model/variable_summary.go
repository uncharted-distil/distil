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

import "github.com/pkg/errors"

const (
	// MinAggPrefix is the prefix used for min aggregations.
	MinAggPrefix = "min_"
	// MaxAggPrefix is the prefix used for max aggregations.
	MaxAggPrefix = "max_"
	// TermsAggPrefix is the prefix used for terms aggregations.
	TermsAggPrefix = "terms_"
	// HistogramAggPrefix is the prefix used for histogram aggregations.
	HistogramAggPrefix = "histogram_"
	// VariableValueField is the field which stores the variable value.
	VariableValueField = "value"
	// VariableTypeField is the field which stores the variable's schema type value.
	VariableTypeField = "schemaType"
)

// Bucket represents a single histogram bucket.
type Bucket struct {
	Key     string    `json:"key"`
	Count   int64     `json:"count"`
	Buckets []*Bucket `json:"buckets,omitempty"`
}

// Histogram represents a single variable histogram.
type Histogram struct {
	Extrema         *Extrema             `json:"extrema,omitempty"`
	Buckets         []*Bucket            `json:"buckets"`
	CategoryBuckets map[string][]*Bucket `json:"categoryBuckets"`
	Exemplars       []string             `json:"exemplars"`
	StdDev          float64              `json:"stddev"`
	Mean            float64              `json:"mean"`
}

// VariableSummary represents a summary of the variable values.
type VariableSummary struct {
	Label            string     `json:"label"`
	Key              string     `json:"key"`
	Description      string     `json:"description"`
	Type             string     `json:"type"`
	VarType          string     `json:"varType"`
	Dataset          string     `json:"dataset"`
	Baseline         *Histogram `json:"baseline"`
	Filtered         *Histogram `json:"filtered"`
	Timeline         *Histogram `json:"timeline"`
	TimelineBaseline *Histogram `json:"timelineBaseline"`
	TimelineType     string     `json:"timelineType"`
	Weighted         bool       `json:"weighted"`
}

// SummaryMode defines the summary display modes.
type SummaryMode int

const (
	// DefaultMode use the default facet for a variable summary given its type, ex. a horizontal histogram for numeric values.
	DefaultMode = iota + 1
	// ClusterMode use computed cluster information for a variable summary if availble, ex. timeseries clusters
	ClusterMode
	// TimeseriesMode use the timeseries grouping to return timeseries counts rather than observation counts.
	TimeseriesMode
	// MultiBandImageMode use the multi-band image grouping to return tile counts rather than image counts.
	MultiBandImageMode
)

// SummaryModeFromString creates a SummaryMode from the supplied string
func SummaryModeFromString(s string) (SummaryMode, error) {
	switch s {
	case "cluster":
		return ClusterMode, nil
	case "timeseries":
		return TimeseriesMode, nil
	case "multiband_image":
		return MultiBandImageMode, nil
	case "default":
		return DefaultMode, nil
	default:
		return 0, errors.Errorf("%s is not a valid SummaryMode", s)
	}
}

// EmptyFilteredHistogram fills the filtered portion of the summary with empty
// bucket counts
func (s *VariableSummary) EmptyFilteredHistogram() {

	if s.Baseline.Buckets != nil {
		s.Filtered = &Histogram{
			Extrema: s.Baseline.Extrema,
		}
		for _, bucket := range s.Baseline.Buckets {
			s.Filtered.Buckets = append(s.Filtered.Buckets, &Bucket{
				Key:   bucket.Key,
				Count: 0,
			})
		}
	}

	if s.Baseline.CategoryBuckets != nil {
		s.Filtered = &Histogram{
			Extrema: s.Baseline.Extrema,
		}
		for category, buckets := range s.Baseline.CategoryBuckets {
			var filtered []*Bucket
			for _, bucket := range buckets {
				filtered = append(filtered, &Bucket{
					Key:   bucket.Key,
					Count: 0,
				})
			}
			s.Filtered.CategoryBuckets[category] = filtered
		}
	}

}

// IsEmpty returns true if no data exists in the histogram (ie sum(buckets) == 0)
func (h *Histogram) IsEmpty() bool {
	return h.bucketsAreEmpty(h.Buckets)
}

func (h *Histogram) bucketsAreEmpty(buckets []*Bucket) bool {
	for _, b := range buckets {
		if b.Count > 0 || !h.bucketsAreEmpty(b.Buckets) {
			return false
		}
	}

	return true
}
