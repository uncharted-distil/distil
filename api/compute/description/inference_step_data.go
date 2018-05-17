package description

import (
	"fmt"

	"github.com/unchartedsoftware/distil/api/pipeline"
	log "github.com/unchartedsoftware/plog"
)

// InferenceStepData provides data for a pipeline description placeholder step,
// which marks the point at which a TA2 should be begin pipeline inference.
type InferenceStepData struct {
	Inputs  []string
	Outputs []string
}

// NewInferenceStepData creates a InferenceStepData instance with default values.
func NewInferenceStepData() *InferenceStepData {
	return &InferenceStepData{
		Inputs:  []string{},
		Outputs: []string{"produce"},
	}
}

// GetPrimitive returns nil since there is no primitive associated with a placeholder
// step.
func (s *InferenceStepData) GetPrimitive() *pipeline.Primitive {
	return nil
}

// GetArguments adapts the internal placeholder step argument type to the primitive
// step argument type.
func (s *InferenceStepData) GetArguments() map[string]string {
	argMap := map[string]string{}
	for i, input := range s.Inputs {
		argMap[fmt.Sprintf("%s.%d", stepInputsKey, i)] = input
	}
	return argMap
}

// UpdateArguments updates the placheolder step argument.
func (s *InferenceStepData) UpdateArguments(key string, value string) {
	if key != stepInputsKey {
		log.Warnf("Compile warning - inference step key `%s` is not `%s` as expected", key, stepInputsKey)
	}
	s.Inputs = append(s.Inputs, value)
}

// GetHyperparameters returns an empty map since inference steps don't
// take hyper parameters.
func (s *InferenceStepData) GetHyperparameters() map[string]interface{} {
	return map[string]interface{}{}
}

// GetOutputMethods returns a list of methods that will be called to generate
// primitive output.  These feed into downstream primitives.
func (s *InferenceStepData) GetOutputMethods() []string {
	return s.Outputs
}

// BuildDescriptionStep creates protobuf structures from a pipeline step
// definition.
func (s *InferenceStepData) BuildDescriptionStep() (*pipeline.PipelineDescriptionStep, error) {
	// generate arguments entries
	inputs := []*pipeline.StepInput{}
	for _, v := range s.Inputs {
		input := &pipeline.StepInput{
			Data: v,
		}
		inputs = append(inputs, input)
	}

	// list of methods that will generate output - order matters because the steps are
	// numbered
	outputs := []*pipeline.StepOutput{}
	for _, v := range s.Outputs {
		output := &pipeline.StepOutput{
			Id: v,
		}
		outputs = append(outputs, output)
	}

	// create the pipeline description structure
	return &pipeline.PipelineDescriptionStep{
		Step: &pipeline.PipelineDescriptionStep_Placeholder{
			Placeholder: &pipeline.PlaceholderPipelineDescriptionStep{
				Inputs:  inputs,
				Outputs: outputs,
			},
		},
	}, nil
}
