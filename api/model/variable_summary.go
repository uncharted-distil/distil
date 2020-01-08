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
	Label        string     `json:"label"`
	Key          string     `json:"key"`
	Description  string     `json:"description"`
	Type         string     `json:"type"`
	VarType      string     `json:"varType"`
	Dataset      string     `json:"dataset"`
	SolutionID   string     `json:"solutionId,omitempty"`
	Baseline     *Histogram `json:"baseline"`
	Filtered     *Histogram `json:"filtered"`
	Timeline     *Histogram `json:"timeline"`
	TimelineType string     `json:"timelineType"`
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
