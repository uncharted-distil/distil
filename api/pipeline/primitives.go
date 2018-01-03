package pipeline

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-ingest/rest"
	"github.com/unchartedsoftware/distil/api/model"
)

const (
	pcaFolder           = "pca"
	rfFolder            = "rf"
	rankingFunctionName = "pca"
)

// ImportanceRankingResult is a result from variable importance ranking.
type ImportanceRankingResult struct {
	DatasetID  string                `json:"datasetID"`
	TargetName string                `json:"targetName"`
	Importance []*VariableImportance `json:"importance"`
}

// VariableImportance captures the ranking importance of a variable.
type VariableImportance struct {
	ColName       string  `json:"colName"`
	ColImportance float64 `json:"colImportance"`
}

func (c *Client) parseImportanceResult(data []byte) (*ImportanceRankingResult, error) {
	importance := &ImportanceRankingResult{}
	err := json.Unmarshal(data, importance)

	return importance, err
}

// Rank ranks the variable importance relative to a target variable.
func (c *Client) Rank(restClient *rest.Client, meta model.MetadataStorage, data model.DataStorage, dataset string, index string, targetName string) (*ImportanceRankingResult, error) {
	// check if the pca request has already been made for this target
	// folder structure is pca folder/dataset/target.csv
	datasetFolder := path.Join(c.DataDir, pcaFolder, dataset)
	err := os.MkdirAll(datasetFolder, os.ModePerm)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create dataset folder")
	}

	// check if the result file exists (target.csv)
	var importance *ImportanceRankingResult
	resultPath := path.Join(datasetFolder, fmt.Sprintf("%s.csv", targetName))
	resultData, err := ioutil.ReadFile(resultPath)
	if os.IsNotExist(err) {
		// pull the dataset
		rawData, err := data.FetchData(dataset, index, &model.FilterParams{}, true)
		if err != nil {
			return nil, errors.Wrap(err, "unable to pull the data")
		}

		// write out the dataset to the file system
		rawFilename := fmt.Sprintf("%s_raw.csv", targetName)
		rawFilePath := path.Join(datasetFolder, rawFilename)
		err = PersistData(datasetFolder, rawFilename, rawData)
		if err != nil {
			return nil, errors.Wrap(err, "unable to store data for ranking")
		}

		// rank the persisted dataset
		ranker := rest.NewRanker(rankingFunctionName, restClient)
		rawResults, err := ranker.RankFileForTarget(rawFilePath, targetName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to rank data")
		}

		// enhance the results a bit
		importance = &ImportanceRankingResult{
			DatasetID:  dataset,
			TargetName: targetName,
			Importance: make([]*VariableImportance, len(rawResults.Features)),
		}
		for i := 0; i < len(importance.Importance); i++ {
			importance.Importance[i] = &VariableImportance{
				ColName:       rawData.Columns[i],
				ColImportance: rawResults.Features[i],
			}
		}

		// store the importance for future requests
		output, err := json.Marshal(importance)
		ioutil.WriteFile(resultPath, output, os.ModePerm)
	} else {
		// previously ranked the data so parse it
		importance, err = c.parseImportanceResult(resultData)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse existing importance ranking")
		}
	}

	return importance, nil
}
