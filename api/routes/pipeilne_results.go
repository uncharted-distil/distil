package routes

import (
	"bufio"
	"encoding/csv"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/model"
)

// PipelineResultHandler fetches predicted pipeline values and returns them to the client
// in a JSON structure
func PipelineResultHandler(esCtor elastic.ClientCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		index := pat.Param(r, "index")
		dataset := pat.Param(r, "dataset")
		pipelineURI, err := url.PathUnescape(pat.Param(r, "result-uri"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape result uri"))
			return
		}

		// load the result data from CSV
		file, err := os.Open(pipelineURI)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable open pipeline result file"))
			return
		}
		csvReader := csv.NewReader(bufio.NewReader(file))
		csvReader.TrimLeadingSpace = true
		records, err := csvReader.ReadAll()
		if err != nil {
			handleError(w, errors.Wrap(err, "unable load pipeline result as csv"))
			return
		}
		if len(records) <= 0 || len(records[0]) <= 0 {
			handleError(w, errors.Wrap(err, "pipeline csv empty"))
			return
		}

		// fetch the variable info to resolve its type - skip the first column since that will be the d3m_index value
		esClient, err := esCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		variable, err := model.FetchVariable(esClient, index, dataset, records[0][1])
		if err != nil {
			handleError(w, err)
			return
		}

		// populate the value array - skip the first line since it contains header info,
		// take only the second column since the first is reserved for the d3m_index.
		values := [][]interface{}{}
		for i := 1; i < len(records); i++ {
			var val interface{}
			err = nil

			switch variable.Type {
			case model.IntegerType:
				val, err = strconv.Atoi(records[i][1])
			case model.FloatType:
				val, err = strconv.ParseFloat(records[i][1], 64)
			case model.CategoricalType:
			case model.TextType:
			case model.DateTimeType:
			case model.OrdinalType:
				val = records[i][0]
			case model.BoolType:
				val, err = strconv.ParseBool(records[i][1])
			default:
				val = records[i][0]
			}
			// handle the parsed result/error
			if err != nil {
				handleError(w, errors.Wrap(err, "failed csv value parsing"))
			} else {
				values = append(values, []interface{}{val})
			}
		}

		// write the data into the filtered data struct
		result := &model.FilteredData{
			Name: dataset,
			Metadata: []*model.Variable{
				&model.Variable{
					Name: variable.Name,
					Type: variable.Type,
				},
			},
			Values: values,
		}

		// marshall data and sent the response back
		err = handleJSON(w, result)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal pipeline result into JSON"))
			return
		}
	}
}
