package pipeline

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	stepInputsKey     = "inputs"
	pipelineInputsKey = "inputs"
)

type DescriptionStep interface {
	GetPrimitive() *Primitive
	GetArguments() map[string]string
	GetOutputMethods() []string
	BuildDescriptionStep() *PipelineDescriptionStep
}

// StepData contains the minimum amount of data used to describe a pipeline step
type StepData struct {
	Primitive     *Primitive
	Arguments     map[string]string
	OutputMethods []string
}

// GetPrimitive returns a primitive definition for a pipeline step.
func (s *StepData) GetPrimitive() *Primitive {
	return s.Primitive
}

// GetArguments returns a map of arguments that will be passed to the methods
// of the primitive step.
func (s *StepData) GetArguments() map[string]string {
	return s.Arguments
}

// GetOutputMethods returns a list of methods that will be called to generate
// primitive output.  These feed into downstream primitives.
func (s *StepData) GetOutputMethods() []string {
	return s.OutputMethods
}

// BuildDescriptionStep creates protobuf structures from a pipeline step
// definition.
func (s *StepData) BuildDescriptionStep() *PipelineDescriptionStep {
	// generate arguments entries
	psArgs := map[string]*PrimitiveStepArgument{}
	for k, v := range s.Arguments {
		psArgs[k] = &PrimitiveStepArgument{
			// only handle container args rights now - extend to others if required
			Argument: &PrimitiveStepArgument_Container{
				Container: &ContainerArgument{
					Data: v,
				},
			},
		}
	}

	// list of methods that will generate output - order matters because
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
				Primitive: s.Primitive,
				Arguments: psArgs,
				Outputs:   outputMethods,
			},
		},
	}
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

func validateStep(steps []DescriptionStep, stepNumber int) (map[string]string, error) {
	step := steps[stepNumber]
	if step == nil {
		return nil, errors.Errorf("compile failed: nil value for step %d", stepNumber)
	}

	primitive := step.GetPrimitive()
	if primitive == nil {
		return nil, errors.Errorf("compile failed: step %d missing primitive definition", stepNumber)
	}

	args := step.GetArguments()
	if args == nil {
		return nil, errors.Errorf("compile failed: step %d missing argument list", stepNumber)
	}

	outputs := step.GetOutputMethods()
	if len(outputs) == 0 {
		return nil, errors.Errorf("compile failed: expected at least 1 output for step %d", stepNumber)
	}
	return args, nil
}

// Compile the pipeline into a PipelineDescription
func (p *builder) Compile() (*PipelineDescription, error) {
	if len(p.steps) == 0 {
		return nil, errors.New("compile failed: pipeline requires at least 1 step")
	}

	// make sure first step has an arg list
	args, err := validateStep(p.steps, 0)
	if err != nil {
		return nil, err
	}

	// first step, set the input to the dataset by default
	_, ok := args[pipelineInputsKey]
	if ok {
		return nil, errors.Errorf("compile failed: argument `%s` is reserved for internal use", stepInputsKey)
	}
	args[stepInputsKey] = fmt.Sprintf("%s.0", pipelineInputsKey)

	// Connect the input of each step to the output of the previous.  Currently
	// only support a single output.
	for i := 1; i < len(p.steps); i++ {
		previousStep := i - 1
		previousOutput := p.steps[i-1].GetOutputMethods()[0]
		args, err := validateStep(p.steps, i)
		if err != nil {
			return nil, err
		}
		args[stepInputsKey] = fmt.Sprintf("steps.%d.%s", previousStep, previousOutput)
	}

	// Set the output from the tail end of the pipeline
	lastStep := len(p.steps) - 1
	lastOutput := p.steps[lastStep].GetOutputMethods()[0]
	pipelineOutputs := []*PipelineDescriptionOutput{
		&PipelineDescriptionOutput{
			Data: fmt.Sprintf("steps.%d.%s", lastStep, lastOutput),
		},
	}

	// build the pipeline descriptions
	descriptionSteps := []*PipelineDescriptionStep{}
	for _, step := range p.steps {
		descriptionSteps = append(descriptionSteps, step.BuildDescriptionStep())
	}

	pipelineDesc := &PipelineDescription{
		Name:        p.name,
		Description: p.description,
		Steps:       descriptionSteps,
		Outputs:     pipelineOutputs,
	}

	return pipelineDesc, nil
}
