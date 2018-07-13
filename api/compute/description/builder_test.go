package description

import (
	fmt "fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unchartedsoftware/distil/api/pipeline"
)

func createTestStep(step int) *StepData {
	return NewStepDataWithHyperparameters(
		&pipeline.Primitive{
			Id:         fmt.Sprintf("0000-primtive-%d", step),
			Version:    "1.0.0",
			Name:       fmt.Sprintf("primitive-%d", step),
			PythonPath: fmt.Sprintf("d3m.primitives.distil.primitive.%d", step),
		},
		[]string{"produce"},
		map[string]interface{}{
			"testString":      fmt.Sprintf("hyperparam-%d", step),
			"testBool":        step%2 == 0,
			"testInt":         step,
			"testStringArray": []string{fmt.Sprintf("alpha-%d", step), fmt.Sprintf("bravo-%d", step)},
			"testBoolArray":   []bool{step%2 == 0, step%2 != 0},
			"testIntArray":    []int{step, step + 1},
		},
	)
}

func testStep(t *testing.T, index int, step *StepData, steps []*pipeline.PipelineDescriptionStep) {
	assert.Equal(t, "produce", steps[index].GetPrimitive().GetOutputs()[0].GetId())

	assert.Equal(t, fmt.Sprintf("hyperparam-%d", index),
		steps[index].GetPrimitive().GetHyperparams()["testString"].GetValue().GetData().GetRaw().GetString_())
	assert.Equal(t, int64(index), steps[index].GetPrimitive().GetHyperparams()["testInt"].GetValue().GetData().GetRaw().GetInt64())
	assert.Equal(t, index%2 == 0, steps[index].GetPrimitive().GetHyperparams()["testBool"].GetValue().GetData().GetRaw().GetBool())
	assert.Equal(t, []string{fmt.Sprintf("alpha-%d", index), fmt.Sprintf("bravo-%d", index)},
		steps[index].GetPrimitive().GetHyperparams()["testStringArray"].GetValue().GetData().GetRaw().GetList())
	assert.Equal(t, []int64{int64(index), int64(index) + 1},
		steps[index].GetPrimitive().GetHyperparams()["testIntArray"].GetValue().GetData().GetRaw().GetList())
	assert.Equal(t, []bool{index%2 == 0, index%2 != 0},
		steps[index].GetPrimitive().GetHyperparams()["testBoolArray"].GetValue().GetData().GetRaw().GetList())

	assert.EqualValues(t, step.GetPrimitive(), steps[index].GetPrimitive().GetPrimitive())
}

// Tests basic pipeline compilation.
func TestPipelineCompile(t *testing.T) {

	step0 := createTestStep(0)
	step1 := createTestStep(1)
	step2 := createTestStep(2)

	desc, err := NewBuilder("test pipeline", "test pipelne consisting of 3 stages").
		Add(step0).
		Add(step1).
		Add(step2).
		Compile()
	assert.NoError(t, err)

	steps := desc.GetSteps()
	assert.Equal(t, len(steps), 3)

	// validate step inputs
	assert.Equal(t, "inputs.0", steps[0].GetPrimitive().GetArguments()[stepInputsKey].GetContainer().GetData())
	testStep(t, 0, step0, steps)

	assert.Equal(t, "steps.0.produce", steps[1].GetPrimitive().GetArguments()[stepInputsKey].GetContainer().GetData())
	testStep(t, 1, step1, steps)

	assert.Equal(t, "steps.1.produce", steps[2].GetPrimitive().GetArguments()[stepInputsKey].GetContainer().GetData())
	testStep(t, 2, step2, steps)

	// validate outputs
	assert.Equal(t, 1, len(desc.GetOutputs()))
	assert.Equal(t, "steps.2.produce", desc.GetOutputs()[0].GetData())
}

// Tests proper compilation of an inference point.
func TestPipelineCompileWithInference(t *testing.T) {

	step0 := createTestStep(0)
	step1 := createTestStep(1)

	desc, err := NewBuilder("test pipeline", "test pipelne consisting of 3 stages").
		Add(step0).
		Add(step1).
		AddInferencePoint().
		Compile()
	assert.NoError(t, err)

	steps := desc.GetSteps()
	assert.Equal(t, len(steps), 3)

	assert.Equal(t, "steps.1.produce", steps[2].GetPlaceholder().GetInputs()[0].GetData())
	assert.Equal(t, "produce", steps[2].GetPlaceholder().GetOutputs()[0].GetId())
	assert.Nil(t, steps[2].GetPrimitive().GetHyperparams())
	assert.Nil(t, steps[2].GetPrimitive().GetPrimitive())

	// validate outputs
	assert.Equal(t, 1, len(desc.GetOutputs()))
	assert.Equal(t, "steps.2.produce", desc.GetOutputs()[0].GetData())
}
