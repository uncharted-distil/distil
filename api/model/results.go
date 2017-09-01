package model

import (
	"bufio"
	"encoding/csv"
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

		switch variable.Type {
		case IntegerType:
			val, err = strconv.ParseInt(records[i][1], 10, 64)
		case FloatType:
			val, err = strconv.ParseFloat(records[i][1], 64)
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
