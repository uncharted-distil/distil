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
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
	log "github.com/unchartedsoftware/plog"
)

const (
	// D3MProblem name of the expected problem file.
	D3MProblem = "problemDoc.json"

	problemVersion       = "2.0"
	problemSchemaVersion = "4.0.0"
)

// VariableProvider defines a function that will get the provided variable.
type VariableProvider func(dataset string, index string, name string) (*model.Variable, error)

// ProblemPersist contains the problem file data.
type ProblemPersist struct {
	About           *ProblemPersistAbout          `json:"about"`
	Inputs          *ProblemPersistInput          `json:"inputs"`
	ExpectedOutputs *ProblemPersistExpectedOutput `json:"expectedOutputs,omitempty"`
}

// ProblemPersistAbout represents the basic information of a problem.
type ProblemPersistAbout struct {
	ProblemID            string   `json:"problemID"`
	ProblemName          string   `json:"problemName"`
	ProblemDescription   string   `json:"problemDescription"`
	TaskKeywords         []string `json:"taskKeywords"`
	ProblemVersion       string   `json:"problemVersion"`
	ProblemSchemaVersion string   `json:"problemSchemaVersion"`
}

// ProblemPersistInput lists the information of a problem.
type ProblemPersistInput struct {
	Data               []*ProblemPersistData              `json:"data"`
	PerformanceMetrics []*ProblemPersistPerformanceMetric `json:"performanceMetrics"`
	DataSplits         *ProblemPersistDataSplits          `json:"dataSplits,omitempty"`
}

// ProblemPersistDataSplits contains the information about the data splits.
type ProblemPersistDataSplits struct {
	Method     string  `json:"method"`
	TestSize   float64 `json:"testSize"`
	Stratified bool    `json:"stratified"`
	NumRepeats int     `json:"numRepeats"`
	RandomSeed int     `json:"randomSeed"`
	SplitsFile string  `json:"splitsFile"`
}

// ProblemPersistData ties targets to a dataset.
type ProblemPersistData struct {
	DatasetID string                  `json:"datasetID"`
	Targets   []*ProblemPersistTarget `json:"targets"`
}

// ProblemPersistTarget represents the target information of the problem.
type ProblemPersistTarget struct {
	TargetIndex int    `json:"targetIndex"`
	ResID       string `json:"resID"`
	ColIndex    int    `json:"colIndex"`
	ColName     string `json:"colName"`
}

// ProblemPersistPerformanceMetric captures the metrics of a problem.
type ProblemPersistPerformanceMetric struct {
	Metric string `json:"metric"`
}

// ProblemPersistExpectedOutput represents the expected output of a problem.
type ProblemPersistExpectedOutput struct {
	PredictionsFile string `json:"predictionsFile"`
}

// DefaultMetrics returns default metric for a given task.
func DefaultMetrics(taskKeywords []string) []string {
	return compute.GetDefaultTaskMetricsTA3(taskKeywords)
}

// DefaultTaskType returns a default task.
func DefaultTaskType(targetType string, problemType string) []string {
	if problemType == compute.ForecastingTask {
		return []string{compute.ForecastingTask, compute.TimeSeriesTask}
	} else if model.IsCategorical(targetType) {
		return []string{compute.ClassificationTask, compute.MultiClassTask}
	}
	return []string{compute.RegressionTask, compute.UnivariateTask}
}

// CreateProblemSchema captures the problem information in the required D3M
// problem format.
func CreateProblemSchema(datasetDir string, dataset string, targetVar *model.Variable, filters *api.FilterParams) (*ProblemPersist, string, error) {
	// parse the dataset and its filter state and generate a hashcode from both
	hash, err := getFilteredDatasetHash(dataset, targetVar.Key, filters, true)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to build dataset filter hash")
	}
	problemIDHash := fmt.Sprintf("p%s", strconv.FormatUint(hash, 10))

	// check to see if we already have this problem saved - return the path
	// if so
	pPath := path.Join(datasetDir, problemIDHash)
	pFilePath := path.Join(pPath, D3MProblem)
	if pathExists(pPath) && util.FileExists(pFilePath) {
		log.Infof("Found stored problem for %s with hash %s", dataset, problemIDHash)
		return nil, pPath, nil
	}

	taskKeywords := DefaultTaskType(targetVar.Type, "")
	metrics := DefaultMetrics(taskKeywords)

	pTarget := &ProblemPersistTarget{
		TargetIndex: 0,
		ResID:       compute.DefaultResourceID,
		ColIndex:    targetVar.Index,
		ColName:     targetVar.DisplayName,
	}

	pMetric := &ProblemPersistPerformanceMetric{
		Metric: metrics[0],
	}

	pData := &ProblemPersistData{
		DatasetID: dataset,
		Targets:   []*ProblemPersistTarget{pTarget},
	}

	pInput := &ProblemPersistInput{
		Data:               []*ProblemPersistData{pData},
		PerformanceMetrics: []*ProblemPersistPerformanceMetric{pMetric},
	}

	problemID := strings.Replace(dataset, "_dataset", "", -1)
	problemID = fmt.Sprintf("%s%s", problemID, "_problem")
	pProps := &ProblemPersistAbout{
		ProblemID:            problemID,
		ProblemVersion:       problemVersion,
		ProblemSchemaVersion: problemSchemaVersion,
		TaskKeywords:         taskKeywords,
	}

	problem := &ProblemPersist{
		About:  pProps,
		Inputs: pInput,
	}

	return problem, problemIDHash, nil
}

// LoadProblemSchemaFromFile loads the problem schema from file.
func LoadProblemSchemaFromFile(filename string) (*ProblemPersist, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	problemInfo := &ProblemPersist{}
	err = json.Unmarshal(b, problemInfo)
	if err != nil {
		return nil, err
	}
	return problemInfo, nil
}
