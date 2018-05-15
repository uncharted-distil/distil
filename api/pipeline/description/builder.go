package description

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/pipeline"
)

type builder struct {
	name        string
	description string
	outputs     []string
	steps       []Step
}

// Builder creates a PipelineDescription from a set of ordered pipeline description
// steps.  Called as
// 		pipelineDesc := NewBuilder("somePrimitive", "somePrimitive description").
//			Add(stepData0).
//			Add(stepData1).
// 			Compile()
type Builder interface {
	Add(stepData Step) Builder
	AddInferencePoint() Builder
	Compile() (*pipeline.PipelineDescription, error)
}

// NewBuilder creates a new Builder instance.
func NewBuilder(name, description string) Builder {
	return &builder{
		name:        name,
		description: description,
		outputs:     []string{},
		steps:       []Step{},
	}
}

// Add a new step to the pipeline builder
func (p *builder) Add(step Step) Builder {
	p.steps = append(p.steps, step)
	return p
}

func (p *builder) AddInferencePoint() Builder {
	// Create the standard inference step  and append it
	p.steps = append(p.steps, NewInferenceStepData())
	return p
}

func validateStep(steps []Step, stepNumber int) error {
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
func (p *builder) Compile() (*pipeline.PipelineDescription, error) {
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
	pipelineOutputs := []*pipeline.PipelineDescriptionOutput{
		{
			Data: fmt.Sprintf("steps.%d.%s", lastStep, lastOutput),
		},
	}

	// build the pipeline descriptions
	descriptionSteps := []*pipeline.PipelineDescriptionStep{}
	for _, step := range p.steps {
		builtStep, err := step.BuildDescriptionStep()
		if err != nil {
			return nil, err
		}
		descriptionSteps = append(descriptionSteps, builtStep)
	}

	pipelineDesc := &pipeline.PipelineDescription{
		Name:        p.name,
		Description: p.description,
		Steps:       descriptionSteps,
		Outputs:     pipelineOutputs,
	}

	return pipelineDesc, nil
}
