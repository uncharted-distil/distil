package pipeline

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestUserPipeline(t *testing.T) {
	pipeline, err := CreateUserDatasetPipeline()
	assert.NoError(t, err)
	t.Logf("\n%s", proto.MarshalTextString(pipeline))
}
