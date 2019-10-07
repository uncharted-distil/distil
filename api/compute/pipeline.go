//
//   Copyright © 2019 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package compute

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/pipeline"
	"github.com/uncharted-distil/distil-compute/primitive/compute"

	"github.com/uncharted-distil/distil/api/env"
)

// SubmitPipeline executes pipelines using the client and returns the result URI.
func SubmitPipeline(client *compute.Client, datasets []string,
	searchRequest *pipeline.SearchSolutionsRequest, step *pipeline.PipelineDescription) (string, error) {

	config, err := env.LoadConfig()
	if err != nil {
		return "", errors.Wrap(err, "unable to load config")
	}

	if config.UseTA2Runner {
		res, err := client.ExecutePipeline(context.Background(), datasets, step)
		if err != nil {
			return "", errors.Wrap(err, "unable to dispatch mocked pipeline")
		}
		resultURI := strings.Replace(res.ResultURI, "file://", "", -1)
		return resultURI, nil
	}

	request := compute.NewExecPipelineRequest(datasets, step)

	err = request.Dispatch(client, searchRequest)
	if err != nil {
		return "", errors.Wrap(err, "unable to dispatch pipeline")
	}

	// listen for completion
	var errPipeline error
	var datasetURI string
	err = request.Listen(func(status compute.ExecPipelineStatus) {
		// check for error
		if status.Error != nil {
			errPipeline = status.Error
		}

		if status.Progress == compute.RequestCompletedStatus {
			datasetURI = status.ResultURI
		}
	})
	if err != nil {
		return "", errors.Wrap(err, "unable to listen to pipeline")
	}

	if errPipeline != nil {
		return "", errors.Wrap(errPipeline, "error executing pipeline")
	}

	datasetURI = strings.Replace(datasetURI, "file://", "", -1)

	return datasetURI, nil
}
