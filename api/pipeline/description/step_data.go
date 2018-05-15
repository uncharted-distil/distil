package description

import (
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/pipeline"
)

const (
	stepInputsKey     = "inputs"
	pipelineInputsKey = "inputs"
)

// Step provides data for a pipeline description step and an operation
// to create a protobuf PipelineDescriptionStep from that data.
type Step interface {
	BuildDescriptionStep() (*pipeline.PipelineDescriptionStep, error)
	GetPrimitive() *pipeline.Primitive
	GetArguments() map[string]string
	UpdateArguments(string, string)
	GetHyperparameters() map[string]interface{}
	GetOutputMethods() []string
}

// StepData contains the minimum amount of data used to describe a pipeline step
type StepData struct {
	Primitive       *pipeline.Primitive
	Arguments       map[string]string
	Hyperparameters map[string]interface{}
	OutputMethods   []string
}

// NewStepData Creates a pipeline step instance from the required field subset.
func NewStepData(primitive *pipeline.Primitive, outputMethods []string) *StepData {
	return NewStepDataWithHyperparameters(primitive, outputMethods, nil)
}

// NewStepDataWithHyperparameters creates a pipeline step instance from the required field subset.  Hyperparameters are
// optional so nil is a valid value, valid types fror hyper parameters are intXX, string, bool.
func NewStepDataWithHyperparameters(primitive *pipeline.Primitive, outputMethods []string, hyperparameters map[string]interface{}) *StepData {
	return &StepData{
		Primitive:       primitive,
		Hyperparameters: hyperparameters, // optional, nil is valid
		Arguments:       map[string]string{},
		OutputMethods:   outputMethods,
	}
}

// GetPrimitive returns a primitive definition for a pipeline step.
func (s *StepData) GetPrimitive() *pipeline.Primitive {
	return s.Primitive
}

// GetArguments returns a map of arguments that will be passed to the methods
// of the primitive step.
func (s *StepData) GetArguments() map[string]string {
	copy := map[string]string{}
	for k, v := range s.Arguments {
		copy[k] = v
	}
	return copy
}

// UpdateArguments updates the arguments map that will be passed to the methods
// of primtive step.
func (s *StepData) UpdateArguments(key string, value string) {
	s.Arguments[key] = value
}

// GetHyperparameters returns a map of arguments that will be passed to the primitive methods
// of the primitive step.  Types are currently restricted to intXX, bool, string
func (s *StepData) GetHyperparameters() map[string]interface{} {
	return s.Hyperparameters
}

// GetOutputMethods returns a list of methods that will be called to generate
// primitive output.  These feed into downstream primitives.
func (s *StepData) GetOutputMethods() []string {
	return s.OutputMethods
}

// BuildDescriptionStep creates protobuf structures from a pipeline step
// definition.
func (s *StepData) BuildDescriptionStep() (*pipeline.PipelineDescriptionStep, error) {
	// generate arguments entries
	arguments := map[string]*pipeline.PrimitiveStepArgument{}
	for k, v := range s.Arguments {
		arguments[k] = &pipeline.PrimitiveStepArgument{
			// only handle container args rights now - extend to others if required
			Argument: &pipeline.PrimitiveStepArgument_Container{
				Container: &pipeline.ContainerArgument{
					Data: v,
				},
			},
		}
	}

	// generate arguments entries - accepted types are currently intXX, string, bool.  The underlying
	// protobuf structure allows for others - introducing them should be a matter of expanding this
	// list.
	hyperparameters := map[string]*pipeline.PrimitiveStepHyperparameter{}
	for k, v := range s.Hyperparameters {
		var value *pipeline.Value
		switch t := v.(type) {
		case int:
			value = &pipeline.Value{
				Value: &pipeline.Value_Int64{
					Int64: int64(t),
				},
			}
		case int8:
			value = &pipeline.Value{
				Value: &pipeline.Value_Int64{
					Int64: int64(t),
				},
			}
		case int16:
			value = &pipeline.Value{
				Value: &pipeline.Value_Int64{
					Int64: int64(t),
				},
			}
		case int32:
			value = &pipeline.Value{
				Value: &pipeline.Value_Int64{
					Int64: int64(t),
				},
			}
		case int64:
			value = &pipeline.Value{
				Value: &pipeline.Value_Int64{
					Int64: t,
				},
			}
		case bool:
			value = &pipeline.Value{
				Value: &pipeline.Value_Bool{
					Bool: t,
				},
			}
		case string:
			value = &pipeline.Value{
				Value: &pipeline.Value_String_{
					String_: t,
				},
			}
		default:
			return nil, errors.Errorf("compile failed: unhandled type `%v` for hyperparameter `%s`", v, k)
		}
		hyperparameters[k] = &pipeline.PrimitiveStepHyperparameter{
			// only handle value args rights now - extend to others if required
			Argument: &pipeline.PrimitiveStepHyperparameter_Value{
				Value: &pipeline.ValueArgument{
					Data: value,
				},
			},
		}
	}

	// list of methods that will generate output - order matters because the steps are
	// numbered
	outputMethods := []*pipeline.StepOutput{}
	for _, outputMethod := range s.OutputMethods {
		outputMethods = append(outputMethods,
			&pipeline.StepOutput{
				Id: outputMethod,
			})
	}

	// create the pipeline description structure
	return &pipeline.PipelineDescriptionStep{
		Step: &pipeline.PipelineDescriptionStep_Primitive{
			Primitive: &pipeline.PrimitivePipelineDescriptionStep{
				Primitive:   s.Primitive,
				Arguments:   arguments,
				Hyperparams: hyperparameters,
				Outputs:     outputMethods,
			},
		},
	}, nil
}
