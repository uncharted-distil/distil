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
func NewSlothStep() *StepData {
	return NewStepDataWithHyperparameters(
		&pipeline.Primitive{
			Id:         "77bf4b92-2faa-3e38-bb7e-804131243a7f",
			Version:    "2.0.0",
			Name:       "Sloth",
			PythonPath: "d3m.primitives.distil.Sloth.cluster",
			Digest:     "f94f1aacc23792b680af0bd895f0fd2bac7336b29967b6ad766df4cb3c1933ab",
		},
		[]string{"produce"},
		map[string]interface{}{
			"nclusters": 4,
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

// NewTargetRankingStep creates a target ranking call that can be added to
// a pipeline.
func NewTargetRankingStep(target string) *StepData {
	return NewStepDataWithHyperparameters(
		&pipeline.Primitive{
			Id:         "04573880-d64f-4791-8932-52b7c3877639",
			Version:    "3.0.0",
			Name:       "PCA Features",
			PythonPath: "d3m.primitives.distil.pcafeatures",
			Digest:     "5302eebf2fb8a80e9f00e7b74888aba9eb448a9c0463d9d26786dab717a62c61",
		},
		[]string{"produce"},
		map[string]interface{}{
			"target": target,
		},
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

// NewDenormalizeStep denormalize data that is contained in multiple resource files.
func NewDenormalizeStep() *StepData {
	return NewStepData(
		&pipeline.Primitive{
			Id:         "f31f8c1f-d1c5-43e5-a4b2-2ae4a761ef2e",
			Version:    "0.2.0",
			Name:       "Denormalize datasets",
			PythonPath: "d3m.primitives.datasets.Denormalize",
			Digest:     "c39e3436373aed1944edbbc9b1cf24af5c71919d73bf0bb545cba0b685812df1",
		},
		[]string{"produce"},
	)
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

// NewTermFilterStep creates a primitive step that filters dataset rows based on a match against a
// term list.  The term match can be partial, or apply to whole terms only.
func NewTermFilterStep(resourceID string, colindex int, inclusive bool, terms []string, matchWhole bool) *StepData {
	return NewStepDataWithHyperparameters(
		&pipeline.Primitive{
			Id:         "622893c7-42fc-4561-a6f6-071fb85d610a",
			Version:    "0.1.0",
			Name:       "Term list dataset filter",
			PythonPath: "d3m.primitives.datasets.TermFilter",
			Digest:     "",
		},
		[]string{"produce"},
		map[string]interface{}{
			"resource_id": resourceID,
			"column":      colindex,
			"inclusive":   inclusive,
			"terms":       terms,
			"match_whole": matchWhole,
		},
	)
}

// NewRegexFilterStep creates a primitive step that filter dataset rows based on a regex match.
func NewRegexFilterStep(resourceID string, colindex int, inclusive bool, regex string) *StepData {
	return NewStepDataWithHyperparameters(
		&pipeline.Primitive{
			Id:         "d1b4c4b7-63ba-4ee6-ab30-035157cccf22",
			Version:    "0.1.0",
			Name:       "Regex dataset filter",
			PythonPath: "d3m.primitives.datasets.RegexFilter",
			Digest:     "",
		},
		[]string{"produce"},
		map[string]interface{}{
			"resource_id": resourceID,
			"column":      colindex,
			"inclusive":   inclusive,
			"regex":       regex,
		},
	)
}

// NewNumericRangeFilterStep creates a primitive step that filters dataset rows based on an
// included/excluded numeric range.  Inclusion of boundaries is controlled by the strict flag.
func NewNumericRangeFilterStep(resourceID string, colindex int, inclusive bool, min float64, max float64, strict bool) *StepData {
	return NewStepDataWithHyperparameters(
		&pipeline.Primitive{
			Id:         "8b1c1140-8c21-4f41-aeca-662b7d35aa29",
			Version:    "0.1.0",
			Name:       "Numeric range filter",
			PythonPath: "d3m.primitives.datasets.NumericRangeFilter",
			Digest:     "",
		},
		[]string{"produce"},
		map[string]interface{}{
			"resource_id": resourceID,
			"column":      colindex,
			"inclusive":   inclusive,
			"min":         min,
			"max":         max,
			"strict":      strict,
		},
	)
}

// NewTimeSeriesLoaderStep creates a primitive step that reads time series values using a dataframe
// containing a file URI column.  The file URIs are expected to point to CSV files, with the
// supplied time and value indices pointing the columns in the CSV that form the series data.
// The result is a new dataframe that stores the timetamps as the column headers,
// and the accompanying values for each file as a row.  Note that the file index column is negative,
// the primitive will use the first CSV file name column if finds.
func NewTimeSeriesLoaderStep(fileColIndex int, timeColIndex int, valueColIndex int) *StepData {
	// exclude the file col index val ue in the case of a negative index so that the
	// primitive will infer the colum
	args := map[string]interface{}{
		"time_col_index":  timeColIndex,
		"value_col_index": valueColIndex,
	}
	if fileColIndex >= 0 {
		args["file_col_index"] = fileColIndex
	}

	return NewStepDataWithHyperparameters(
		&pipeline.Primitive{
			Id:         "1689aafa-16dc-4c55-8ad4-76cadcf46086",
			Version:    "0.1.0",
			Name:       "Time series loader",
			PythonPath: "d3m.primitives.distil.TimeSeriesLoader",
			Digest:     "",
		},
		[]string{"produce"},
		args,
	)
}
