package model

import (
	"bufio"
	"encoding/csv"
	"math"
	"os"
	"strconv"

	elastic "gopkg.in/olivere/elastic.v5"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
)

// FetchResults returns the set of test predictions made by a given pipeline.
func FetchResults(client *elastic.Client, pipelineURI string, index string, dataset string) (*FilteredData, error) {
	// load the result data from CSV
	file, err := os.Open(pipelineURI)
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
	variable, err := FetchVariable(client, index, dataset, records[0][1])
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
		case IntegerType:
			val, err = strconv.ParseInt(records[i][1], 10, 64)
			if err != nil {
				var floatVal float64
				floatVal, err = strconv.ParseFloat(records[i][1], 64)
				if err == nil {
					val = roundToInt(floatVal)
				}
			}
		case FloatType:
			val, err = strconv.ParseFloat(records[i][1], 64)
			if err != nil {
				var intVal int64
				intVal, err = strconv.ParseInt(records[i][1], 10, 64)
				if err == nil {
					val = float64(intVal)
				}
			}
		case CategoricalType:
			fallthrough
		case TextType:
			fallthrough
		case DateTimeType:
			fallthrough
		case OrdinalType:
			val = records[i][1]
		case BoolType:
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
	return &FilteredData{
		Name: dataset,
		Metadata: []*Variable{
			&Variable{
				Name: variable.Name,
				Type: variable.Type,
			},
		},
		Values: values,
	}, nil
}

func roundToInt(a float64) int64 {
	if a < 0 {
		return int64(math.Ceil(a - 0.5))
	}
	return int64(math.Floor(a + 0.5))
}
