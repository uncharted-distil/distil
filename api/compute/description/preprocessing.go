package description

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/pipeline"
)

const defaultResource = "0"

// CreateUserDatasetPipeline creates a pipeline description to capture user feature selection and
// semantic type information.
func CreateUserDatasetPipeline(name string, description string, allFeatures []*model.Variable,
	selectedFeatures []string, target string) (*pipeline.PipelineDescription, error) {

	// save the selected features in a set for quick lookup
	selectedSet := map[string]bool{}
	for _, v := range selectedFeatures {
		selectedSet[strings.ToLower(v)] = true
	}

	// create a list of features to remove
	removeFeatures := []string{}
	for _, v := range allFeatures {
		if !selectedSet[strings.ToLower(v.Key)] {
			removeFeatures = append(removeFeatures, v.Key)
		}
	}

	// create the added/removed semantic types
	addedTypes := []*ColumnUpdate{}
	removedTypes := []*ColumnUpdate{}
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
				addedTypes = append(addedTypes, &ColumnUpdate{
					Name:         v.Key,
					SemanticType: addType,
				})

				removedTypes = append(removedTypes, &ColumnUpdate{
					Name:         v.Key,
					SemanticType: removeType,
				})
			}
		}

		if strings.EqualFold(v.Key, target) {
			// Add the target role type to the target variable.  TA2 systems can key off of the
			// problem description or the presence of this semantic type when searching solutions.
			addedTypes = append(addedTypes, &ColumnUpdate{
				Name:         v.Key,
				SemanticType: model.TA2TargetType,
			})
		}
	}

	featureSelect, err := NewRemoveColumnsStep(defaultResource, removeFeatures)
	if err != nil {
		return nil, err
	}

	semanticTypeUpdate, _ := NewUpdateSemanticTypeStep(defaultResource, addedTypes, removedTypes)
	if err != nil {
		return nil, err
	}

	// insantiate the pipeline
	pipeline, err := NewBuilder(name, description).
		Add(featureSelect).
		Add(semanticTypeUpdate).
		AddInferencePoint().
		Compile()

	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

// CreateSlothPipeline creates a pipeline to peform timeseries clustering on a dataset.
func CreateSlothPipeline(name string, description string) (*pipeline.PipelineDescription, error) {
	// insantiate the pipeline
	pipeline, err := NewBuilder(name, description).
		Add(NewDatasetToDataframeStep()).
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
		Add(NewDatasetToDataframeStep()).
		Add(NewCrocStep(targetColumns, outputLabels)).
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
