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
		},
		{
			Key:          "test_var_1",
			OriginalType: "categorical",
			Type:         "integer",
		},
		{
			Key:          "test_var_2",
			OriginalType: "categorical",
			Type:         "integer",
		},
	}

	pipeline, err := CreateUserDatasetPipeline(
		"test_user_pipeline", "a test user pipeline", variables, []string{"test_var_0", "test_var_1"}, "test_var_1")
	assert.NoError(t, err)
	t.Logf("\n%s", proto.MarshalTextString(pipeline))
}

func TestCreateUserDatasetPipelineMappingError(t *testing.T) {

	variables := []*model.Variable{
		{
			Key:          "test_var_0",
			OriginalType: "blordinal",
			Type:         "categorical",
		},
	}

	pipeline, err := CreateUserDatasetPipeline(
		"test_user_pipeline", "a test user pipeline", variables, []string{"test_var_0"}, "test_var_1")
	assert.Error(t, err)
	t.Logf("\n%s", proto.MarshalTextString(pipeline))
}
