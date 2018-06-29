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
			Version:    "1.1.1",
			Name:       "simon",
			PythonPath: "d3m.primitives.distil.simon",
		},
		[]string{"produce"},
	)
}

// NewSlothStep creates a Sloth timeseries clustering step.
func NewSlothStep() *StepData {
	return NewStepData(
		&pipeline.Primitive{
			Id:         "77bf4b92-2faa-3e38-bb7e-804131243a7f",
			Version:    "1.0.0",
			Name:       "Sloth",
			PythonPath: "d3m.primitives.distil.Sloth.cluster",
		},
		[]string{"produce"},
	)
}

// NewPCAFeaturesStep creates a PCA-based feature ranking call that can be added to
// a pipeline.
func NewPCAFeaturesStep() *StepData {
	return NewStepData(
		&pipeline.Primitive{
			Id:         "04573880-d64f-4791-8932-52b7c3877639",
			Version:    "3.0.0",
			Name:       "PCA Features",
			PythonPath: "d3m.primitives.distil.pcafeatures",
		},
		[]string{"produce"},
	)
}

// NewDukeStep creates a wrapper for the Duke dataset classifier.
func NewDukeStep() *StepData {
	return NewStepData(
		&pipeline.Primitive{
			Id:         "46612a42-6120-3559-9db9-3aa9a76eb94f",
			Version:    "1.1.1",
			Name:       "duke",
			PythonPath: "d3m.primitives.distil.duke",
		},
		[]string{"produce"},
	)
}

// NewCrocStep creates a wrapper for the Croc image classifier.
func NewCrocStep(targetColumns []string, outputLabels []string) *StepData {
	return NewStepDataWithHyperparameters(
		&pipeline.Primitive{
			Id:         "404fae2a-2f0a-4c9b-9ad2-fb1528990561",
			Version:    "1.2.2",
			Name:       "croc",
			PythonPath: "d3m.primitives.distil.croc",
		},
		[]string{"produce"},
		map[string]interface{}{
			"target_columns": targetColumns,
			"output_labels":  outputLabels,
		},
	)
}

// NewDatasetToDataframeStep creates a primitive call that transforms an input dataset
// into a PANDAS dataframe.
func NewDatasetToDataframeStep() *StepData {
	return NewStepData(
		&pipeline.Primitive{
			Id:         "4b42ce1e-9b98-4a25-b68e-fad13311eb65",
			Version:    "0.2.0",
			Name:       "Dataset to DataFrame converter",
			PythonPath: "d3m.primitives.datasets.DatasetToDataFrame",
		},
		[]string{"produce"},
	)
}

// ColumnUpdate defines a set of column indices to add/remvoe
// a set of semantic types to/from.
type ColumnUpdate struct {
	Indices       []int
	SemanticTypes []string
}

// NewUpdateSemanticTypeStep adds and removes semantic data values from an input
// dataset.  An add of (1, 2), ("type a", "type b") would result in "type a" and "type b"
// being added to index 1 and 2.
func NewUpdateSemanticTypeStep(resourceID string, add *ColumnUpdate, remove *ColumnUpdate) (*StepData, error) {
	return NewStepDataWithHyperparameters(
		&pipeline.Primitive{
			Id:         "98c79128-555a-4a6b-85fb-d4f4064c94ab",
			Version:    "0.2.0",
			Name:       "Semantic type updater",
			PythonPath: "d3m.primitives.data.UpdateSemanticTypes",
		},
		[]string{"produce"},
		map[string]interface{}{
			"resource_id":    resourceID,
			"add_indices":    add.Indices,
			"add_types":      add.SemanticTypes,
			"remove_indices": remove.Indices,
			"remove_types":   remove.SemanticTypes,
		},
	), nil
}

// NewRemoveColumnsStep removes columns from an input dataframe.  Columns
// are specified by name and the match is case insensitive.
func NewRemoveColumnsStep(resourceID string, colIndices []int) (*StepData, error) {
	return NewStepDataWithHyperparameters(
		&pipeline.Primitive{
			Id:         "2eeff053-395a-497d-88db-7374c27812e6",
			Version:    "0.2.0",
			Name:       "Column remover",
			PythonPath: "d3m.primitives.data.RemoveColumns",
		},
		[]string{"produce"},
		map[string]interface{}{
			"resource_id": resourceID,
			"columns":     colIndices,
		},
	), nil
}
