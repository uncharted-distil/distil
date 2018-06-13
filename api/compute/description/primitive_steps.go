package description

import (
	"github.com/unchartedsoftware/distil/api/pipeline"
)

// NewSimonStep creates a SIMON data classification step.  It examines an input
// dataframe, and assigns types to the columns based on the exposed metadata.
func NewSimonStep() *StepData {
	return NewStepData(
		&pipeline.Primitive{
			Id:         "d2fa8df2-6517-3c26-bafc-87b701c4043a",
			Version:    "1.0.0",
			Name:       "simon",
			PythonPath: "d3m.primitives.distil.simon",
		},
		[]string{"produce"},
	)
}

// NewPunkStep creates a punk PCA-based feature ranking call that can be added to
// a pipeline.
// ** TODO: Placeholder.  Not yet in TA1 image.
func NewPunkStep() *StepData {
	return NewStepData(
		&pipeline.Primitive{
			Id:         "d2fa8df2-6517-3c26-bafc-87b701c4043a",
			Version:    "1.0.0",
			Name:       "punk",
			PythonPath: "d3m.primitives.distil.punk",
		},
		[]string{"produce"},
	)
}

// NewDukeStep creates a wrapper for the Duke dataset classifier.
func NewDukeStep() *StepData {
	return NewStepData(
		&pipeline.Primitive{
			Id:         "46612a42-6120-3559-9db9-3aa9a76eb94f",
			Version:    "1.0.0",
			Name:       "duke",
			PythonPath: "d3m.primitives.distil.duke",
		},
		[]string{"produce"},
	)
}

// NewCrocStep creates a wrapper for the Croc image classifier.
// **TODO: Not yet in TA1 image.
func NewCrocStep() *StepData {
	return NewStepData(
		&pipeline.Primitive{
			Id:         "46612a42-6120-3559-9db9-3aa9a76eb94f",
			Version:    "1.0.0",
			Name:       "croc",
			PythonPath: "d3m.primitives.distil.croc",
		},
		[]string{"produce"},
	)
}

// NewDatasetToDataframeStep creates a primitive call that transforms an input dataset
// into a PANDAS dataframe.
func NewDatasetToDataframeStep() *StepData {
	return &StepData{
		Primitive: &pipeline.Primitive{
			Id:         "4b42ce1e-9b98-4a25-b68e-fad13311eb65",
			Version:    "0.2.0",
			Name:       "Dataset to DataFrame converter",
			PythonPath: "d3m.primitives.datasets.DatasetToDataFrame",
		},
		Arguments:     map[string]string{},
		OutputMethods: []string{"produce"},
	}
}

// ColumnUpdate defines a column name and a semantic type to add/remove
// from that column.
type ColumnUpdate struct {
	Name         string
	SemanticType string
}

// NewUpdateSemanticTypeStep adds and removes semantic data values from an input
// dataset.
func NewUpdateSemanticTypeStep(resourceID string, add []*ColumnUpdate, remove []*ColumnUpdate) (*StepData, error) {
	// extract into two lists for compatibility with hyperparams interface
	addNames := []string{}
	addTypes := []string{}

	for _, val := range add {
		addNames = append(addNames, val.Name)
		addTypes = append(addTypes, val.SemanticType)
	}

	removeNames := []string{}
	removeTypes := []string{}
	for _, val := range remove {
		removeNames = append(removeNames, val.Name)
		removeTypes = append(removeTypes, val.SemanticType)
	}

	return NewStepDataWithHyperparameters(
		&pipeline.Primitive{
			Id:         "98c79128-555a-4a6b-85fb-d4f4064c94ab",
			Version:    "0.1.0",
			Name:       "Semantic type updater",
			PythonPath: "d3m.primitives.datasets.UpdateSemanticTypes",
		},
		[]string{"produce"},
		map[string]interface{}{
			"resource_id":    resourceID,
			"add_columns":    addNames,
			"add_types":      addTypes,
			"remove_columns": removeNames,
			"remove_types":   removeTypes,
		},
	), nil
}

// NewRemoveColumnsStep removes columns from an input dataframe.  Columns
// are specified by name and the match is case insensitive.
func NewRemoveColumnsStep(resourceID string, colNames []string) (*StepData, error) {
	return NewStepDataWithHyperparameters(
		&pipeline.Primitive{
			Id:         "2eeff053-395a-497d-88db-7374c27812e6",
			Version:    "0.1.0",
			Name:       "Column remover",
			PythonPath: "d3m.primitives.datasets.RemoveColumns",
		},
		[]string{"produce"},
		map[string]interface{}{
			"resource_id": resourceID,
			"columns":     colNames,
		},
	), nil
}
