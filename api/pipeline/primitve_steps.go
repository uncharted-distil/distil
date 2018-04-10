package pipeline

// NewClassificationStep creates a SIMON classification call that can be added to
// a pipeline.
func NewClassificationStep() *StepData {
	return &StepData{
		Primitive: &Primitive{
			Id:         "d2fa8df2-6517-3c26-bafc-87b701c4043a",
			Version:    "1.0.0",
			Name:       "simon",
			PythonPath: "d3m.primitives.distil.simon",
		},
		Arguments:     map[string]string{},
		OutputMethods: []string{"produce"},
	}
}

// NewRankingStep creates a punk PCA-based feature ranking call that can be added to
// a pipeline.
// ** TODO: Placeholder.  Not yet in TA1 image.
func NewRankingStep() *StepData {
	return &StepData{
		Primitive: &Primitive{
			Id:         "d2fa8df2-6517-3c26-bafc-87b701c4043a",
			Version:    "1.0.0",
			Name:       "punk",
			PythonPath: "d3m.primitives.distil.punk",
		},
		Arguments:     map[string]string{},
		OutputMethods: []string{"produce"},
	}
}

// NewDatasetToDataframeStep creates a primitive call that transforms an input dataset
// into a PANDAS dataframe.
func NewDatasetToDataframeStep() *StepData {
	return &StepData{
		Primitive: &Primitive{
			Id:         "c1cf1981-7257-497d-abd6-1d46275d308e",
			Version:    "1.0.0",
			Name:       "",
			PythonPath: "d3m.primitives.datasets.DatasetToDataframePrimitive",
		},
		Arguments:     map[string]string{},
		OutputMethods: []string{"produce"},
	}
}
