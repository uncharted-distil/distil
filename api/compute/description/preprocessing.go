package description

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/pipeline"
)

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
		if !selectedSet[strings.ToLower(v.Name)] {
			removeFeatures = append(removeFeatures, v.Name)
		}
	}

	// create the added/removed semantic types
	addedTypes := []*ColumnUpdate{}
	removedTypes := []*ColumnUpdate{}
	for _, v := range allFeatures {
		if selectedSet[strings.ToLower(v.Name)] {
			addType := model.MapTA2Type(v.Type)
			if addType == "" {
				return nil, errors.Errorf("variable `%s` internal type `%s` can't be mapped to ta2", v.Name, v.Type)
			}
			removeType := model.MapTA2Type(v.OriginalType)
			if removeType == "" {
				return nil, errors.Errorf("remove variable `%s` internal type `%s` can't be mapped to ta2", v.Name, v.OriginalType)
			}

			// only apply change when types are different
			if addType != removeType {
				addedTypes = append(addedTypes, &ColumnUpdate{
					Name:         v.Name,
					SemanticType: addType,
				})

				removedTypes = append(removedTypes, &ColumnUpdate{
					Name:         v.Name,
					SemanticType: removeType,
				})
			}
		}

		if strings.EqualFold(v.Name, target) {
			// Add the target role type to the target variable.  TA2 systems can key off of the
			// problem description or the presence of this semantic type when searching solutions.
			addedTypes = append(addedTypes, &ColumnUpdate{
				Name:         v.Name,
				SemanticType: model.TA2TargetType,
			})
		}
	}

	featureSelect, err := NewRemoveColumnsStep(removeFeatures)
	if err != nil {
		return nil, err
	}

	semanticTypeUpdate, _ := NewUpdateSemanticTypeStep(addedTypes, removedTypes)
	if err != nil {
		return nil, err
	}

	// insantiate the pipeline
	pipeline, err := NewBuilder(name, description).
		Add(NewDatasetToDataframeStep()).
		Add(featureSelect).
		Add(semanticTypeUpdate).
		AddInferencePoint().
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
func CreateCrocPipeline(name string, description string) (*pipeline.PipelineDescription, error) {
	// insantiate the pipeline
	pipeline, err := NewBuilder(name, description).
		Add(NewDatasetToDataframeStep()).
		Add(NewCrocStep()).
		Compile()

	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

// CreatePunkPipeline creates a pipeline to run feature ranking on an input dataset.
func CreatePunkPipeline(name string, description string) (*pipeline.PipelineDescription, error) {
	// insantiate the pipeline
	pipeline, err := NewBuilder(name, description).
		Add(NewDatasetToDataframeStep()).
		Add(NewPunkStep()).
		Compile()

	if err != nil {
		return nil, err
	}
	return pipeline, nil
}
