package compute

import (
	"fmt"
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

	numericalMetric   = "r_squared"
	categoricalMetric = "accuracy"

	problemTaskTypeNumerical   = "regression"
	problemTaskTypeCategorical = "classification"

	problemTaskSubTypeNumerical   = "univariate"
	problemTaskSubTypeCategorical = "multiClass"
)

// VariableProvider defines a function that will get the provided variable.
type VariableProvider func(dataset string, index string, name string) (*model.Variable, error)

// ProblemPersist contains the problem file data.
type ProblemPersist struct {
	Properties *ProblemPersistProperties `json:"about"`
	Inputs     *ProblemPersistInput      `json:"inputs"`
}

// ProblemPersistProperties represents the basic information of a problem.
type ProblemPersistProperties struct {
	ProblemID            string `json:"problemID"`
	TaskType             string `json:"taskType"`
	TaskSubType          string `json:"taskSubType"`
	ProblemVersion       string `json:"problemVersion"`
	ProblemSchemaVersion string `json:"problemSchemaVersion"`
}

// ProblemPersistInput lists the information of a problem.
type ProblemPersistInput struct {
	Data               *ProblemPersistData                `json:"data"`
	PerformanceMetrics []*ProblemPersistPerformanceMetric `json:"performanceMetrics"`
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

func getTaskSubType(targetType string) string {
	if model.IsCategorical(targetType) {
		return problemTaskSubTypeCategorical
	}
	return problemTaskSubTypeNumerical
}

// CreateProblemSchema captures the problem information in the required D3M
// problem format.
func CreateProblemSchema(datasetDir string, dataset string, targetVar *model.Variable, filters *model.FilterParams) (*ProblemPersist, string, error) {
	// parse the dataset and its filter state and generate a hashcode from both
	hash, err := getFilteredDatasetHash(dataset, targetVar.Name, filters, true)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to build dataset filter hash")
	}

	// check to see if we already have this problem saved - return the path
	// if so
	pPath := path.Join(datasetDir, strconv.FormatUint(hash, 10))
	pFilePath := path.Join(pPath, D3MProblem)
	if dirExists(pPath) && fileExists(pFilePath) {
		log.Infof("Found stored problem for %s with hash %d", dataset, hash)
		return nil, pPath, nil
	}

	metric := getMetric(targetVar.Type)

	targetIdx := -1

	pTarget := &ProblemPersistTarget{
		TargetIndex: 0,
		ResID:       "0",
		ColIndex:    targetIdx,
		ColName:     targetVar.DisplayVariable,
	}

	pMetric := &ProblemPersistPerformanceMetric{
		Metric: metric,
	}

	pData := &ProblemPersistData{
		DatasetID: dataset,
		Targets:   []*ProblemPersistTarget{pTarget},
	}

	pInput := &ProblemPersistInput{
		Data:               pData,
		PerformanceMetrics: []*ProblemPersistPerformanceMetric{pMetric},
	}

	problemID := strings.Replace(dataset, "_dataset", "", -1)
	problemID = fmt.Sprintf("%s%s", problemID, "_problem")
	pProps := &ProblemPersistProperties{
		ProblemID:            problemID,
		ProblemVersion:       problemVersion,
		ProblemSchemaVersion: problemSchemaVersion,
		TaskType:             getTaskType(targetVar.Type),
		TaskSubType:          getTaskSubType(targetVar.Type),
	}

	problem := &ProblemPersist{
		Properties: pProps,
		Inputs:     pInput,
	}

	return problem, strconv.FormatUint(hash, 10), nil
}
