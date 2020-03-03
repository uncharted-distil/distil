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
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/mitchellh/hashstructure"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/pipeline"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	log "github.com/unchartedsoftware/plog"
)

var (
	pipelineCache *Cache
)

// Cache uses a simple map to lookup data stored in memory. Access to the cache
// is threadsafe.
type Cache struct {
	cache map[string]interface{}
	mu    sync.RWMutex
}

// Set sets the cached value for the specified key.
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	c.cache[key] = value
	c.mu.Unlock()
}

// Get reads cached value using the key.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	value, found := c.cache[key]
	c.mu.RUnlock()

	return value, found
}

// InitializeCache sets up an empty cache
func InitializeCache() {
	pipelineCache = &Cache{
		cache: make(map[string]interface{}),
	}
}

// SubmitPipeline executes pipelines using the client and returns the result URI.
func SubmitPipeline(client *compute.Client, datasets []string, datasetsProduce []string,
	searchRequest *pipeline.SearchSolutionsRequest, step *pipeline.PipelineDescription) (string, error) {

	request := compute.NewExecPipelineRequest(datasets, datasetsProduce, step)

	// check cache to see if results are already available
	stepString, err := marshalSteps(step)
	if err != nil {
		return "", err
	}

	hashedPipeline, err := hashstructure.Hash([]interface{}{stepString, datasets, datasetsProduce, searchRequest}, nil)
	if err != nil {
		return "", errors.Wrap(err, "unable to hash pipeline")
	}
	hashedPipelineKey := fmt.Sprintf("%d", hashedPipeline)
	log.Infof("hash key: %s\traw: %v", hashedPipelineKey, hashedPipeline)

	entry, found := pipelineCache.Get(hashedPipelineKey)
	if found {
		log.Infof("returning cached entry for pipeline")
		return entry.(string), nil
	}

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

	pipelineCache.Set(hashedPipelineKey, datasetURI)

	return datasetURI, nil
}

func marshalSteps(step *pipeline.PipelineDescription) (string, error) {
	stepJSON, err := json.Marshal(step)
	if err != nil {
		return "", errors.Wrapf(err, "unable to marshal steps")
	}

	return string(stepJSON), nil
}
