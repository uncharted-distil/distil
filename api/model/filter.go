//
//   Copyright © 2021 Uncharted Software Inc.
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

import (
	encoding "encoding/json"
	"fmt"
	"math"
	"sort"

	"github.com/pkg/errors"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil/api/util/json"
)

// DataMode defines the data filter modes.
type DataMode int

const (
	// DefaultDataMode use the id field for filtering, ex. clustering not applied
	DefaultDataMode = iota + 1
	// ClusterDataMode use computed cluster information for filtering if availble, ex. timeseries clusters
	ClusterDataMode
)

// DataModeFromString creates a DataMode from the supplied string
func DataModeFromString(s string) (DataMode, error) {
	switch s {
	case "cluster":
		return ClusterDataMode, nil
	case "default":
		return DefaultDataMode, nil
	default:
		return 0, errors.Errorf("%s is not a valid DataMode", s)
	}
}

// FilterParams defines the set of filters to use. Note that this is to be used
// by the server only, and not the client. Filters are gathered by mode (include/exclude),
// with each mode being a list of features that are used as filters.
type FilterParams struct {
	Size      int          `json:"size"`
	Filters   []*FilterSet `json:"filters"`
	Variables []string     `json:"variables"`
	DataMode  DataMode     `json:"dataMode"`
	Invert    bool         `json:"invert"`
}

// FilterParamsRaw defines the set of numeric range and categorical filters. Variables
// with no range or category filters are also allowed.
type FilterParamsRaw struct {
	Size       int          `json:"size"`
	Highlights FilterObject `json:"highlights"`
	Filters    FilterObject `json:"filters"`
	Variables  []string     `json:"variables"`
	DataMode   DataMode     `json:"dataMode"`
}

// FilterObject captures a collection of invertable filters.
type FilterObject struct {
	List   []*model.Filter `json:"list"`
	Invert bool            `json:"invert"`
}

// FilterSet captures a set of filters representing one subset of data.
type FilterSet struct {
	FeatureFilters []FilterObject `json:"featureFilters"`
	Mode           string         `json:"mode"`
}

// NewFilterParamsFromFilters creates a wrapping container for all filters.
func NewFilterParamsFromFilters(filters []*model.Filter) *FilterParams {
	// group filters by feature and mode
	params := &FilterParams{
		Filters: []*FilterSet{},
	}

	// add filters to the params
	for _, f := range filters {
		params.AddFilter(f)
	}

	return params
}

// NewFilterParamsFromRaw creates a wrapping container from raw filter params.
func NewFilterParamsFromRaw(raw *FilterParamsRaw) *FilterParams {
	rawClone := raw.Clone()

	params := &FilterParams{
		Size:      rawClone.Size,
		Variables: rawClone.Variables,
		DataMode:  rawClone.DataMode,
		Filters:   []*FilterSet{},
	}

	// add filters and highlights to the params
	for _, f := range rawClone.Filters.List {
		params.AddFilter(f)
	}
	for _, h := range rawClone.Highlights.List {
		params.AddFilter(h)
	}
	// TODO: FIGURE OUT A NICE WAY TO INVERT THINGS!
	if len(rawClone.Filters.List) > 0 {
		for _, set := range params.Filters {
			if set.Mode == rawClone.Filters.List[0].Mode {
				for i := range set.FeatureFilters {
					set.FeatureFilters[i].Invert = rawClone.Filters.Invert
				}
			}
		}
	}
	if len(rawClone.Highlights.List) > 0 {
		for _, set := range params.Filters {
			if set.Mode == rawClone.Highlights.List[0].Mode {
				for i := range set.FeatureFilters {
					set.FeatureFilters[i].Invert = rawClone.Highlights.Invert
				}
			}
		}
	}

	return params
}

// GetBaselineFilter returns a filter params that only has the baseline filters.
func GetBaselineFilter(filterParam *FilterParams) *FilterParams {
	if filterParam == nil {
		return nil
	}

	// highlights should not be applied to the baseline
	clone := &FilterParams{
		Filters: []*FilterSet{},
	}
	for _, filters := range filterParam.Filters {
		baselineFilters := []FilterObject{}
		for _, f := range filters.FeatureFilters {
			baseline := f.getBaselineFilter()
			if len(baseline) > 0 {
				baselineFilters = append(baselineFilters, FilterObject{
					Invert: f.Invert,
					List:   f.getBaselineFilter(),
				})
			}
		}
		if len(baselineFilters) > 0 {
			clone.Filters = append(clone.Filters, &FilterSet{
				FeatureFilters: baselineFilters,
				Mode:           filters.Mode,
			})
		}
	}
	clone.Variables = append(clone.Variables, filterParam.Variables...)
	clone.Size = filterParam.Size
	clone.DataMode = filterParam.DataMode
	return clone
}

func (f FilterObject) getBaselineFilter() []*model.Filter {
	baseline := []*model.Filter{}
	for _, filter := range f.List {
		if filter.IsBaselineFilter {
			baseline = append(baseline, filter)
		}
	}

	return baseline
}

// Clone returns a deep copy of the filter params.
func (f *FilterParams) Clone() *FilterParams {
	clone := &FilterParams{
		Filters: []*FilterSet{},
	}
	for _, filters := range f.Filters {
		featureSet := &FilterSet{
			Mode:           filters.Mode,
			FeatureFilters: []FilterObject{},
		}
		for _, fo := range filters.FeatureFilters {
			cloneFilterObject := FilterObject{
				Invert: fo.Invert,
				List:   []*model.Filter{},
			}
			for _, f := range fo.List {
				c := *f
				cloneFilterObject.List = append(cloneFilterObject.List, &c)
			}
			featureSet.FeatureFilters = append(featureSet.FeatureFilters, cloneFilterObject)
		}
		clone.Filters = append(clone.Filters, featureSet)
	}
	clone.Invert = f.Invert
	clone.Variables = append(clone.Variables, f.Variables...)
	clone.Size = f.Size
	clone.DataMode = f.DataMode
	return clone
}

// AddFilter adds a filter to the filter params, inserting it in the proper collection.
func (f *FilterParams) AddFilter(filter *model.Filter) {
	// currently assume all include filters are one filter set, and exclude another
	// need to add it to the right mode (include, exclude)
	for _, set := range f.Filters {
		if set.Mode == filter.Mode {
			// find the list of filters for that feature
			for i, feature := range set.FeatureFilters {
				if feature.List[0].Key == filter.Key {
					set.FeatureFilters[i].List = append(set.FeatureFilters[i].List, filter)
					return
				}
			}

			// feature not filtered yet
			set.FeatureFilters = append(set.FeatureFilters, FilterObject{
				Invert: false,
				List:   []*model.Filter{filter},
			})
			return
		}
	}
	// no filter for that mode exists yet
	f.Filters = append(f.Filters, &FilterSet{
		Mode: filter.Mode,
		FeatureFilters: []FilterObject{{
			Invert: false,
			List:   []*model.Filter{filter},
		}},
	})
}

// Empty returns if the filter set is empty.
func (f *FilterParams) Empty(ignoreBaselineFilters bool) bool {
	for _, set := range f.Filters {
		for _, filters := range set.FeatureFilters {
			for _, filter := range filters.List {
				if !filter.IsBaselineFilter || !ignoreBaselineFilters {
					return false
				}
			}
		}
	}
	return true
}

// AddVariable adds a variable, preventing duplicates
func (f *FilterParams) AddVariable(nv string) {
	for _, v := range f.Variables {
		if v == nv {
			return
		}
	}
	f.Variables = append(f.Variables, nv)
}

// InvertFilters inverts filters and highlights.
func (f *FilterParams) InvertFilters() {
	for _, set := range f.Filters {
		for _, fo := range set.FeatureFilters {
			fo.Invert = !fo.Invert
		}
	}
	f.Invert = !f.Invert
}

// Empty returns if the filter set is empty.
func (f *FilterParamsRaw) Empty(ignoreBaselineFilters bool) bool {
	for _, filter := range f.Filters.List {
		if !filter.IsBaselineFilter || !ignoreBaselineFilters {
			return false
		}
	}
	for _, highlight := range f.Highlights.List {
		if !highlight.IsBaselineFilter || !ignoreBaselineFilters {
			return false
		}
	}
	return true
}

// Clone returns a deep copy of the filter params.
func (f *FilterParamsRaw) Clone() *FilterParamsRaw {
	clone := &FilterParamsRaw{}

	for _, h := range f.Highlights.List {
		c := *h
		clone.Highlights.List = append(clone.Highlights.List, &c)
	}
	for _, f := range f.Filters.List {
		c := *f
		clone.Filters.List = append(clone.Filters.List, &c)
	}
	clone.Highlights.Invert = f.Highlights.Invert
	clone.Filters.Invert = f.Filters.Invert
	clone.Variables = append(clone.Variables, f.Variables...)
	clone.Size = f.Size
	clone.DataMode = f.DataMode
	return clone
}

func filtersEqual(first *model.Filter, second *model.Filter) bool {
	baseEquals := first.Key == second.Key &&
		first.Min == second.Min &&
		first.Max == second.Max &&
		first.Mode == second.Mode
	boundsEquals := (first.Bounds == nil && second.Bounds == nil) ||
		(first.Bounds != nil && second.Bounds != nil &&
			first.Bounds.MinX == second.Bounds.MinX &&
			first.Bounds.MaxX == second.Bounds.MaxX &&
			first.Bounds.MinY == second.Bounds.MinY &&
			first.Bounds.MaxY == second.Bounds.MaxY)

	return baseEquals && boundsEquals && model.StringSliceEqual(first.Categories, second.Categories)
}

// Merge merges another set of filter params into this set, expanding all
// properties.
func (f *FilterParams) Merge(other *FilterParamsRaw) {

	// If the filters has a nil or negative value, we use the value use by default on distil-model
	// https://github.com/uncharted-distil/distil/blob/master/api/model/storage/postgres/request.go#L239
	if other.Size >= 0 {
		f.Size = other.Size
	}

	for _, highlight := range other.Highlights.List {
		found := false
		for _, set := range f.Filters {
			if set.Mode == highlight.Mode {
				for _, filters := range set.FeatureFilters {
					for _, currentFilter := range filters.List {
						if filtersEqual(highlight, currentFilter) {
							found = true
							break
						}
					}
				}
			}
		}
		if !found {
			f.AddFilter(highlight)
		}
	}

	for _, filter := range other.Filters.List {
		found := false
		for _, set := range f.Filters {
			if set.Mode == filter.Mode {
				for _, filters := range set.FeatureFilters {
					for _, currentFilter := range filters.List {
						if filtersEqual(filter, currentFilter) {
							found = true
							break
						}
					}
				}
			}
		}
		if !found {
			f.AddFilter(filter)
		}
	}

	for _, variable := range other.Variables {
		found := false
		for _, currentVariable := range f.Variables {
			if variable == currentVariable {
				found = true
				break
			}
		}
		if !found {
			f.Variables = append(f.Variables, variable)
		}
	}
}

// MergeParams merges another set of filter params into this set, expanding all
// properties.
func (f *FilterParams) MergeParams(other *FilterParams) {

	// If the filters has a nil or negative value, we use the value use by default on distil-model
	// https://github.com/uncharted-distil/distil/blob/master/api/model/storage/postgres/request.go#L239
	if other.Size >= 0 {
		f.Size = other.Size
	}

	for _, set := range other.Filters {
		for _, features := range set.FeatureFilters {
			for _, filter := range features.List {
				found := false
				for _, setOther := range f.Filters {
					if setOther.Mode == set.Mode {
						for _, filters := range setOther.FeatureFilters {
							for _, currentFilter := range filters.List {
								if filtersEqual(filter, currentFilter) {
									found = true
									break
								}
							}
						}
					}
				}
				if !found {
					f.AddFilter(filter)
				}
			}
		}
	}

	for _, variable := range other.Variables {
		found := false
		for _, currentVariable := range f.Variables {
			if variable == currentVariable {
				found = true
				break
			}
		}
		if !found {
			f.Variables = append(f.Variables, variable)
		}
	}
}

// MergeFilterObjects merges a slice of filter objects with the existing filter params.
func (f *FilterParams) MergeFilterObjects(filters []FilterObject) {
	for _, features := range filters {
		for _, filter := range features.List {
			found := false
			for _, setOther := range f.Filters {
				for _, filters := range setOther.FeatureFilters {
					for _, currentFilter := range filters.List {
						if filtersEqual(filter, currentFilter) {
							found = true
							break
						}
					}
				}
			}
			if !found {
				f.AddFilter(filter)
			}
		}
	}
}

// Column represents a column for filtered data.
type Column struct {
	Label  string  `json:"label"`
	Key    string  `json:"key"`
	Type   string  `json:"type"`
	Weight float64 `json:"weight"`
	Index  int     `json:"index"`
}

// FilteredDataValue represents a data value combined with an optional weight.
type FilteredDataValue struct {
	Value      interface{}     `json:"value"`
	Weight     float64         `json:"weight,omitempty"`
	Confidence NullableFloat64 `json:"confidence,omitempty"`
	Rank       NullableFloat64 `json:"rank,omitempty"`
}

// FilteredData provides the metadata and raw data values that match a supplied
// input filter.
type FilteredData struct {
	NumRows         int                    `json:"numRows"`
	NumRowsFiltered int                    `json:"numRowsFiltered"`
	Columns         map[string]*Column     `json:"columns"`
	Values          [][]*FilteredDataValue `json:"values"`
}

// EmptyFilterData returns an empty FilteredData object
func EmptyFilterData() *FilteredData {
	return &FilteredData{NumRows: 0, NumRowsFiltered: 0, Columns: map[string]*Column{}, Values: [][]*FilteredDataValue{}}
}

// GetFilterVariables builds the filtered list of fields based on the filtering parameters.
func GetFilterVariables(filterVariables []string, variables []*model.Variable) []*model.Variable {

	variableLookup := make(map[string]*model.Variable)
	for _, v := range variables {
		variableLookup[v.Key] = v
	}

	filtered := make([]*model.Variable, 0)
	for _, variable := range filterVariables {

		v := variableLookup[variable]
		if v == nil {
			continue
		}
		filtered = append(filtered, v)
	}

	return filtered
}

func parseFilter(filter map[string]interface{}) (*model.Filter, error) {

	// type
	typ, ok := json.String(filter, "type")
	if !ok {
		return nil, errors.Errorf("no `type` provided for filter")
	}

	// mode
	mode, ok := json.String(filter, "mode")
	if !ok {
		return nil, errors.Errorf("no `mode` provided for filter")
	}

	// TODO: update to a switch statement with a default to error

	// datetime
	if typ == model.DatetimeFilter {
		key, ok := json.String(filter, "key")
		if !ok {
			return nil, errors.Errorf("no `key` provided for filter")
		}
		min, ok := json.Float(filter, "min")
		if !ok {
			return nil, errors.Errorf("no `min` provided for filter")
		}
		max, ok := json.Float(filter, "max")
		if !ok {
			return nil, errors.Errorf("no `max` provided for filter")
		}

		return model.NewDatetimeFilter(key, mode, min, max), nil
	}

	// numeric
	if typ == model.NumericalFilter {
		key, ok := json.String(filter, "key")
		if !ok {
			return nil, errors.Errorf("no `key` provided for filter")
		}
		min, ok := json.Float(filter, "min")
		if !ok {
			return nil, errors.Errorf("no `min` provided for filter")
		}
		max, ok := json.Float(filter, "max")
		if !ok {
			return nil, errors.Errorf("no `max` provided for filter")
		}
		return model.NewNumericalFilter(key, mode, min, max), nil
	}

	// vector
	if typ == model.VectorFilter {
		key, ok := json.String(filter, "key")
		if !ok {
			return nil, errors.Errorf("no `key` provided for filter")
		}
		min, ok := json.Float(filter, "min")
		if !ok {
			return nil, errors.Errorf("no `min` provided for filter")
		}
		max, ok := json.Float(filter, "max")
		if !ok {
			return nil, errors.Errorf("no `max` provided for filter")
		}
		nestedType, ok := json.String(filter, "nestedType")
		if !ok {
			return nil, errors.Errorf("no `nestedType` provided for filter")
		}
		return model.NewVectorFilter(key, nestedType, mode, min, max), nil
	}

	// bivariate
	if typ == model.BivariateFilter {
		key, ok := json.String(filter, "key")
		if !ok {
			return nil, errors.Errorf("no `key` provided for filter")
		}
		minX, ok := json.Float(filter, "minX")
		if !ok {
			return nil, errors.Errorf("no `minX` provided for filter")
		}
		maxX, ok := json.Float(filter, "maxX")
		if !ok {
			return nil, errors.Errorf("no `maxX` provided for filter")
		}
		minY, ok := json.Float(filter, "minY")
		if !ok {
			return nil, errors.Errorf("no `minY` provided for filter")
		}
		maxY, ok := json.Float(filter, "maxY")
		if !ok {
			return nil, errors.Errorf("no `maxY` provided for filter")
		}
		return model.NewBivariateFilter(key, mode, minX, maxX, minY, maxY), nil
	}

	// geobounds
	if typ == model.GeoBoundsFilter {
		key, ok := json.String(filter, "key")
		if !ok {
			return nil, errors.Errorf("no `key` provided for filter")
		}
		minX, ok := json.Float(filter, "minX")
		if !ok {
			return nil, errors.Errorf("no `minX` provided for filter")
		}
		maxX, ok := json.Float(filter, "maxX")
		if !ok {
			return nil, errors.Errorf("no `maxX` provided for filter")
		}
		minY, ok := json.Float(filter, "minY")
		if !ok {
			return nil, errors.Errorf("no `minY` provided for filter")
		}
		maxY, ok := json.Float(filter, "maxY")
		if !ok {
			return nil, errors.Errorf("no `maxY` provided for filter")
		}
		return model.NewGeoBoundsFilter(key, mode, minX, maxX, minY, maxY), nil
	}

	// categorical
	if typ == model.CategoricalFilter {
		key, ok := json.String(filter, "key")
		if !ok {
			return nil, errors.Errorf("no `key` provided for filter")
		}
		categories, ok := json.StringArray(filter, "categories")
		if !ok {
			return nil, errors.Errorf("no `categories` provided for filter")
		}
		return model.NewCategoricalFilter(key, mode, categories), nil
	}

	// cluster
	if typ == model.ClusterFilter {
		key, ok := json.String(filter, "key")
		if !ok {
			return nil, errors.Errorf("no `key` provided for filter")
		}
		categories, ok := json.StringArray(filter, "categories")
		if !ok {
			return nil, errors.Errorf("no `categories` provided for filter")
		}
		return model.NewClusterFilter(key, mode, categories), nil
	}

	// text
	if typ == model.TextFilter {
		key, ok := json.String(filter, "key")
		if !ok {
			return nil, errors.Errorf("no `key` provided for filter")
		}
		categories, ok := json.StringArray(filter, "categories")
		if !ok {
			return nil, errors.Errorf("no `categories` provided for filter")
		}
		return model.NewTextFilter(key, mode, categories), nil
	}

	// row
	if typ == model.RowFilter {
		indices, ok := json.StringArray(filter, "d3mIndices")
		if !ok {
			return nil, errors.Errorf("no `d3mIndices` provided for filter")
		}
		return model.NewRowFilter(mode, indices), nil
	}

	return nil, fmt.Errorf("filter not recognized")
}

// ParseFilterParamsFromJSONRaw parses filter parameters out of a json.RawMessage
func ParseFilterParamsFromJSONRaw(raw encoding.RawMessage) (*FilterParamsRaw, error) {
	filterParamsMap, err := json.Unmarshal(raw)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse raw filter params")
	}
	return ParseFilterParamsFromJSON(filterParamsMap)
}

// ParseFilterParamsFromJSON parses filter parameters out of a map[string]interface{}
func ParseFilterParamsFromJSON(params map[string]interface{}) (*FilterParamsRaw, error) {
	dataMode := json.StringDefault(params, "default", "dataMode")
	dataModeParsed, err := DataModeFromString(dataMode)
	if err != nil {
		return nil, err
	}

	filterParams := &FilterParamsRaw{
		Size:     json.IntDefault(params, model.DefaultFilterSize, "size"),
		DataMode: dataModeParsed,
	}

	if params == nil {
		return filterParams, nil
	}

	highlights, ok := json.Array(params, "highlights", "list")
	if ok {
		for _, highlight := range highlights {
			h, err := parseFilter(highlight)
			if err != nil {
				return nil, err
			}
			filterParams.Highlights.List = append(filterParams.Highlights.List, h)
		}
	}
	invertHighlights, ok := json.Bool(params, "highlights", "invert")
	if ok {
		filterParams.Highlights.Invert = invertHighlights
	} else {
		return nil, errors.New("Missing required param highlights.Invert")
	}
	filters, ok := json.Array(params, "filters", "list")
	if ok {
		for _, filter := range filters {
			f, err := parseFilter(filter)
			if err != nil {
				return nil, err
			}
			filterParams.Filters.List = append(filterParams.Filters.List, f)
		}
	}

	invertFilters, ok := json.Bool(params, "filters", "invert")
	if ok {
		filterParams.Filters.Invert = invertFilters
	} else {
		return nil, errors.New("Missing required param filters.Invert")
	}

	variables, ok := json.StringArray(params, "variables")
	if ok {
		filterParams.Variables = variables
	}

	sort.SliceStable(filterParams.Highlights.List, func(i, j int) bool {
		return filterParams.Highlights.List[i].Key < filterParams.Highlights.List[j].Key
	})

	sort.SliceStable(filterParams.Filters.List, func(i, j int) bool {
		return filterParams.Filters.List[i].Key < filterParams.Filters.List[j].Key
	})

	return filterParams, nil
}

// NaNReplacement defines the type of replacement value to use for NaNs
type NaNReplacement int

const (
	// Null replaces NaN values with Nil, which will result in 'null' being encoded into the JSON structure
	Null NaNReplacement = iota + 1
	// EmptyString replaces NaN values with an empty string, which will result in "" being encoded into the JSON structure
	EmptyString
)

// ReplaceNaNs replaces NaN values found in numerical columns with empty values.  This allows
// for downstream JSON encoding, as the Go JSON encoder doesn't properly handle NaN values.
func ReplaceNaNs(data *FilteredData, replacementType NaNReplacement) *FilteredData {
	// go does not marshal NaN values properly so make them empty
	numericColumns := make([]int, 0)
	for _, c := range data.Columns {
		if model.IsNumerical(c.Type) {
			numericColumns = append(numericColumns, c.Index)
		}
	}

	if len(numericColumns) > 0 {
		for _, r := range data.Values {
			for _, nc := range numericColumns {
				f, ok := r[nc].Value.(float64)
				if ok && math.IsNaN(f) {
					if replacementType == Null {
						r[nc].Value = nil
					} else if replacementType == EmptyString {
						r[nc].Value = ""
					}
				}
			}
		}
	}

	return data
}
