package pipeline

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/plog"
)

const (
	// D3MProblem name of the expected problem file.
	D3MProblem           = "problemDoc.json"
	problemVersion       = "1.0"
	problemSchemaVersion = "3.0"
)

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

// ProblemInput lists the inputs of a problem.
type ProblemInput struct {
	Data *ProblemData `json:"data"`
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

// PersistProblem stores the problem information in the required D3M
// problem format.
func PersistProblem(datasetDir string, dataset string, index string, target string, filters *model.FilterParams) (string, error) {
	// parse the dataset and its filter state and generate a hashcode from both
	hash, err := getFilteredDatasetHash(dataset, target, filters)
	if err != nil {
		return "", err
	}

	// check to see if we already have this filtered dataset saved - return the path
	// if so
	pPath := path.Join(datasetDir, strconv.FormatUint(hash, 10))
	if dirExists(pPath) {
		log.Infof("Found cached data for %s with hash %d", dataset, hash)
		return pPath, nil
	}

	targetIdx := -1

	pTarget := &ProblemTarget{
		TargetIndex: 0,
		ResID:       "0",
		ColIndex:    targetIdx,
		ColName:     target,
	}

	pData := &ProblemData{
		DatasetID: dataset,
		Targets:   []*ProblemTarget{pTarget},
	}

	pInput := &ProblemInput{
		Data: pData,
	}

	pProps := &ProblemProperties{
		ProblemID:            strings.Replace(dataset, "dataset", "problem", -1),
		ProblemVersion:       problemVersion,
		ProblemSchemaVersion: problemSchemaVersion,
	}

	problem := &Problem{
		Properties: pProps,
		Inputs:     []*ProblemInput{pInput},
	}

	data, err := json.Marshal(problem)
	if err != nil {
		return "", errors.Wrap(err, "unable to marshal problem data")
	}

	err = ioutil.WriteFile(path.Join(pPath, D3MProblem), data, 0644)
	if err != nil {
		return "", errors.Wrap(err, "Unable to write problem data")
	}

	return pPath, nil
}
