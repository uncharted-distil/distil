package compute

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-compute/model"
	api "github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/plog"
)

const (
	// D3MProblem name of the expected problem file.
	D3MProblem = "problemDoc.json"

	problemVersion       = "1.0"
	problemSchemaVersion = "3.0"

	defaultNumericalMetric   = "rSquared"
	defaultCategoricalMetric = "f1Micro"

	defaultTaskTypeNumerical   = "regression"
	defaultTaskTypeCategorical = "classification"

	defaultTaskSubTypeNumerical   = "univariate"
	defaultTaskSubTypeCategorical = "multiClass"
)

// VariableProvider defines a function that will get the provided variable.
type VariableProvider func(dataset string, index string, name string) (*model.Variable, error)

// ProblemPersist contains the problem file data.
type ProblemPersist struct {
	About           *ProblemPersistAbout          `json:"about"`
	Inputs          *ProblemPersistInput          `json:"inputs"`
	ExpectedOutputs *ProblemPersistExpectedOutput `json:"expectedOutputs"`
}

// ProblemPersistAbout represents the basic information of a problem.
type ProblemPersistAbout struct {
	ProblemID            string `json:"problemID"`
	ProblemName          string `json:"problemName"`
	TaskType             string `json:"taskType"`
	TaskSubType          string `json:"taskSubType"`
	ProblemVersion       string `json:"problemVersion"`
	ProblemSchemaVersion string `json:"problemSchemaVersion"`
}

// ProblemPersistInput lists the information of a problem.
type ProblemPersistInput struct {
	Data               []*ProblemPersistData              `json:"data"`
	PerformanceMetrics []*ProblemPersistPerformanceMetric `json:"performanceMetrics"`
	DataSplits         *ProblemPersistDataSplits          `json:"dataSplits"`
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

// DefaultMetrics returns default metrics.
func DefaultMetrics(targetType string) []string {
	if model.IsCategorical(targetType) {
		return []string{defaultCategoricalMetric}
	}
	return []string{defaultNumericalMetric}
}

// DefaultTaskType returns a default task.
func DefaultTaskType(targetType string) string {
	if model.IsCategorical(targetType) {
		return defaultTaskTypeCategorical
	}
	return defaultTaskTypeNumerical
}

// DefaultTaskSubType returns a default sub task.
func DefaultTaskSubType(targetType string) string {
	if model.IsCategorical(targetType) {
		return defaultTaskSubTypeCategorical
	}
	return defaultTaskSubTypeNumerical
}

// CreateProblemSchema captures the problem information in the required D3M
// problem format.
func CreateProblemSchema(datasetDir string, dataset string, targetVar *model.Variable, filters *api.FilterParams) (*ProblemPersist, string, error) {
	// parse the dataset and its filter state and generate a hashcode from both
	hash, err := getFilteredDatasetHash(dataset, targetVar.Name, filters, true)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to build dataset filter hash")
	}
	problemIDHash := fmt.Sprintf("p%s", strconv.FormatUint(hash, 10))

	// check to see if we already have this problem saved - return the path
	// if so
	pPath := path.Join(datasetDir, problemIDHash)
	pFilePath := path.Join(pPath, D3MProblem)
	if dirExists(pPath) && fileExists(pFilePath) {
		log.Infof("Found stored problem for %s with hash %d", dataset, problemIDHash)
		return nil, pPath, nil
	}

	metrics := DefaultMetrics(targetVar.Type)

	targetIdx := -1

	pTarget := &ProblemPersistTarget{
		TargetIndex: 0,
		ResID:       "0",
		ColIndex:    targetIdx,
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
		TaskType:             DefaultTaskType(targetVar.Type),
		TaskSubType:          DefaultTaskSubType(targetVar.Type),
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
