package description

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-ingest/metadata"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/pipeline"
)

const defaultResource = "0"

// CreateUserDatasetPipeline creates a pipeline description to capture user feature selection and
// semantic type information.
func CreateUserDatasetPipeline(name string, description string, allFeatures []*model.Variable,
	targetFeature string, selectedFeatures []string, filters []*model.Filter) (*pipeline.PipelineDescription, error) {

	// save the selected features in a set for quick lookup
	selectedSet := map[string]bool{}
	for _, v := range selectedFeatures {
		selectedSet[strings.ToLower(v)] = true
	}
	columnIndices := mapColumns(allFeatures, selectedSet)

	// create the semantic type update primitive
	updateSemanticTypes, err := createUpdateSemanticTypes(allFeatures, selectedSet)
	if err != nil {
		return nil, err
	}

	// create the feature selection primitive
	removeFeatures, err := createRemoveFeatures(allFeatures, selectedSet)
	if err != nil {
		return nil, err
	}

	// If neither have any content, we'll skip the template altogether.
	if len(updateSemanticTypes) == 0 && removeFeatures == nil {
		return nil, nil
	}

	filterData := createFilterData(filters, columnIndices)

	// instantiate the pipeline
	builder := NewBuilder(name, description)
	for _, v := range updateSemanticTypes {
		builder = builder.Add(v)
	}
	if removeFeatures != nil {
		builder = builder.Add(removeFeatures)
	}
	for _, f := range filterData {
		builder = builder.Add(f)
	}

	pip, err := builder.AddInferencePoint().Compile()
	if err != nil {
		return nil, err
	}

	// Input set to arbitrary string for now
	pip.Inputs = []*pipeline.PipelineDescriptionInput{{
		Name: "dataset",
	}}
	return pip, nil
}

func createRemoveFeatures(allFeatures []*model.Variable, selectedSet map[string]bool) (*StepData, error) {
	// create a list of features to remove
	removeFeatures := []int{}
	for _, v := range allFeatures {
		if !selectedSet[strings.ToLower(v.Key)] {
			removeFeatures = append(removeFeatures, v.Index)
		}
	}

	if len(removeFeatures) == 0 {
		return nil, nil
	}

	// instantiate the feature remove primitive
	featureSelect, err := NewRemoveColumnsStep(defaultResource, removeFeatures)
	if err != nil {
		return nil, err
	}
	return featureSelect, nil
}

type update struct {
	removeIndices []int
	addIndices    []int
}

func newUpdate() *update {
	return &update{
		addIndices:    []int{},
		removeIndices: []int{},
	}
}

func createUpdateSemanticTypes(allFeatures []*model.Variable, selectedSet map[string]bool) ([]*StepData, error) {
	// create maps of (semantic type, index list) - primitive allows for semantic types to be added to /
	// remove from multiple columns in a single operation
	updateMap := map[string]*update{}
	for _, v := range allFeatures {
		if selectedSet[strings.ToLower(v.Key)] {
			addType := model.MapTA2Type(v.Type)
			if addType == "" {
				return nil, errors.Errorf("variable `%s` internal type `%s` can't be mapped to ta2", v.Key, v.Type)
			}
			removeType := model.MapTA2Type(v.OriginalType)
			if removeType == "" {
				return nil, errors.Errorf("remove variable `%s` internal type `%s` can't be mapped to ta2", v.Key, v.OriginalType)
			}

			// only apply change when types are different
			if addType != removeType {
				if _, ok := updateMap[addType]; !ok {
					updateMap[addType] = newUpdate()
				}
				updateMap[addType].addIndices = append(updateMap[addType].addIndices, v.Index)

				if _, ok := updateMap[removeType]; !ok {
					updateMap[removeType] = newUpdate()
				}
				updateMap[removeType].removeIndices = append(updateMap[removeType].removeIndices, v.Index)
			}
		}
	}

	// Copy the created maps into the column update structure used by the primitive
	semanticTypeUpdates := []*StepData{}
	for k, v := range updateMap {
		var addKey string
		if len(v.addIndices) > 0 {
			addKey = k
		}
		add := &ColumnUpdate{
			SemanticTypes: []string{addKey},
			Indices:       v.addIndices,
		}
		var removeKey string
		if len(v.removeIndices) > 0 {
			removeKey = k
		}
		remove := &ColumnUpdate{
			SemanticTypes: []string{removeKey},
			Indices:       v.removeIndices,
		}
		semanticTypeUpdate, err := NewUpdateSemanticTypeStep(defaultResource, add, remove)
		if err != nil {
			return nil, err
		}
		semanticTypeUpdates = append(semanticTypeUpdates, semanticTypeUpdate)
	}
	return semanticTypeUpdates, nil
}

func createFilterData(filters []*model.Filter, columnIndices map[string]int) []*StepData {

	// Map the fiters to pipeline primitives
	filterSteps := []*StepData{}
	for _, f := range filters {
		var filter *StepData
		inclusive := f.Mode == model.IncludeFilter
		colIndex := columnIndices[f.Key]

		switch f.Type {
		case model.NumericalFilter:
			filter = NewNumericRangeFilterStep(defaultResource, colIndex, inclusive, *f.Min, *f.Max, false)
		case model.CategoricalFilter:
			filter = NewTermFilterStep(defaultResource, colIndex, inclusive, f.Categories, true)
		case model.RowFilter:
			filter = NewTermFilterStep(defaultResource, colIndex, inclusive, f.D3mIndices, true)
		case model.FeatureFilter, model.TextFilter:
			filter = NewTermFilterStep(defaultResource, colIndex, inclusive, f.Categories, false)
		}

		filterSteps = append(filterSteps, filter)
	}
	return filterSteps
}

// CreateSlothPipeline creates a pipeline to peform timeseries clustering on a dataset.
func CreateSlothPipeline(name string, description string, timeColumn string, valueColumn string,
	timeSeriesFeatures []*metadata.Variable) (*pipeline.PipelineDescription, error) {

	timeIdx, err := getIndex(timeSeriesFeatures, timeColumn)
	if err != nil {
		return nil, err
	}

	valueIdx, err := getIndex(timeSeriesFeatures, valueColumn)
	if err != nil {
		return nil, err
	}

	// insantiate the pipeline
	pipeline, err := NewBuilder(name, description).
		Add(NewDenormalizeStep()).
		Add(NewDatasetToDataframeStep()).
		Add(NewTimeSeriesLoaderStep(-1, timeIdx, valueIdx)).
		Add(NewSlothStep()).
		Compile()

	if err != nil {
		return nil, err
	}

	return pipeline, nil
}

// CreateDukePipeline creates a pipeline to peform image featurization on a dataset.
func CreateDukePipeline(name string, description string) (*pipeline.PipelineDescription, error) {
	// insantiate the pipeline
	pipeline, err := NewBuilder(name, description).
		Add(NewDatasetToDataframeStep()).
		Add(NewDukeStep()).
		Compile()

	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

// CreateSimonPipeline creates a pipeline to run semantic type inference on a dataset's
// columns.
func CreateSimonPipeline(name string, description string) (*pipeline.PipelineDescription, error) {
	// insantiate the pipeline
	pipeline, err := NewBuilder(name, description).
		Add(NewDatasetToDataframeStep()).
		Add(NewSimonStep()).
		Compile()

	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

// CreateCrocPipeline creates a pipeline to run image featurization on a dataset.
func CreateCrocPipeline(name string, description string, targetColumns []string, outputLabels []string) (*pipeline.PipelineDescription, error) {
	// insantiate the pipeline
	pipeline, err := NewBuilder(name, description).
		Add(NewDenormalizeStep()).
		Add(NewDatasetToDataframeStep()).
		Add(NewCrocStep(targetColumns, outputLabels)).
		Compile()

	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

// CreateUnicornPipeline creates a pipeline to run image clustering on a dataset.
func CreateUnicornPipeline(name string, description string, targetColumns []string, outputLabels []string) (*pipeline.PipelineDescription, error) {
	// insantiate the pipeline
	pipeline, err := NewBuilder(name, description).
		Add(NewDenormalizeStep()).
		Add(NewDatasetToDataframeStep()).
		Add(NewUnicornStep(targetColumns, outputLabels)).
		Compile()

	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

// CreatePCAFeaturesPipeline creates a pipeline to run feature ranking on an input dataset.
func CreatePCAFeaturesPipeline(name string, description string) (*pipeline.PipelineDescription, error) {
	// insantiate the pipeline
	pipeline, err := NewBuilder(name, description).
		Add(NewDatasetToDataframeStep()).
		Add(NewPCAFeaturesStep()).
		Compile()

	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

// CreateDenormalizePipeline creates a pipeline to run the denormalize primitive on an input dataset.
func CreateDenormalizePipeline(name string, description string) (*pipeline.PipelineDescription, error) {
	// insantiate the pipeline
	pipeline, err := NewBuilder(name, description).
		Add(NewDenormalizeStep()).
		Add(NewDatasetToDataframeStep()).
		Compile()

	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

// CreateTargetRankingPipeline creates a pipeline to run feature ranking on an input dataset.
func CreateTargetRankingPipeline(name string, description string, target string) (*pipeline.PipelineDescription, error) {
	// insantiate the pipeline
	pipeline, err := NewBuilder(name, description).
		Add(NewTargetRankingStep(target)).
		Compile()

	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

func mapColumns(allFeatures []*model.Variable, selectedSet map[string]bool) map[string]int {
	colIndices := make(map[string]int)
	index := 0
	for _, f := range allFeatures {
		if selectedSet[f.Key] {
			colIndices[f.Key] = index
			index = index + 1
		}
	}

	return colIndices
}

func getIndex(allFeatures []*metadata.Variable, name string) (int, error) {
	for _, f := range allFeatures {
		if strings.EqualFold(name, f.Name) {
			return f.Index, nil
		}
	}
	return -1, errors.Errorf("can't find var '%s'", name)
}
