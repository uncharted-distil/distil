package pipeline

/*
import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)
func TestClient(t *testing.T) {

	client, err := NewClient("localhost:45042", "./datasets", true)
	assert.NoError(t, err)

	searchID, err := client.StartSearch(context.Background())
	assert.NoError(t, err)

	pipelines, err := client.GenerateCandidatePipelines(context.Background(), searchID)
	assert.NoError(t, err)

	for _, pipeline := range pipelines {

		assert.NotEmpty(t, pipeline.PipelineId)

		_, err := client.GenerateScoresForCandidatePipeline(context.Background(), pipeline.PipelineId)
		assert.NoError(t, err)

		_, err = client.GeneratePipelineFit(context.Background(), pipeline.PipelineId)
		assert.NoError(t, err)

		_, err = client.GeneratePredictions(context.Background(), pipeline.PipelineId)
		assert.NoError(t, err)
	}

	err = client.EndSearch(context.Background(), searchID)
	assert.NoError(t, err)
}
*/
