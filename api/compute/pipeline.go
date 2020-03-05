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
	"github.com/uncharted-distil/distil/api/env"
	log "github.com/unchartedsoftware/plog"
)

var (
	pipelineCache *Cache
	queue         *Queue
)

type pipelineQueueTask struct {
	client          *compute.Client
	request         *compute.ExecPipelineRequest
	searchRequest   *pipeline.SearchSolutionsRequest
	step            *pipeline.PipelineDescription
	datasets        []string
	datasetsProduce []string
}

func (t *pipelineQueueTask) hash() (string, error) {
	// use the json representation of the step since the nested structures
	// require casting that the library can't handle
	stepString, err := marshalSteps(t.step)
	if err != nil {
		return "", err
	}

	hashedPipeline, err := hashstructure.Hash([]interface{}{stepString, t.datasets, t.datasetsProduce, t.searchRequest}, nil)
	if err != nil {
		return "", errors.Wrap(err, "unable to hash pipeline")
	}
	hashedPipelineKey := fmt.Sprintf("%d", hashedPipeline)

	return hashedPipelineKey, nil
}

// QueueItem is the wrapper for the data to process and the response channel.
type QueueItem struct {
	key    string
	output []chan *QueueResponse
	data   interface{}
}

// QueueResponse represents the result from processing a queue item.
type QueueResponse struct {
	Output interface{}
	Error  error
}

// Queue uses a buffered channel to queue tasks and provides the result via channels.
type Queue struct {
	mu            sync.RWMutex
	tasks         chan *QueueItem
	alreadyQueued map[string]*QueueItem
}

// Enqueue adds one entry to the queue, providing the response channel as result.
// If the key is already in the queue, then the data is not added a second time.
// Rather, a new output channel is added
func (q *Queue) Enqueue(key string, data interface{}) chan *QueueResponse {
	log.Infof("enqueuing data in the queue")
	output := make(chan *QueueResponse, 1)

	// use key to check if it is already in the queue
	q.mu.Lock()
	queuedItem := q.alreadyQueued[key]
	if queuedItem != nil {
		log.Infof("'%s' already in queue so adding one more channel to output", key)
		queuedItem.output = append(queuedItem.output, output)
		q.mu.Unlock()

		return output
	}
	log.Infof("'%s' not in queue so creating new item", key)
	item := &QueueItem{
		key:    key,
		data:   data,
		output: []chan *QueueResponse{output},
	}
	q.alreadyQueued[key] = item
	q.mu.Unlock()

	q.tasks <- item

	return output
}

// Dequeue removes one item from the queue.
func (q *Queue) Dequeue() *QueueItem {
	log.Infof("dequeuing data from the queue")
	item := <-q.tasks

	q.mu.Lock()
	q.alreadyQueued[item.key] = nil
	q.mu.Unlock()

	return item
}

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

// InitializeQueue creates the pipeline queue and runs go routine to process pipeline requests
func InitializeQueue(config *env.Config) {
	queue = &Queue{
		tasks: make(chan *QueueItem, config.PipelineQueueSize),
	}

	go runPipelineQueue(queue)
}

// SubmitPipeline executes pipelines using the client and returns the result URI.
func SubmitPipeline(client *compute.Client, datasets []string, datasetsProduce []string,
	searchRequest *pipeline.SearchSolutionsRequest, step *pipeline.PipelineDescription) (string, error) {

	request := compute.NewExecPipelineRequest(datasets, datasetsProduce, step)

	queueTask := &pipelineQueueTask{
		request:         request,
		searchRequest:   searchRequest,
		client:          client,
		step:            step,
		datasets:        datasets,
		datasetsProduce: datasetsProduce,
	}

	// check cache to see if results are already available
	hashedPipelineKey, err := queueTask.hash()
	if err != nil {
		return "", err
	}
	entry, found := pipelineCache.Get(hashedPipelineKey)
	if found {
		log.Infof("returning cached entry for pipeline")
		return entry.(string), nil
	}

	resultChan := queue.Enqueue(hashedPipelineKey, queueTask)

	result := <-resultChan
	if result.Error != nil {
		return "", result.Error
	}

	datasetURI := result.Output.(string)
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

func runPipelineQueue(queue *Queue) {
	for queueTask := range queue.tasks {
		log.Infof("processing data pulled from the queue")

		pipelineTask, ok := queueTask.data.(*pipelineQueueTask)
		if !ok {
			queueTask.returnResult(&QueueResponse{
				Error: errors.Errorf("data pulled from queue is not a pipeline"),
			})
			continue
		}

		err := pipelineTask.request.Dispatch(pipelineTask.client, pipelineTask.searchRequest)
		if err != nil {
			queueTask.returnResult(&QueueResponse{
				Error: errors.Wrap(err, "unable to dispatch pipeline"),
			})
			continue
		}

		// listen for completion
		var errPipeline error
		var datasetURI string
		err = pipelineTask.request.Listen(func(status compute.ExecPipelineStatus) {
			// check for error
			if status.Error != nil {
				errPipeline = status.Error
			}

			if status.Progress == compute.RequestCompletedStatus {
				datasetURI = status.ResultURI
			}
		})
		if err != nil {
			queueTask.returnResult(&QueueResponse{
				Error: errors.Wrap(err, "unable to listen to pipeline"),
			})
			continue
		}

		if errPipeline != nil {
			queueTask.returnResult(&QueueResponse{
				Error: errors.Wrap(errPipeline, "error executing pipeline"),
			})
			continue
		}

		datasetURI = strings.Replace(datasetURI, "file://", "", -1)

		queueTask.returnResult(&QueueResponse{Output: datasetURI})
	}
}

func (qi *QueueItem) returnResult(response *QueueResponse) {
	for _, oc := range qi.output {
		oc <- response
	}
}
