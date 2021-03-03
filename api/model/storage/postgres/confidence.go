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
		baseline, err = s.fetchExplainHistogram(dataset, storageName, targetName, explainName, resultURI, api.GetBaselineFilter(filterParams), mode)
		if err != nil {
			return nil, err
		}
		// if the field does not exist, then no baseline will be returned
		if baseline == nil {
			continue
		}
		if !filterParams.Empty(true) {
			filtered, err = s.fetchExplainHistogram(dataset, storageName, targetName, explainName, resultURI, filterParams, mode)
			if err != nil {
				return nil, err
			}
		}

		explainedSummaries[explainName] = &api.VariableSummary{
			Label:    variable.DisplayName,
			Key:      variable.Key,
			Type:     model.NumericalType,
			VarType:  variable.Type,
			Baseline: baseline,
			Filtered: filtered,
		}
	}

	return explainedSummaries, nil
}

func (s *Storage) fetchExplainHistogram(dataset string, storageName string, targetName string, explainFieldName string,
	resultURI string, filterParams *api.FilterParams, mode api.SummaryMode) (*api.Histogram, error) {
	explainFieldAlias := fmt.Sprintf("%s_nested", explainFieldName)
	// use a numerical sub select
	field := NewNumericalFieldSubSelect(s, dataset, storageName, explainFieldAlias, explainFieldName, model.RealType, "", s.explainSubSelect(storageName, explainFieldName, explainFieldAlias))

	// use predefined ranged of [0,1] for everything except rank - we'll leave that as nil so that it
	// will be computed when the histogram is fetched
	var extrema *api.Extrema
	if explainFieldName != "rank" {
		extrema, _ = api.NewExtrema(0.0, 1.0)
	}

	// filter for the single result confidences instead of having all result confidences
	if filterParams == nil {
		filterParams = &api.FilterParams{}
	}

	// filter info derived from the sub select function
	filterParams.Filters = append(filterParams.Filters, model.NewCategoricalFilter("result_key", model.IncludeFilter, []string{resultURI}))

	return field.fetchHistogramByResult(resultURI, filterParams, extrema, 20)
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

func (s *Storage) explainSubSelect(storageName string, fieldName string, aliasName string) func() string {
	return func() string {
		return fmt.Sprintf(`(
			SELECT (explain_values ->> '%s')::double precision as "%s", result_id as result_key, b.* from %s as d
			inner join %s as b on b."d3mIndex" = d.index
			) as data`,
			fieldName, aliasName, s.getResultTable(storageName), storageName)
	}
}
