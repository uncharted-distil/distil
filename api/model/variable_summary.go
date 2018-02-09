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
	// maxNumBuckets is the maximum number of buckets to use for histograms
	maxNumBuckets = 50
)

// Extrema represents the extrema for a single variable.
type Extrema struct {
	Name string  `json:"-"`
	Type string  `json:"-"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
}

// Bucket represents a single histogram bucket.
type Bucket struct {
	Key     string    `json:"key"`
	Count   int64     `json:"count"`
	Buckets []*Bucket `json:"buckets,omitempty"`
}

// Histogram represents a single variable histogram.
type Histogram struct {
	Name    string    `json:"name"`
	Type    string    `json:"type"`
	Dataset string    `json:"dataset"`
	VarType string    `json:"varType"`
	NumRows int       `json:"numRows"`
	Extrema *Extrema  `json:"extrema,omitempty"`
	Buckets []*Bucket `json:"buckets"`
}

// GetBucketCount calculates the number of buckets for the extrema.
func (e *Extrema) GetBucketCount() int {
	// Floating point uses the maximum number of buckets
	if IsFloatingPoint(e.Type) {
		return maxNumBuckets
	}

	// Integer should have 1 value per bucket up to the maximum number of buckets
	max := int(e.Max)
	min := int(e.Min)

	if max-min+1 < maxNumBuckets {
		return max - min + 1
	}

	return maxNumBuckets
}

// GetBucketInterval calculates the size of the buckets given the extrema.
func (e *Extrema) GetBucketInterval() float64 {
	if IsFloatingPoint(e.Type) {
		return (e.Max - e.Min) / float64(e.GetBucketCount())
	}

	// Account for inclusive range
	return (e.Max - e.Min + 1) / float64(e.GetBucketCount())
}
