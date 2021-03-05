//
//   Copyright © 2021 Uncharted Software Inc.
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

	elastic "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util/json"
)

const (
	modelsListSize = 1000
)

func (s *Storage) parseRawSolutionVariable(rsv map[string]interface{}) (*api.SolutionVariable, error) {
	key, ok := json.String(rsv, "key")
	if !ok {
		return nil, errors.New("unable to parse key from variable data")
	}
	displayName, ok := json.String(rsv, "displayName")
	if !ok {
		return nil, errors.New("unable to parse display name from variable data")
	}
	headerName, ok := json.String(rsv, "headerName")
	if !ok {
		return nil, errors.New("unable to parse header name from variable data")
	}

	rank, ok := json.Float(rsv, "rank")
	if !ok {
		return nil, errors.New("unable to parse rank from variable data")
	}

	typ, ok := json.String(rsv, "varType")
	if !ok {
		return nil, errors.New("unable to parse type from variable data")
	}

	return &api.SolutionVariable{
		Key:         key,
		DisplayName: displayName,
		HeaderName:  headerName,
		Rank:        rank,
		Type:        typ,
	}, nil
}

func (s *Storage) parseSolutionVariables(src map[string]interface{}) ([]*api.SolutionVariable, error) {
	rawSolutionVariables, ok := json.Array(src, "variableDetails")
	if !ok {
		return nil, errors.New("failed to parse variable list")
	}
	var solutionVariables []*api.SolutionVariable
	for _, rsv := range rawSolutionVariables {
		sv, err := s.parseRawSolutionVariable(rsv)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse variable")
		}
		if sv != nil {
			solutionVariables = append(solutionVariables, sv)
		}
	}
	return solutionVariables, nil
}

func (s *Storage) parseModels(res *elastic.SearchResult) ([]*api.ExportedModel, error) {
	var models []*api.ExportedModel
	for _, hit := range res.Hits.Hits {
		// parse hit into JSON
		src, err := json.Unmarshal(hit.Source)
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
		targetInfo, ok := json.Get(src, "target")
		if !ok {
			return nil, errors.New("failed to parse the target")
		}
		target, err := s.parseRawSolutionVariable(targetInfo)
		if err != nil {
			return nil, err
		}

		// extract the name
		name, ok := json.String(src, "datasetName")
		if !ok {
			name = datasetID
		}

		variables, _ := json.StringArray(src, "variables")

		variableDetails, err := s.parseSolutionVariables(src)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse solution variables")
		}

		// write everything out to result struct
		models = append(models, &api.ExportedModel{
			ModelName:        modelName,
			ModelDescription: modelDescription,
			FilePath:         filePath,
			FittedSolutionID: fittedSolutionID,
			DatasetID:        datasetID,
			DatasetName:      name,
			Target:           target,
			Variables:        variables,
			VariableDetails:  variableDetails,
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

// FetchModel returns a model in the provided index.  Model name is the named assigend
// to the model by the user.
func (s *Storage) FetchModel(modelName string) (*api.ExportedModel, error) {
	query := elastic.NewMatchQuery("id", modelName)
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
	if len(models) > 0 {
		return models[0], nil
	}
	return nil, nil
}

// FetchModelByID returns a model in the provided index using the model's fitted solution ID.
func (s *Storage) FetchModelByID(fittedSolutionID string) (*api.ExportedModel, error) {
	query := elastic.NewMatchQuery("_id", fittedSolutionID)
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
	if len(models) > 0 {
		return models[0], nil
	}
	return nil, nil
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
