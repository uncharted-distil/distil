package description

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/unchartedsoftware/distil/api/model"
)

func TestCreateUserDatasetPipeline(t *testing.T) {

	variables := []*model.Variable{
		{
			Key:          "test_var_0",
			OriginalType: "ordinal",
			Type:         "categorical",
			Index:        0,
		},
		{
			Key:          "test_var_1",
			OriginalType: "categorical",
			Type:         "integer",
			Index:        1,
		},
		{
			Key:          "test_var_2",
			OriginalType: "categorical",
			Type:         "integer",
			Index:        2,
		},
		{
			Key:          "test_var_3",
			OriginalType: "categorical",
			Type:         "integer",
			Index:        3,
		},
	}

	pipeline, err := CreateUserDatasetPipeline(
		"test_user_pipeline", "a test user pipeline", variables, "test_target", []string{"test_var_0", "test_var_1", "test_var_3"})

	// assert 1st is a semantic type update
	hyperParams := pipeline.GetSteps()[0].GetPrimitive().GetHyperparams()
	assert.Equal(t, []int64{1, 3}, ConvertToIntArray(hyperParams["add_columns"].GetValue().GetData().GetRaw().GetList()))
	assert.Equal(t, []string{"http://schema.org/Integer"}, ConvertToStringArray(hyperParams["add_types"].GetValue().GetData().GetRaw().GetList()))
	assert.Equal(t, []int64{}, ConvertToIntArray(hyperParams["remove_columns"].GetValue().GetData().GetRaw().GetList()))
	assert.Equal(t, []string{""}, ConvertToStringArray(hyperParams["remove_types"].GetValue().GetData().GetRaw().GetList()))

	// assert 2nd is a semantic type update
	hyperParams = pipeline.GetSteps()[1].GetPrimitive().GetHyperparams()
	assert.Equal(t, []int64{}, ConvertToIntArray(hyperParams["add_columns"].GetValue().GetData().GetRaw().GetList()))
	assert.Equal(t, []string{""}, ConvertToStringArray(hyperParams["add_types"].GetValue().GetData().GetRaw().GetList()))
	assert.Equal(t, []int64{1, 3}, ConvertToIntArray(hyperParams["remove_columns"].GetValue().GetData().GetRaw().GetList()))
	assert.Equal(t, []string{"https://metadata.datadrivendiscovery.org/types/CategoricalData"},
		ConvertToStringArray(hyperParams["remove_types"].GetValue().GetData().GetRaw().GetList()))

	// assert 3rd step is column remove and index two was remove
	hyperParams = pipeline.GetSteps()[2].GetPrimitive().GetHyperparams()
	assert.Equal(t, "0", hyperParams["resource_id"].GetValue().GetData().GetRaw().GetString_(), "0")
	assert.Equal(t, []int64{2}, ConvertToIntArray(hyperParams["columns"].GetValue().GetData().GetRaw().GetList()))

	assert.NoError(t, err)
	t.Logf("\n%s", proto.MarshalTextString(pipeline))
}

func TestCreateUserDatasetPipelineMappingError(t *testing.T) {

	variables := []*model.Variable{
		{
			Key:          "test_var_0",
			OriginalType: "blordinal",
			Type:         "categorical",
			Index:        0,
		},
	}

	pipeline, err := CreateUserDatasetPipeline(
		"test_user_pipeline", "a test user pipeline", variables, "test_target", []string{"test_var_0"})
	assert.Error(t, err)
	t.Logf("\n%s", proto.MarshalTextString(pipeline))
}

func TestCreateUserDatasetEmpty(t *testing.T) {

	variables := []*model.Variable{
		{
			Key:          "test_var_0",
			OriginalType: "categorical",
			Type:         "categorical",
			Index:        0,
		},
	}

	pipeline, err := CreateUserDatasetPipeline(
		"test_user_pipeline", "a test user pipeline", variables, "test_target", []string{"test_var_0"})

	assert.Nil(t, pipeline)
	assert.Nil(t, err)

	t.Logf("\n%s", proto.MarshalTextString(pipeline))
}
