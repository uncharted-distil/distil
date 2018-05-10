package pipeline

import (
	fmt "fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func CreateTestStep(step int) *StepData {
	return NewStepDataWithHyperparameters(
		&Primitive{
			Id:         fmt.Sprintf("0000-primtive-%d", step),
			Version:    "1.0.0",
			Name:       fmt.Sprintf("primitive-%d", step),
			PythonPath: fmt.Sprintf("d3m.primitives.distil.primitive.%d", step),
		},
		[]string{"produce"},
		map[string]interface{}{
			"test": fmt.Sprintf("hyperparam-%d", step),
		},
	)
}
func TestPipelineCompile(t *testing.T) {

	step0 := CreateTestStep(0)
	step1 := CreateTestStep(1)
	step2 := CreateTestStep(2)

	desc, err := NewDescriptionBuilder("test pipeline", "test pipelne consisting of 3 stages").
		Add(step0).
		Add(step1).
		Add(step2).
		Compile()
	assert.NoError(t, err)

	steps := desc.GetSteps()
	assert.Equal(t, len(steps), 3)

	// validate step inputs
	assert.Equal(t, "inputs.0", steps[0].GetPrimitive().GetArguments()["inputs"].GetContainer().GetData())
	assert.Equal(t, "produce", steps[0].GetPrimitive().GetOutputs()[0].GetId())
	assert.Equal(t, "hyperparam-0", steps[0].GetPrimitive().GetHyperparams()["test"].GetValue().GetData().GetString_())
	assert.EqualValues(t, step0.GetPrimitive(), steps[0].GetPrimitive().GetPrimitive())

	assert.Equal(t, "steps.0.produce", steps[1].GetPrimitive().GetArguments()["inputs"].GetContainer().GetData())
	assert.Equal(t, "produce", steps[1].GetPrimitive().GetOutputs()[0].GetId())
	assert.Equal(t, "hyperparam-1", steps[1].GetPrimitive().GetHyperparams()["test"].GetValue().GetData().GetString_())
	assert.EqualValues(t, step1.GetPrimitive(), steps[1].GetPrimitive().GetPrimitive())

	assert.Equal(t, "steps.1.produce", steps[2].GetPrimitive().GetArguments()["inputs"].GetContainer().GetData())
	assert.Equal(t, "produce", steps[2].GetPrimitive().GetOutputs()[0].GetId())
	assert.Equal(t, "hyperparam-2", steps[2].GetPrimitive().GetHyperparams()["test"].GetValue().GetData().GetString_())
	assert.EqualValues(t, step2.GetPrimitive(), steps[2].GetPrimitive().GetPrimitive())

	// validate outputs
	assert.Equal(t, 1, len(desc.GetOutputs()))
	assert.Equal(t, "steps.2.produce", desc.GetOutputs()[0].GetData())
}
