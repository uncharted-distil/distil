//
//   Copyright © 2021 Uncharted Software Inc.
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

package task

import (
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil/api/env"
	apiModel "github.com/uncharted-distil/distil/api/model"
)

// VerticalConcat will bring mastery.
func VerticalConcat(dataStorage apiModel.DataStorage, joinLeft *JoinSpec, joinRight *JoinSpec) (string, *apiModel.FilteredData, error) {
	pipelineDesc, err := description.CreateVerticalConcatPipeline("Unioner", "Combine existing data")
	if err != nil {
		return "", nil, err
	}
	datasetLeftURI := env.ResolvePath(joinLeft.DatasetSource, joinLeft.DatasetPath)
	datasetRightURI := env.ResolvePath(joinRight.DatasetSource, joinRight.DatasetPath)

	return join(joinLeft, joinRight, pipelineDesc, []string{datasetLeftURI, datasetRightURI}, defaultSubmitter{}, true)
}
