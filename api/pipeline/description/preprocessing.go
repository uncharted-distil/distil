package description

import (
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
		selectedSet[v] = true
	}

	// create a list of features to remove and a list of semantic type updaates
	removeFeatures := []string{}
	addedTypes := []*ColumnUpdate{}
	removedTypes := []*ColumnUpdate{}
	for _, v := range allFeatures {
		if !selectedSet[v.Name] {
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

			addedTypes = append(addedTypes, &ColumnUpdate{
				Name:         v.Name,
				SemanticType: v.Type,
			})

			removedTypes = append(removedTypes, &ColumnUpdate{
				Name:         v.Name,
				SemanticType: v.OriginalType,
			})
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
