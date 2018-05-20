package description

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/pipeline"
)

// CreateUserDatasetPipeline creates a pipeline description to capture user feature selection and
// semantic type information.
func CreateUserDatasetPipeline(name string, description string,
	allFeatures []*model.Variable, selectedFeatures []string) (*pipeline.PipelineDescription, error) {

	// save the selected features in a set for quick lookup
	selectedSet := map[string]bool{}
	for _, v := range selectedFeatures {
		selectedSet[strings.ToLower(v)] = true
	}

	// create a list of features to remove and a list of semantic type updaates
	removeFeatures := []string{}
	addedTypes := []*ColumnUpdate{}
	removedTypes := []*ColumnUpdate{}
	for _, v := range allFeatures {
		if !selectedSet[strings.ToLower(v.Name)] {
			removeFeatures = append(removeFeatures, v.Name)
		} else {
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
					SemanticType: removeType,
				})

				removedTypes = append(removedTypes, &ColumnUpdate{
					Name:         v.Name,
					SemanticType: addType,
				})
			}

		}
	}

	featureSelect, _ := NewRemoveColumnsStep(removeFeatures)
	semanticTypeUpdate, _ := NewUpdateSemanticTypeStep(addedTypes, removedTypes)

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
