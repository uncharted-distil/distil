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
		// extract the file path
		filePath, ok := json.String(src, "filePath")
		if !ok {
			return nil, errors.Wrap(err, "failed to parse file path")
		}
		// get id
		id, ok := json.String(src, "datasetID")
		if !ok {
			return nil, errors.Wrap(err, "failed to parse the dataset id")
		}
		// extract the target
		target, ok := json.String(src, "target")
		if !ok {
			return nil, errors.Wrap(err, "failed to parse the target")
		}
		// extract the name
		name, ok := json.String(src, "datasetName")
		if !ok {
			name = id
		}

		variables, _ := json.StringArray(src, "variables")

		// write everythign out to result struct
		models = append(models, &api.ExportedModel{
			FilePath:    filePath,
			DatasetID:   id,
			DatasetName: name,
			Target:      target,
			Variables:   variables,
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
