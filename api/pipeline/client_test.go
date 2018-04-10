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

	searchPipelinesRequest := &SearchPipelinesRequest{
		Problem: &ProblemDescription{
			Problem: &Problem{
				TaskType: TaskType_REGRESSION,
				PerformanceMetrics: []*ProblemPerformanceMetric{
					&ProblemPerformanceMetric{
						Metric: PerformanceMetric_MEAN_SQUARED_ERROR,
					},
				},
			},
			Inputs: []*ProblemInput{
				&ProblemInput{
					DatasetId: "196_autoMpg",
					Targets: []*ProblemTarget{
						&ProblemTarget{
							TargetIndex: 0,
							ResourceId:  "0",
							ColumnIndex: 8,
							ColumnName:  "class",
						},
					},
				},
			},
		},
	}

	searchID, err := client.StartSearch(context.Background(), searchPipelinesRequest)
	assert.NoError(t, err)

	pipelines, err := client.SearchPipelines(context.Background(), searchID)
	assert.NoError(t, err)

	for _, pipeline := range pipelines {

		assert.NotEmpty(t, pipeline.PipelineId)

		_, err := client.GeneratePipelineScores(context.Background(), pipeline.PipelineId)
		assert.NoError(t, err)

		_, err = client.GeneratePipelineFit(context.Background(), pipeline.PipelineId)
		assert.NoError(t, err)

		producePipelineRequest := &ProducePipelineRequest{
			PipelineId: pipeline.PipelineId,
			Inputs: []*Value{
				{
					Value: &Value_DatasetUri{
						DatasetUri: "testURI",
					},
				},
			},
		}

		_, err = client.GeneratePredictions(context.Background(), producePipelineRequest)
		assert.NoError(t, err)
	}

	err = client.EndSearch(context.Background(), searchID)
	assert.NoError(t, err)
}
*/
