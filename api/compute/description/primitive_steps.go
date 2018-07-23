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
			Digest:     "0673d166f157944d3b6fdfa451f31fdfdbead7315ede3d6d9edb20f3f220b836",
		},
		[]string{"produce"},
	)
}

// NewSlothStep creates a Sloth timeseries clustering step.
func NewSlothStep(targetColumns []string, outputLabels []string) *StepData {
	return NewStepDataWithHyperparameters(
		&pipeline.Primitive{
			Id:         "77bf4b92-2faa-3e38-bb7e-804131243a7f",
			Version:    "1.0.0",
			Name:       "Sloth",
			PythonPath: "d3m.primitives.distil.Sloth.cluster",
			Digest:     "f94f1aacc23792b680af0bd895f0fd2bac7336b29967b6ad766df4cb3c1933ab",
		},
		[]string{"produce"},
		map[string]interface{}{
			"target_columns": targetColumns,
			"output_labels":  outputLabels,
		},
	)
}

// NewUnicornStep creates a unicorn image clustering step.
func NewUnicornStep(targetColumns []string, outputLabels []string) *StepData {
	return NewStepDataWithHyperparameters(
		&pipeline.Primitive{
			Id:         "475c26dc-eb2e-43d3-acdb-159b80d9f099",
			Version:    "1.0.0",
			Name:       "unicorn",
			PythonPath: "d3m.primitives.distil.unicorn",
			Digest:     "2b0c0784fc077b106a9547a197be92ab02298dc206d60610929c50f831e86e84",
		},
		[]string{"produce"},
		map[string]interface{}{
			"target_columns": targetColumns,
			"output_labels":  outputLabels,
		},
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
			Digest:     "5302eebf2fb8a80e9f00e7b74888aba9eb448a9c0463d9d26786dab717a62c61",
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
			Digest:     "ea522d2adc756c3ad76f5848d28cd396304d4dfdc0cc55aa8b90fbaf04e8fc30",
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
			Digest:     "09cd99d609e317559feff580b8d893d0188f12915ab8d84a98de34eb344e340c",
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
			Version:    "0.3.0",
			Name:       "Dataset to DataFrame converter",
			PythonPath: "d3m.primitives.datasets.DatasetToDataFrame",
			Digest:     "85b946aa6123354fe51a288c3be56aaca82e76d4071c1edc13be6f9e0e100144",
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
			PythonPath: "d3m.primitives.datasets.UpdateSemanticTypes",
			Digest:     "85b946aa6123354fe51a288c3be56aaca82e76d4071c1edc13be6f9e0e100144",
		},
		[]string{"produce"},
		map[string]interface{}{
			"resource_id":    resourceID,
			"add_columns":    add.Indices,
			"add_types":      add.SemanticTypes,
			"remove_columns": remove.Indices,
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
			PythonPath: "d3m.primitives.datasets.RemoveColumns",
			Digest:     "85b946aa6123354fe51a288c3be56aaca82e76d4071c1edc13be6f9e0e100144",
		},
		[]string{"produce"},
		map[string]interface{}{
			"resource_id": resourceID,
			"columns":     colIndices,
		},
	), nil
}
