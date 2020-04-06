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

package elastic

import (
	"context"

	"github.com/pkg/errors"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util/json"
	elastic "gopkg.in/olivere/elastic.v5"
)

const (
	modelsListSize = 1000
)

func (s *Storage) parseModels(res *elastic.SearchResult) ([]*api.ExportedModel, error) {
	var models []*api.ExportedModel
	for _, hit := range res.Hits.Hits {
		// parse hit into JSON
		src, err := json.Unmarshal(*hit.Source)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse model")
		}
		// extract the model name
		modelName, ok := json.String(src, "modelName")
		if !ok {
			return nil, errors.New("failed to parse model name")
		}
		// extract the model description
		modelDescription, ok := json.String(src, "modelDescription")
		if !ok {
			return nil, errors.New("failed to parse model description")
		}
		// extract the file path
		filePath, ok := json.String(src, "filePath")
		if !ok {
			return nil, errors.New("failed to parse file path")
		}
		// get id
		fittedSolutionID, ok := json.String(src, "fittedSolutionId")
		if !ok {
			return nil, errors.New("failed to parse the fitted solution id")
		}
		// get dataset id
		datasetID, ok := json.String(src, "datasetId")
		if !ok {
			return nil, errors.New("failed to parse the dataset id")
		}
		// extract the target
		target, ok := json.String(src, "target")
		if !ok {
			return nil, errors.New("failed to parse the target")
		}
		// extract the name
		name, ok := json.String(src, "datasetName")
		if !ok {
			name = datasetID
		}

		variables, _ := json.StringArray(src, "variables")

		// write everythign out to result struct
		models = append(models, &api.ExportedModel{
			ModelName:        modelName,
			ModelDescription: modelDescription,
			FilePath:         filePath,
			FittedSolutionID: fittedSolutionID,
			DatasetID:        datasetID,
			DatasetName:      name,
			Target:           target,
			Variables:        variables,
		})
	}
	return models, nil
}

// PersistExportedModel writes an exported model to ES storage.
func (s *Storage) PersistExportedModel(model *api.ExportedModel) error {
	bytes, err := json.Marshal(model)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal model")
	}

	// push the document into the model index
	_, err = s.client.Index().
		Index(s.modelIndex).
		Type("model").
		Id(model.FittedSolutionID).
		BodyString(string(bytes)).
		Refresh("true").
		Do(context.Background())
	if err != nil {
		return errors.Wrapf(err, "failed to add document to index `%s`", s.modelIndex)
	}
	return nil
}

// FetchModels returns all exported models  in the provided index.
func (s *Storage) FetchModels() ([]*api.ExportedModel, error) {
	// execute the ES query
	res, err := s.client.Search().
		Index(s.modelIndex).
		FetchSource(true).
		Size(modelsListSize).
		Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "elasticsearch model fetch query failed")
	}
	return s.parseModels(res)
}

// FetchModel returns a model in the provided index.
func (s *Storage) FetchModel(modelName string) (*api.ExportedModel, error) {
	query := elastic.NewMatchQuery("modelName", modelName)
	// execute the ES query
	res, err := s.client.Search().
		Query(query).
		Index(s.modelIndex).
		FetchSource(true).
		Size(modelsListSize).
		Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "elasticsearch model fetch query failed")
	}
	models, err := s.parseModels(res)
	if err != nil {
		return nil, err
	}
	return models[0], nil
}

// SearchModels returns the models that match the search criteria in the
// provided index.
func (s *Storage) SearchModels(terms string) ([]*api.ExportedModel, error) {
	query := elastic.NewMultiMatchQuery(terms, "_id", "modelName", "modelDescription", "datasetId", "datasetName", "target", "variables").
		Analyzer("standard")
	// execute the ES query
	res, err := s.client.Search().
		Query(query).
		Index(s.modelIndex).
		FetchSource(true).
		Size(modelsListSize).
		Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "elasticsearch model search query failed")
	}
	return s.parseModels(res)
}
