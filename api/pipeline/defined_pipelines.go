package pipeline

func CreateUserDatasetPipeline() (*PipelineDescription, error) {

	featureSelect, _ := NewRemoveColumnsStep([]string{"doubles, triples, player_name"})

	semanticTypeUpdate, _ := NewUpdateSemanticTypeStep(
		// adds
		[]*ColumnUpdate{
			&ColumnUpdate{
				Name:         "hall_of_fame",
				SemanticType: "https://metadata.datadrivendiscovery.org/types/CategoricalData",
			},
		},
		// removes
		[]*ColumnUpdate{
			&ColumnUpdate{
				Name:         "hall_of_fame",
				SemanticType: "http://schema.org/Integer",
			},
		},
	)

	// insantiate the pipeline
	pipeline, err := NewDescriptionBuilder("baseball_user_cleaning", "Baseball dataset with user driven feature selection and typing").
		Add(featureSelect).
		Add(semanticTypeUpdate).
		AddInferencePoint().
		Compile()

	if err != nil {
		return nil, err
	}
	return pipeline, nil
}
