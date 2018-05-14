package pipeline

import (
	"fmt"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
)

const (
	stepInputsKey     = "inputs"
	pipelineInputsKey = "inputs"
)

// DescriptionStep provides data for a pipeline description step and an operation
// to create a protobuf PipelineDescriptionStep from that data.
type DescriptionStep interface {
	BuildDescriptionStep() (*PipelineDescriptionStep, error)
	GetPrimitive() *Primitive
	GetArguments() map[string]string
	UpdateArguments(string, string)
	GetHyperparameters() map[string]interface{}
	GetOutputMethods() []string
}

// StepData contains the minimum amount of data used to describe a pipeline step
type StepData struct {
	Primitive       *Primitive
	Arguments       map[string]string
	Hyperparameters map[string]interface{}
	OutputMethods   []string
}

// NewStepData Creates a pipeline step instance from the required field subset.
func NewStepData(primitive *Primitive, outputMethods []string) *StepData {
	return NewStepDataWithHyperparameters(primitive, outputMethods, nil)
}

// NewStepDataWithHyperparameters creates a pipeline step instance from the required field subset.  Hyperparameters are
// optional so nil is a valid value, valid types fror hyper parameters are intXX, string, bool.
func NewStepDataWithHyperparameters(primitive *Primitive, outputMethods []string, hyperparameters map[string]interface{}) *StepData {
	return &StepData{
		Primitive:       primitive,
		Hyperparameters: hyperparameters, // optional, nil is valid
		Arguments:       map[string]string{},
		OutputMethods:   outputMethods,
	}
}

// GetPrimitive returns a primitive definition for a pipeline step.
func (s *StepData) GetPrimitive() *Primitive {
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
func (s *StepData) BuildDescriptionStep() (*PipelineDescriptionStep, error) {
	// generate arguments entries
	arguments := map[string]*PrimitiveStepArgument{}
	for k, v := range s.Arguments {
		arguments[k] = &PrimitiveStepArgument{
			// only handle container args rights now - extend to others if required
			Argument: &PrimitiveStepArgument_Container{
				Container: &ContainerArgument{
					Data: v,
				},
			},
		}
	}

	// generate arguments entries - accepted types are currently intXX, string, bool.  The underlying
	// protobuf structure allows for others - introducing them should be a matter of expanding this
	// list.
	hyperparameters := map[string]*PrimitiveStepHyperparameter{}
	for k, v := range s.Hyperparameters {
		var value *Value
		switch t := v.(type) {
		case int:
			value = &Value{
				Value: &Value_Int64{
					Int64: int64(t),
				},
			}
		case int8:
			value = &Value{
				Value: &Value_Int64{
					Int64: int64(t),
				},
			}
		case int16:
			value = &Value{
				Value: &Value_Int64{
					Int64: int64(t),
				},
			}
		case int32:
			value = &Value{
				Value: &Value_Int64{
					Int64: int64(t),
				},
			}
		case int64:
			value = &Value{
				Value: &Value_Int64{
					Int64: t,
				},
			}
		case bool:
			value = &Value{
				Value: &Value_Bool{
					Bool: t,
				},
			}
		case string:
			value = &Value{
				Value: &Value_String_{
					String_: t,
				},
			}
		default:
			return nil, errors.Errorf("compile failed: unhandled type `%v` for hyperparameter `%s`", v, k)
		}
		hyperparameters[k] = &PrimitiveStepHyperparameter{
			// only handle value args rights now - extend to others if required
			Argument: &PrimitiveStepHyperparameter_Value{
				Value: &ValueArgument{
					Data: value,
				},
			},
		}
	}

	// list of methods that will generate output - order matters because the steps are
	// numbered
	outputMethods := []*StepOutput{}
	for _, outputMethod := range s.OutputMethods {
		outputMethods = append(outputMethods,
			&StepOutput{
				Id: outputMethod,
			})
	}

	// create the pipeline description structure
	return &PipelineDescriptionStep{
		Step: &PipelineDescriptionStep_Primitive{
			Primitive: &PrimitivePipelineDescriptionStep{
				Primitive:   s.Primitive,
				Arguments:   arguments,
				Hyperparams: hyperparameters,
				Outputs:     outputMethods,
			},
		},
	}, nil
}

type InferenceStepData struct {
	Inputs  []string
	Outputs []string
}

func NewInferenceStepData() *InferenceStepData {
	return &InferenceStepData{
		Inputs:  []string{},
		Outputs: []string{"produce"},
	}
}

// GetPrimitive returns nil since there is no primitive associated with a placeholder
// step.
func (s *InferenceStepData) GetPrimitive() *Primitive {
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
func (s *InferenceStepData) BuildDescriptionStep() (*PipelineDescriptionStep, error) {
	// generate arguments entries
	inputs := []*StepInput{}
	for _, v := range s.Inputs {
		input := &StepInput{
			Data: v,
		}
		inputs = append(inputs, input)
	}

	// list of methods that will generate output - order matters because the steps are
	// numbered
	outputs := []*StepOutput{}
	for _, v := range s.Outputs {
		output := &StepOutput{
			Id: v,
		}
		outputs = append(outputs, output)
	}

	// create the pipeline description structure
	return &PipelineDescriptionStep{
		Step: &PipelineDescriptionStep_Placeholder{
			Placeholder: &PlaceholderPipelineDescriptionStep{
				Inputs:  inputs,
				Outputs: outputs,
			},
		},
	}, nil
}

type builder struct {
	name        string
	description string
	outputs     []string
	steps       []DescriptionStep
}

// DescriptionBuilder creates a PipelineDescription from a set of ordered pipeline description
// steps.  Called as
// 		pipelineDesc := NewBuilder("somePrimitive", "somePrimitive description").
//			Add(stepData0).
//			Add(stepData1).
// 			Compile()
type DescriptionBuilder interface {
	Add(stepData DescriptionStep) DescriptionBuilder
	AddInferencePoint() DescriptionBuilder
	Compile() (*PipelineDescription, error)
}

// NewDescriptionBuilder creates a new Builder instance.
func NewDescriptionBuilder(name, description string) DescriptionBuilder {
	return &builder{
		name:        name,
		description: description,
		outputs:     []string{},
		steps:       []DescriptionStep{},
	}
}

// Add a new step to the pipeline builder
func (p *builder) Add(step DescriptionStep) DescriptionBuilder {
	p.steps = append(p.steps, step)
	return p
}

func (p *builder) AddInferencePoint() DescriptionBuilder {
	// Create the standard inference step  and append it
	p.steps = append(p.steps, NewInferenceStepData())
	return p
}

func validateStep(steps []DescriptionStep, stepNumber int) error {
	// Validate step parameters.  This is currently pretty surface level, but we could
	// go in validate the struct hierarchy to catch more potential caller errors during
	// the compile step.
	//
	// NOTE: Hyperparameters and Primitive are optional so there is no included check at this time.

	step := steps[stepNumber]
	if step == nil {
		return errors.Errorf("compile failed: nil value for step %d", stepNumber)
	}

	args := step.GetArguments()
	if args == nil {
		return errors.Errorf("compile failed: step %d missing argument list", stepNumber)
	}

	outputs := step.GetOutputMethods()
	if len(outputs) == 0 {
		return errors.Errorf("compile failed: expected at least 1 output for step %d", stepNumber)
	}
	return nil
}

// Compile the pipeline into a PipelineDescription
func (p *builder) Compile() (*PipelineDescription, error) {
	if len(p.steps) == 0 {
		return nil, errors.New("compile failed: pipeline requires at least 1 step")
	}

	// make sure first step has an arg list
	err := validateStep(p.steps, 0)
	if err != nil {
		return nil, err
	}

	// first step, set the input to the dataset by default
	args := p.steps[0].GetArguments()
	_, ok := args[pipelineInputsKey]
	if ok {
		return nil, errors.Errorf("compile failed: argument `%s` is reserved for internal use", stepInputsKey)
	}
	p.steps[0].UpdateArguments(stepInputsKey, fmt.Sprintf("%s.0", pipelineInputsKey))

	// Connect the input of each step to the output of the previous.  Currently
	// only support a single output.
	for i := 1; i < len(p.steps); i++ {
		previousStep := i - 1
		previousOutput := p.steps[i-1].GetOutputMethods()[0]
		err := validateStep(p.steps, i)
		if err != nil {
			return nil, err
		}
		p.steps[i].UpdateArguments(stepInputsKey, fmt.Sprintf("steps.%d.%s", previousStep, previousOutput))
	}

	// Set the output from the tail end of the pipeline
	lastStep := len(p.steps) - 1
	lastOutput := p.steps[lastStep].GetOutputMethods()[0]
	pipelineOutputs := []*PipelineDescriptionOutput{
		{
			Data: fmt.Sprintf("steps.%d.%s", lastStep, lastOutput),
		},
	}

	// build the pipeline descriptions
	descriptionSteps := []*PipelineDescriptionStep{}
	for _, step := range p.steps {
		builtStep, err := step.BuildDescriptionStep()
		if err != nil {
			return nil, err
		}
		descriptionSteps = append(descriptionSteps, builtStep)
	}

	pipelineDesc := &PipelineDescription{
		Name:        p.name,
		Description: p.description,
		Steps:       descriptionSteps,
		Outputs:     pipelineOutputs,
	}

	return pipelineDesc, nil
}
