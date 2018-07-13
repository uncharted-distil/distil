package description

import (
	fmt "fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unchartedsoftware/distil/api/pipeline"
)

func createLabels(counter int64) []string {
	return []string{fmt.Sprintf("alpha-%d", counter), fmt.Sprintf("bravo-%d", counter)}
}

func createTestStep(step int64) *StepData {
	labels := createLabels(step)
	return NewStepDataWithHyperparameters(
		&pipeline.Primitive{
			Id:         fmt.Sprintf("0000-primtive-%d", step),
			Version:    "1.0.0",
			Name:       fmt.Sprintf("primitive-%d", step),
			PythonPath: fmt.Sprintf("d3m.primitives.distil.primitive.%d", step),
		},
		[]string{"produce"},
		map[string]interface{}{
			"testString":         fmt.Sprintf("hyperparam-%d", step),
			"testBool":           step%2 == 0,
			"testInt":            step,
			"testStringArray":    labels,
			"testBoolArray":      []bool{step%2 == 0, step%2 != 0},
			"testIntArray":       []int64{step, step + 1},
			"testIntMap":         map[string]int64{labels[0]: int64(step), labels[1]: int64(step)},
			"testNestedIntArray": [][]int64{{step, step + 1}, {step + 2, step + 3}},
			"testNestedIntMap":   map[string][]int64{labels[0]: {step, step + 1}, labels[1]: {step + 2, step + 3}},
		},
	)
}

func convertToStringArray(list *pipeline.ValueList) []string {
	arr := []string{}
	for _, v := range list.Items {
		arr = append(arr, v.GetString_())
	}
	return arr
}

func convertToBoolArray(list *pipeline.ValueList) []bool {
	arr := []bool{}
	for _, v := range list.Items {
		arr = append(arr, v.GetBool())
	}
	return arr
}

func convertToIntArray(list *pipeline.ValueList) []int64 {
	arr := []int64{}
	for _, v := range list.Items {
		arr = append(arr, v.GetInt64())
	}
	return arr
}

func convertToIntMap(dict *pipeline.ValueDict) map[string]int64 {
	mp := map[string]int64{}
	for k, v := range dict.Items {
		mp[k] = v.GetInt64()
	}
	return mp
}

func convertToNestedIntArray(list *pipeline.ValueList) [][]int64 {
	arr := [][]int64{}
	for _, v := range list.Items {
		inner := []int64{}
		for _, w := range v.GetList().Items {
			inner = append(inner, w.GetInt64())
		}
		arr = append(arr, inner)
	}
	return arr
}

func convertToNestedIntMap(dict *pipeline.ValueDict) map[string][]int64 {
	mp := map[string][]int64{}
	for k, v := range dict.Items {
		inner := []int64{}
		for _, w := range v.GetList().Items {
			inner = append(inner, w.GetInt64())
		}
		mp[k] = inner
	}
	return mp
}

func testStep(t *testing.T, index int64, step *StepData, steps []*pipeline.PipelineDescriptionStep) {
	labels := createLabels(index)

	assert.Equal(t, "produce", steps[index].GetPrimitive().GetOutputs()[0].GetId())

	assert.Equal(t, fmt.Sprintf("hyperparam-%d", index),
		steps[index].GetPrimitive().GetHyperparams()["testString"].GetValue().GetData().GetRaw().GetString_())

	assert.Equal(t, int64(index), steps[index].GetPrimitive().GetHyperparams()["testInt"].GetValue().GetData().GetRaw().GetInt64())

	assert.Equal(t, index%2 == 0, steps[index].GetPrimitive().GetHyperparams()["testBool"].GetValue().GetData().GetRaw().GetBool())

	assert.Equal(t, labels,
		convertToStringArray(steps[index].GetPrimitive().GetHyperparams()["testStringArray"].GetValue().GetData().GetRaw().GetList()))

	assert.Equal(t, []int64{int64(index), int64(index) + 1},
		convertToIntArray(steps[index].GetPrimitive().GetHyperparams()["testIntArray"].GetValue().GetData().GetRaw().GetList()))

	assert.Equal(t, []bool{index%2 == 0, index%2 != 0},
		convertToBoolArray(steps[index].GetPrimitive().GetHyperparams()["testBoolArray"].GetValue().GetData().GetRaw().GetList()))

	assert.Equal(t, map[string]int64{labels[0]: int64(index), labels[1]: int64(index)},
		convertToIntMap(steps[index].GetPrimitive().GetHyperparams()["testIntMap"].GetValue().GetData().GetRaw().GetDict()))

	assert.Equal(t, [][]int64{{index, index + 1}, {index + 2, index + 3}},
		convertToNestedIntArray(steps[index].GetPrimitive().GetHyperparams()["testNestedIntArray"].GetValue().GetData().GetRaw().GetList()))

	assert.Equal(t, map[string][]int64{labels[0]: {index, index + 1}, labels[1]: {index + 2, index + 3}},
		convertToNestedIntMap(steps[index].GetPrimitive().GetHyperparams()["testNestedIntMap"].GetValue().GetData().GetRaw().GetDict()))

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
