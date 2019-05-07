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

package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/primitive/compute"
	api "github.com/uncharted-distil/distil/api/compute"
	"github.com/uncharted-distil/distil/api/env"
	"github.com/uncharted-distil/distil/api/model"
	jutil "github.com/uncharted-distil/distil/api/util/json"
)

const (
	createSolutions   = "CREATE_SOLUTIONS"
	stopSolutions     = "STOP_SOLUTIONS"
	categoricalType   = "categorical"
	numericalType     = "numerical"
	defaultResourceID = "learningData"
	datasetSizeLimit  = 10000
)

var (
	problemFile = ""
)

// SetProblemFile sets the problem file containing the metrics to use
// when submitting pipelines
func SetProblemFile(file string) {
	problemFile = file
}

// SolutionHandler represents a solution websocket handler.
func SolutionHandler(client *compute.Client, metadataCtor model.MetadataStorageCtor, dataCtor model.DataStorageCtor, solutionCtor model.SolutionStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// create conn
		conn, err := NewConnection(w, r, handleSolutionMessage(client, metadataCtor, dataCtor, solutionCtor))
		if err != nil {
			log.Warn(err)
			return
		}
		// listen for requests and respond
		err = conn.ListenAndRespond()
		if err != nil {
			log.Info(err)
		}
		// clean up conn internals
		conn.Close()
	}
}

func handleSolutionMessage(client *compute.Client, metadataCtor model.MetadataStorageCtor, dataCtor model.DataStorageCtor, solutionCtor model.SolutionStorageCtor) func(conn *Connection, bytes []byte) {
	return func(conn *Connection, bytes []byte) {
		// parse the message
		msg, err := NewMessage(bytes)
		if err != nil {
			// send error response
			handleErr(conn, nil, errors.Wrap(err, fmt.Sprintf("unable to parse solution request message: %s", string(bytes))))
			return
		}
		// handle message
		go handleMessage(conn, client, metadataCtor, dataCtor, solutionCtor, msg)
	}
}

func parseMessage(bytes []byte) (*Message, error) {
	var msg *Message
	err := json.Unmarshal(bytes, &msg)
	if err != nil {
		return nil, err
	}
	msg.Timestamp = time.Now()
	return msg, nil
}

func handleMessage(conn *Connection, client *compute.Client, metadataCtor model.MetadataStorageCtor, dataCtor model.DataStorageCtor, solutionCtor model.SolutionStorageCtor, msg *Message) {
	switch msg.Type {
	case createSolutions:
		handleCreateSolutions(conn, client, metadataCtor, dataCtor, solutionCtor, msg)
		return
	case stopSolutions:
		handleStopSolutions(conn, client, msg)
		return
	default:
		// unrecognized type
		handleErr(conn, msg, errors.New("unrecognized message type"))
		return
	}
}

func handleCreateSolutions(conn *Connection, client *compute.Client, metadataCtor model.MetadataStorageCtor, dataCtor model.DataStorageCtor, solutionCtor model.SolutionStorageCtor, msg *Message) {
	// unmarshal request
	request, err := api.NewSolutionRequest(msg.Raw)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to unmarshal create solutions request"))
		return
	}

	// initialize the storage
	dataStorage, err := dataCtor()
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to initialize data storage"))
		return
	}

	// initialize metadata storage
	metaStorage, err := metadataCtor()
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to initialize meta storage"))
		return
	}

	// initialize solution storage
	solutionStorage, err := solutionCtor()
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to initialize solution storage"))
		return
	}

	targetVar, err := metaStorage.FetchVariable(request.Dataset, request.TargetFeature)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to fetch target variable definition"))
		return
	}

	// load defaults
	config, _ := env.LoadConfig()
	if request.Task == "" {
		request.Task = api.DefaultTaskType(targetVar.Type, request.ProblemType)
		log.Infof("Defaulting task type to `%s`", request.Task)
	}
	if request.SubTask == "" {
		request.SubTask = api.DefaultTaskSubType(request.Task)
		log.Infof("Defaulting task sub type to `%s`", request.SubTask)
	}
	if len(request.Metrics) == 0 {
		request.Metrics = api.DefaultMetrics(request.Task)
		log.Infof("Defaulting metrics to `%s`", strings.Join(request.Metrics, ","))
	}
	if request.MaxTime == 0 {
		request.MaxTime = int64(config.SolutionSearchMaxTime)
		log.Infof("Defaulting max search time to `%d`", request.MaxTime)
	}

	// persist the request information and dispatch the request
	err = request.PersistAndDispatch(client, solutionStorage, metaStorage, dataStorage)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to dispatch solution request to TA2"))
		return
	}

	// listen for solution updates
	err = request.Listen(func(status api.SolutionStatus) {
		// check for error
		if status.Error != nil {
			handleErr(conn, msg, errors.Wrap(status.Error, "received error from TA2 system"))
			return
		}
		// send status to client
		handleSuccess(conn, msg, jutil.StructToMap(status))
	})
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "received error from TA2 system"))
		return
	}

	// complete the request
	handleComplete(conn, msg)
}

func handleStopSolutions(conn *Connection, client *compute.Client, msg *Message) {
	// unmarshal request
	request, err := api.NewStopSolutionSearchRequest(msg.Raw)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to unmarshal stop solutions request"))
		return
	}

	// dispatch request
	err = request.Dispatch(client)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "received error from TA2 system"))
		return
	}
	return
}
