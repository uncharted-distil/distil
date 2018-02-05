package pipeline

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/plog"
)

const (
	// D3MProblem name of the expected problem file.
	D3MProblem = "problemDoc.json"

	problemVersion       = "1.0"
	problemSchemaVersion = "3.0"

	numericalMetric   = "rSquared"
	categoricalMetric = "accuracy"

	problemTaskTypeNumerical   = "regression"
	problemTaskTypeCategorical = "classification"
)

// VariableProvider defines a function that will get the provided variable.
type VariableProvider func(dataset string, index string, name string) (*model.Variable, error)

// Problem contains the problem file data.
type Problem struct {
	Properties *ProblemProperties `json:"about"`
	Inputs     []*ProblemInput    `json:"inputs"`
}

// ProblemProperties represents the basic information of a problem.
type ProblemProperties struct {
	ProblemID            string `json:"problemID"`
	TaskType             string `json:"taskType"`
	TaskSubType          string `json:"taskSubType"`
	ProblemVersion       string `json:"problemVersion"`
	ProblemSchemaVersion string `json:"problemSchemaVersion"`
}

// ProblemInput lists the information of a problem.
type ProblemInput struct {
	Data               *ProblemData                `json:"data"`
	PerformanceMetrics []*ProblemPerformanceMetric `json:"performanceMetrics"`
}

// ProblemData ties targets to a dataset.
type ProblemData struct {
	DatasetID string           `json:"datasetID"`
	Targets   []*ProblemTarget `json:"targets"`
}

// ProblemTarget represents the target information of the problem.
type ProblemTarget struct {
	TargetIndex int    `json:"targetIndex"`
	ResID       string `json:"resID"`
	ColIndex    int    `json:"colIndex"`
	ColName     string `json:"colName"`
}

// ProblemPerformanceMetric captures the metrics of a problem.
type ProblemPerformanceMetric struct {
	Metric string `json:"metric"`
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func getMetric(targetType string) string {
	if model.IsCategorical(targetType) {
		return categoricalMetric
	}
	return numericalMetric
}

func getTaskType(targetType string) string {
	if model.IsCategorical(targetType) {
		return problemTaskTypeCategorical
	}
	return problemTaskTypeNumerical
}

// PersistProblem stores the problem information in the required D3M
// problem format.
func PersistProblem(fetchVariable VariableProvider, datasetDir string, dataset string, index string, target string, filters *model.FilterParams) (string, error) {
	// parse the dataset and its filter state and generate a hashcode from both
	hash, err := getFilteredDatasetHash(dataset, target, filters)
	if err != nil {
		return "", errors.Wrap(err, "unable to build dataset filter hash")
	}

	// check to see if we already have this problem saved - return the path
	// if so
	pPath := path.Join(datasetDir, strconv.FormatUint(hash, 10))
	pFilePath := path.Join(pPath, D3MProblem)
	if dirExists(pPath) && fileExists(pFilePath) {
		log.Infof("Found stored problem for %s with hash %d", dataset, hash)
		return pPath, nil
	}

	// pull the target variable to determine the problem metric
	targetVar, err := fetchVariable(dataset, index, target)
	if err != nil {
		return "", errors.Wrap(err, "unable to pull target variable")
	}
	metric := getMetric(targetVar.Type)

	targetIdx := -1

	pTarget := &ProblemTarget{
		TargetIndex: 0,
		ResID:       "0",
		ColIndex:    targetIdx,
		ColName:     target,
	}

	pMetric := &ProblemPerformanceMetric{
		Metric: metric,
	}

	pData := &ProblemData{
		DatasetID: dataset,
		Targets:   []*ProblemTarget{pTarget},
	}

	pInput := &ProblemInput{
		Data:               pData,
		PerformanceMetrics: []*ProblemPerformanceMetric{pMetric},
	}

	problemID := strings.Replace(dataset, "_dataset", "", -1)
	problemID = fmt.Sprintf("%s%s", problemID, "_problem")
	pProps := &ProblemProperties{
		ProblemID:            problemID,
		ProblemVersion:       problemVersion,
		ProblemSchemaVersion: problemSchemaVersion,
		TaskType:             getTaskType(targetVar.Type),
	}

	problem := &Problem{
		Properties: pProps,
		Inputs:     []*ProblemInput{pInput},
	}

	data, err := json.Marshal(problem)
	if err != nil {
		return "", errors.Wrap(err, "unable to marshal problem data")
	}

	err = ioutil.WriteFile(pFilePath, data, 0644)
	if err != nil {
		return "", errors.Wrap(err, "Unable to write problem data")
	}

	return pPath, nil
}
