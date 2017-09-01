package model

import (
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
	elastic "gopkg.in/olivere/elastic.v5"
)

// FetchResultsSummary returns a histogram summarizing prediction results
func FetchResultsSummary(esClient *elastic.Client, pipelineURI string, index string, dataset string) (*Histogram, error) {

	results, err := FetchResults(esClient, pipelineURI, index, dataset)
	if err != nil {
		return nil, err
	}

	// currently only support a single result column.
	if len(results.Metadata) > 1 {
		log.Warnf("Result contains %s variables, expected 1.  Additional variables will be ignored.", len(results.Metadata))
	}

	varType := results.Metadata[0].Type
	var histogram *Histogram

	if IsCategorical(varType) {
		histogram, err = computeCategoricalHistogram(results)
		if err != nil {
			return nil, err
		}
	} else if IsNumerical(varType) {
		histogram, err = computeNumericalHistogram(MaxNumBuckets, results)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.Errorf("unhandled histogram type %s", varType)
	}
	return histogram, nil
}

func computeNumericalHistogram(numBins int64, data *FilteredData) (*Histogram, error) {
	varName := data.Metadata[0].Name
	varType := data.Metadata[0].Type

	var min, max float64
	var bins []*Bucket

	if varType == FloatType {
		// compute extrema and bin sizes
		min, max = computeFloatMinMax(data)
		bins = computeFloatCounts(data, numBins, min, max)
	} else if varType == IntegerType {
		// compute extrema and generate bin counts
		intMin, intMax := computeIntMinMax(data)
		bins = computeIntegerCounts(data, numBins, intMin, intMax)
		min = float64(intMin)
		max = float64(intMax)
	} else {
		return nil, errors.Errorf("can't create numeric histogram for type %s", varType)
	}

	return &Histogram{
		Name: varName,
		Extrema: &Extrema{
			Min: min,
			Max: max,
		},
		Type:    "numerical",
		Buckets: bins,
	}, nil
}

func computeCategoricalHistogram(data *FilteredData) (*Histogram, error) {
	// generate the counts
	counts := map[string]int64{}
	for _, value := range data.Values {
		label := fmt.Sprintf("%v", value[0])
		if count, ok := counts[label]; ok {
			counts[label] = count + 1
		} else {
			counts[label] = 1
		}
	}

	// reformat as buckets
	bins := []*Bucket{}
	for key, value := range counts {
		bins = append(bins, &Bucket{
			Key:   key,
			Count: value,
		})
	}

	return &Histogram{
		Name:    data.Metadata[0].Name,
		Type:    "categorical",
		Buckets: bins,
	}, nil
}

func computeIntegerCounts(data *FilteredData, numBins int64, min int64, max int64) []*Bucket {
	// compute bin size and adjust bin count if necessary
	binSize := (max - min) / numBins
	if binSize < 1 {
		binSize = 1
		numBins = (max - min) + 1
	}

	// collect the counts
	counts := map[int64]int64{}
	for _, value := range data.Values {
		bin := (value[0].(int64) - min) / binSize
		key := min + bin*binSize

		if val, ok := counts[key]; ok {
			counts[key] = val + 1
		} else {
			counts[key] = 1
		}
	}

	// sort the keys
	keys := []int64{}
	for key := range counts {
		keys = append(keys, key)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		if keys[i] < keys[j] {
			return true
		}
		return false
	})

	// covert to buckets and write out ordered by label
	bins := []*Bucket{}
	for _, key := range keys {
		val := counts[int64(key)]
		bins = append(bins, &Bucket{
			Key:   strconv.FormatInt(key, 10),
			Count: val,
		})
	}
	return bins
}

func computeIntMinMax(data *FilteredData) (int64, int64) {
	min := int64(math.MaxInt64)
	max := int64(math.MinInt64)
	for _, value := range data.Values {
		v := value[0].(int64)
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

func computeFloatCounts(data *FilteredData, numBins int64, min float64, max float64) []*Bucket {
	// compute bin size and adjust bin count if necessary
	binSize := (max - min) / float64(numBins)

	// collect the counts
	counts := map[float64]int64{}
	for _, value := range data.Values {
		bin := (value[0].(float64) - min) / binSize
		key := min + bin*binSize

		if val, ok := counts[key]; ok {
			counts[key] = val + 1
		} else {
			counts[key] = 1
		}
	}

	// sort the keys
	keys := []float64{}
	for key := range counts {
		keys = append(keys, key)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		if keys[i] < keys[j] {
			return true
		}
		return false
	})

	// covert to buckets and write out ordered by label
	bins := []*Bucket{}
	for _, key := range keys {
		val := counts[float64(key)]
		bins = append(bins, &Bucket{
			Key:   strconv.FormatFloat(key, 'f', -1, 64),
			Count: val,
		})
	}
	return bins
}

func computeFloatMinMax(data *FilteredData) (float64, float64) {
	min := math.MaxFloat64
	max := -math.MaxFloat64
	for _, value := range data.Values {
		v := value[0].(float64)
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}
