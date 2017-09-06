package elastic

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
	log "github.com/unchartedsoftware/plog"
)

// FetchResults returns the set of test predictions made by a given pipeline.
func (s *Storage) FetchResults(dataset string, resultURI string, index string) (*model.FilteredData, error) {
	// load the result data from CSV
	file, err := os.Open(resultURI)
	if err != nil {
		return nil, errors.Wrap(err, "unable open pipeline result file")
	}
	csvReader := csv.NewReader(bufio.NewReader(file))
	csvReader.TrimLeadingSpace = true
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, errors.Wrap(err, "unable load pipeline result as csv")
	}
	if len(records) <= 0 || len(records[0]) <= 0 {
		return nil, errors.Wrap(err, "pipeline csv empty")
	}

	// currently only support a single result column.
	if len(records[0]) > 2 {
		log.Warnf("Result contains %s columns, expected 2.  Additional columns will be ignored.", len(records[0]))
	}

	// fetch the variable info to resolve its type - skip the first column since that will be the d3m_index value
	variable, err := model.FetchVariable(s.client, index, dataset, records[0][1])
	if err != nil {
		return nil, err
	}

	// populate the value array - skip the first line since it contains header info,
	// take only the second column since the first is reserved for the d3m_index.
	values := [][]interface{}{}
	for i := 1; i < len(records); i++ {
		var val interface{}
		err = nil

		// parse input values - fix up number types if ta2 returned them in an unexpected format
		switch variable.Type {
		case model.IntegerType:
			val, err = strconv.ParseInt(records[i][1], 10, 64)
			if err != nil {
				var floatVal float64
				floatVal, err = strconv.ParseFloat(records[i][1], 64)
				if err == nil {
					val = s.roundToInt(floatVal)
				}
			}
		case model.FloatType:
			val, err = strconv.ParseFloat(records[i][1], 64)
			if err != nil {
				var intVal int64
				intVal, err = strconv.ParseInt(records[i][1], 10, 64)
				if err == nil {
					val = float64(intVal)
				}
			}
		case model.CategoricalType:
			fallthrough
		case model.TextType:
			fallthrough
		case model.DateTimeType:
			fallthrough
		case model.OrdinalType:
			val = records[i][1]
		case model.BoolType:
			val, err = strconv.ParseBool(records[i][1])
		default:
			val = records[i][1]
		}
		// handle the parsed result/error
		if err != nil {
			return nil, errors.Wrap(err, "failed csv value parsing")
		}
		values = append(values, []interface{}{val})
	}

	// write the data into the filtered data struct
	return &model.FilteredData{
		Name: dataset,
		Metadata: []*model.Variable{
			{
				Name: variable.Name,
				Type: variable.Type,
			},
		},
		Values: values,
	}, nil
}

func (s *Storage) roundToInt(a float64) int64 {
	if a < 0 {
		return int64(math.Ceil(a - 0.5))
	}
	return int64(math.Floor(a + 0.5))
}

// FetchResultsSummary returns a histogram summarizing prediction results
func (s *Storage) FetchResultsSummary(dataset string, resultURI string, index string) (*model.Histogram, error) {

	results, err := s.FetchResults(dataset, resultURI, index)
	if err != nil {
		return nil, err
	}

	// currently only support a single result column.
	if len(results.Metadata) > 1 {
		log.Warnf("Result contains %s variables, expected 1.  Additional variables will be ignored.", len(results.Metadata))
	}

	varType := results.Metadata[0].Type
	var histogram *model.Histogram

	if model.IsCategorical(varType) {
		histogram, err = s.computeCategoricalHistogram(results)
		if err != nil {
			return nil, err
		}
	} else if model.IsNumerical(varType) {
		histogram, err = s.computeNumericalHistogram(model.MaxNumBuckets, results)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.Errorf("unhandled histogram type %s", varType)
	}
	return histogram, nil
}

func (s *Storage) computeNumericalHistogram(numBins int64, data *model.FilteredData) (*model.Histogram, error) {
	varName := data.Metadata[0].Name
	varType := data.Metadata[0].Type

	var min, max float64
	var bins []*model.Bucket

	if varType == model.FloatType {
		// compute extrema and bin sizes
		min, max = s.computeFloatMinMax(data)
		bins = s.computeFloatCounts(data, numBins, min, max)
	} else if varType == model.IntegerType {
		// compute extrema and generate bin counts
		intMin, intMax := s.computeIntMinMax(data)
		bins = s.computeIntegerCounts(data, numBins, intMin, intMax)
		min = float64(intMin)
		max = float64(intMax)
	} else {
		return nil, errors.Errorf("can't create numeric histogram for type %s", varType)
	}

	return &model.Histogram{
		Name: varName,
		Extrema: &model.Extrema{
			Min: min,
			Max: max,
		},
		Type:    "numerical",
		Buckets: bins,
	}, nil
}

func (s *Storage) computeCategoricalHistogram(data *model.FilteredData) (*model.Histogram, error) {
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

	// sort the keys to guarantee a stable ordering
	keys := []string{}
	for key := range counts {
		keys = append(keys, key)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		if keys[i] < keys[j] {
			return true
		}
		return false
	})

	// reformat as buckets
	bins := []*model.Bucket{}
	for _, key := range keys {
		value := counts[key]
		bins = append(bins, &model.Bucket{
			Key:   key,
			Count: value,
		})
	}

	return &model.Histogram{
		Name:    data.Metadata[0].Name,
		Type:    "categorical",
		Buckets: bins,
	}, nil
}

func (s *Storage) computeIntegerCounts(data *model.FilteredData, numBins int64, min int64, max int64) []*model.Bucket {
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
	bins := []*model.Bucket{}
	for _, key := range keys {
		val := counts[int64(key)]
		bins = append(bins, &model.Bucket{
			Key:   strconv.FormatInt(key, 10),
			Count: val,
		})
	}
	return bins
}

func (s *Storage) computeIntMinMax(data *model.FilteredData) (int64, int64) {
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

func (s *Storage) computeFloatCounts(data *model.FilteredData, numBins int64, min float64, max float64) []*model.Bucket {
	// compute bin size and adjust bin count if necessary
	binSize := (max - min) / float64(numBins)

	// collect the counts
	counts := map[float64]int64{}
	for _, value := range data.Values {
		bin := math.Floor((value[0].(float64) - min) / binSize)
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
	bins := []*model.Bucket{}
	for _, key := range keys {
		val := counts[float64(key)]
		bins = append(bins, &model.Bucket{
			Key:   strconv.FormatFloat(key, 'f', -1, 64),
			Count: val,
		})
	}
	return bins
}

func (s *Storage) computeFloatMinMax(data *model.FilteredData) (float64, float64) {
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
