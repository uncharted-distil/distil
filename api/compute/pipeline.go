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
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/hashstructure"
	gc "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"

	"github.com/uncharted-distil/distil-compute/pipeline"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil/api/env"
	"github.com/uncharted-distil/distil/api/util"
	log "github.com/unchartedsoftware/plog"
)

var (
	queue *Queue
	cache *Cache
)

// Cache is used to cache data in memory. It can be persisted to disk as needed.
type Cache struct {
	cache      *gc.Cache
	sourceFile string
}

// PersistCache stores the cache to disk.
func (c *Cache) PersistCache() error {
	items := cache.cache.Items()
	b := new(bytes.Buffer)

	encoder := gob.NewEncoder(b)

	err := encoder.Encode(items)
	if err != nil {
		return errors.Wrap(err, "unable to encode cache")
	}

	err = util.WriteFileWithDirs(c.sourceFile, b.Bytes(), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

type pipelineQueueTask struct {
	client          *compute.Client
	request         *compute.ExecPipelineRequest
	searchRequest   *pipeline.SearchSolutionsRequest
	step            *description.FullySpecifiedPipeline
	datasets        []string
	datasetsProduce []string
}

func (t *pipelineQueueTask) hashUnique() (string, error) {
	// use the json representation of the step since the nested structures
	// require casting that the library can't handle
	stepString, err := description.MarshalSteps(t.step.Pipeline)
	if err != nil {
		return "", err
	}

	hashedPipeline, err := hashstructure.Hash([]interface{}{stepString, t.datasets, t.datasetsProduce, t.searchRequest}, nil)
	if err != nil {
		return "", errors.Wrap(err, "unable to uniquely hash pipeline")
	}
	hashedPipelineKey := fmt.Sprintf("%d", hashedPipeline)

	return hashedPipelineKey, nil
}

func (t *pipelineQueueTask) hashEquivalent() (string, error) {
	hashedPipeline, err := hashstructure.Hash([]interface{}{t.step.EquivalentValues, t.datasets, t.datasetsProduce, t.searchRequest}, nil)
	if err != nil {
		return "", errors.Wrap(err, "unable to equivalently hash pipeline")
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
	inProgress    *QueueItem
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
		queuedItem.data = data
		q.mu.Unlock()

		return output
	}
	if q.inProgress != nil && q.inProgress.key == key {
		log.Infof("'%s' already in progress so adding one more channel to output", key)
		q.inProgress.output = append(q.inProgress.output, output)
		q.inProgress.data = data
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
func (q *Queue) Dequeue() (*QueueItem, bool) {
	log.Infof("dequeuing data from the queue")
	item, ok := <-q.tasks
	if !ok {
		return item, ok
	}

	q.mu.Lock()
	q.alreadyQueued[item.key] = nil
	q.inProgress = item
	q.mu.Unlock()

	return item, true
}

// Done flags a task queue as being completed, which removes it from the in progress slot.
func (q *Queue) Done() {
	q.mu.Lock()
	q.inProgress = nil
	q.mu.Unlock()
}

// InitializeCache sets up an empty cache or if a source file provided, reads
// the cache from the source file.
func InitializeCache(sourceFile string) error {
	var c *gc.Cache
	if util.FileExists(sourceFile) {
		b, err := ioutil.ReadFile(sourceFile)
		if err != nil {
			return errors.Wrap(err, "unable to read cache file")
		}

		var decodedMap map[string]gc.Item
		d := gob.NewDecoder(bytes.NewReader(b))
		err = d.Decode(&decodedMap)
		if err != nil {
			return errors.Wrap(err, "unable to decode cache map")
		}

		c = gc.NewFrom(24*time.Hour, 48*time.Hour, decodedMap)
	} else {
		c = gc.New(24*time.Hour, 48*time.Hour)
	}

	cache = &Cache{
		cache:      c,
		sourceFile: sourceFile,
	}

	return nil
}

// InitializeQueue creates the pipeline queue and runs go routine to process pipeline requests
func InitializeQueue(config *env.Config) {
	queue = &Queue{
		tasks:         make(chan *QueueItem, config.PipelineQueueSize),
		alreadyQueued: make(map[string]*QueueItem),
	}

	go runPipelineQueue(queue)
}

// SubmitPipeline executes pipelines using the client and returns the result URI.
func SubmitPipeline(client *compute.Client, datasets []string, datasetsProduce []string,
	searchRequest *pipeline.SearchSolutionsRequest, fullySpecifiedStep *description.FullySpecifiedPipeline, shouldCache bool) (string, error) {

	request := compute.NewExecPipelineRequest(datasets, datasetsProduce, fullySpecifiedStep.Pipeline)

	queueTask := &pipelineQueueTask{
		request:         request,
		searchRequest:   searchRequest,
		client:          client,
		step:            fullySpecifiedStep,
		datasets:        datasets,
		datasetsProduce: datasetsProduce,
	}
	// check cache to see if results are already available
	hashedPipelineUniqueKey, err := queueTask.hashUnique()
	if shouldCache{
	if err != nil {
		return "", err
	}
	entry, found := cache.cache.Get(hashedPipelineUniqueKey)
	if found {
		log.Infof("returning cached entry for pipeline")
		return entry.(string), nil
	}
	}
	// get equivalency key for enqueuing
	hashedPipelineEquivKey, err := queueTask.hashEquivalent()
	if err != nil {
		return "", err
	}

	resultChan := queue.Enqueue(hashedPipelineEquivKey, queueTask)

	result := <-resultChan
	if result.Error != nil {
		return "", result.Error
	}

	datasetURI := result.Output.(string)
	cache.cache.Set(hashedPipelineUniqueKey, datasetURI, gc.DefaultExpiration)
	queue.Done()
	err = cache.PersistCache()
	if err != nil {
		log.Warnf("error persisting cache: %v", err)
	}

	return datasetURI, nil
}

func runPipelineQueue(queue *Queue) {
	for {
		queueTask, ok := queue.Dequeue()
		if !ok {
			break
		}
		log.Infof("processing data pulled from the queue (key '%s')", queueTask.key)

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

	log.Infof("ending queue processing")
}

func (qi *QueueItem) returnResult(response *QueueResponse) {
	for _, oc := range qi.output {
		oc <- response
	}
}
