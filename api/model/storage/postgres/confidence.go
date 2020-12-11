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

package postgres

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// FetchConfidenceSummary fetches a histogram of the confidence and explanations associated with a set of classification predictions.
func (s *Storage) FetchConfidenceSummary(dataset string, storageName string, resultURI string, filterParams *api.FilterParams, mode api.SummaryMode) (map[string]*api.VariableSummary, error) {
	explainFields := s.listExplainFields()
	storageNameResult := s.getResultTable(storageName)
	targetName, err := s.getResultTargetName(storageNameResult, resultURI)
	if err != nil {
		return nil, err
	}

	variable, err := s.getResultTargetVariable(dataset, targetName)
	if err != nil {
		return nil, err
	}

	explainedSummaries := map[string]*api.VariableSummary{}
	for _, explainName := range explainFields {
		var baseline *api.Histogram
		var filtered *api.Histogram
		baseline, err = s.fetchExplainHistogram(dataset, storageName, targetName, explainName, resultURI, nil, mode)
		if err != nil {
			return nil, err
		}
		// if the field does not exist, then no baseline will be returned
		if baseline == nil {
			continue
		}
		if !filterParams.Empty() {
			filtered, err = s.fetchExplainHistogram(dataset, storageName, targetName, explainName, resultURI, filterParams, mode)
			if err != nil {
				return nil, err
			}
		}

		explainedSummaries[explainName] = &api.VariableSummary{
			Label:    variable.DisplayName,
			Key:      variable.StorageName,
			Type:     model.NumericalType,
			VarType:  variable.Type,
			Baseline: baseline,
			Filtered: filtered,
		}
	}

	return explainedSummaries, nil
}

func (s *Storage) fetchHistograms(dataset string, storageName string, variable *model.Variable, targetName string,
	resultURI string, filterParams *api.FilterParams, mode api.SummaryMode) (map[string]*api.Histogram, error) {
	explainFields := s.listExplainFields()

	explainedSummaries := map[string]*api.Histogram{}
	for _, explainName := range explainFields {
		// get the histogram for that explaination field
		histo, err := s.fetchExplainHistogram(dataset, storageName, targetName, explainName, resultURI, filterParams, mode)
		if err != nil {
			return nil, err
		}
		explainedSummaries[explainName] = histo
	}

	return explainedSummaries, nil
}

func (s *Storage) fetchExplainHistogram(dataset string, storageName string, targetName string, explainFieldName string,
	resultURI string, filterParams *api.FilterParams, mode api.SummaryMode) (*api.Histogram, error) {
	// use a numerical sub select
	field := NewNumericalFieldSubSelect(s, dataset, storageName, explainFieldName, explainFieldName, model.IntegerType, "", s.explainSubSelect(storageName, explainFieldName))
	return field.fetchHistogram(filterParams, false, 20)
}

func (s *Storage) listExplainFields() []string {
	v := reflect.TypeOf(&api.SolutionExplainValues{}).Elem()
	jsonNames := []string{}
	for j := 0; j < v.NumField(); j++ {
		f := v.Field(j)
		if f.Type.Kind() == reflect.Float64 {
			tag := f.Tag.Get("json")
			tag = strings.Split(tag, ",")[0]
			jsonNames = append(jsonNames, tag)
		}
	}

	return jsonNames
}

func (s *Storage) explainSubSelect(storageName string, fieldName string) func() string {
	return func() string {
		return fmt.Sprintf("(SELECT (explain_values ->> '%s')::double precision as \"%s\", index as \"%s\" from %s)",
			fieldName, fieldName, model.D3MIndexFieldName, s.getResultTable(storageName))
	}
}
