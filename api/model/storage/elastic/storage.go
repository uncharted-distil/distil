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
	"fmt"

	elastic "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	es "github.com/uncharted-distil/distil/api/elastic"
	"github.com/uncharted-distil/distil/api/model"
)

// Storage accesses the underlying ES instance.
type Storage struct {
	client       *elastic.Client
	datasetIndex string
	modelIndex   string
}

// NewMetadataStorage returns a constructor for a metadata storage.
func NewMetadataStorage(datasetIndex string, initialize bool, clientCtor es.ClientCtor) model.MetadataStorageCtor {
	return func() (model.MetadataStorage, error) {
		esClient, err := clientCtor()
		if err != nil {
			return nil, err
		}

		storage := &Storage{
			client:       esClient,
			datasetIndex: datasetIndex,
		}

		if initialize {
			err = storage.InitializeMetadataStorage(false)
			if err != nil {
				return nil, err
			}
		}

		return storage, nil
	}
}

// NewExportedModelStorage returns a constructor for an exported model storage.
func NewExportedModelStorage(modelIndex string, initialize bool, clientCtor es.ClientCtor) model.ExportedModelStorageCtor {
	return func() (model.ExportedModelStorage, error) {
		esClient, err := clientCtor()
		if err != nil {
			return nil, err
		}

		storage := &Storage{
			client:     esClient,
			modelIndex: modelIndex,
		}

		if initialize {
			err = storage.InitializeModelStorage(false)
			if err != nil {
				return nil, err
			}
		}

		return storage, nil
	}
}

// InitializeMetadataStorage creates a new ElasticSearch index with our target
// mappings. An ngram analyze is defined and applied to the variable names to
// allow for substring searching.
func (s *Storage) InitializeMetadataStorage(overwrite bool) error {
	// check if it already exists
	exists, err := s.client.IndexExists(s.datasetIndex).Do(context.Background())
	if err != nil {
		return errors.Wrapf(err, "failed to complete check for existence of index %s", s.datasetIndex)
	}

	// delete the index if it already exists
	if exists {
		if overwrite {
			deleted, err := s.client.
				DeleteIndex(s.datasetIndex).
				Do(context.Background())
			if err != nil {
				return errors.Wrapf(err, "failed to delete index %s", s.datasetIndex)
			}
			if !deleted.Acknowledged {
				return fmt.Errorf("failed to create index `%s`, index could not be deleted", s.datasetIndex)
			}
		} else {
			return nil
		}
	}

	// create body
	body := `{
		"settings": {
			"max_ngram_diff": 20,
			"analysis": {
				"filter": {
					"ngram_filter": {
						"type": "ngram",
						"min_gram": 4,
						"max_gram": 20
					},
					"search_filter": {
						"type": "edge_ngram",
						"min_gram": 1,
						"max_gram": 20
					}
				},
				"tokenizer": {
					"search_tokenizer": {
						"type": "edge_ngram",
						"min_gram": 1,
						"max_gram": 20,
						"token_chars": [
							"letter",
							"digit"
						]
					}
				},
				"analyzer": {
					"ngram_analyzer": {
						"type": "custom",
						"tokenizer": "standard",
						"filter": [
							"lowercase",
							"ngram_filter"
						]
					},
					"search_analyzer": {
						"type": "custom",
						"tokenizer": "search_tokenizer",
						"filter": [
							"lowercase",
							"search_filter"
						]
					},
					"id_analyzer": {
						"type":	  "pattern",
						"pattern":   "\\W|_",
						"lowercase": true
					}
				}
			}
		},
		"mappings": {
			"properties": {
				"datasetID": {
					"type": "text",
					"analyzer": "search_analyzer"
				},
				"datasetName": {
					"type": "text",
					"analyzer": "search_analyzer",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"parentDatasetIDs": {
					"type": "text",
					"analyzer": "search_analyzer"
				},
				"storageName": {
					"type": "text"
				},
				"datasetFolder": {
					"type": "text"
				},
				"description": {
					"type": "text",
					"analyzer": "search_analyzer"
				},
				"summary": {
					"type": "text",
					"analyzer": "search_analyzer"
				},
				"summaryMachine": {
					"type": "text",
					"analyzer": "search_analyzer"
				},
				"numRows": {
					"type": "long"
				},
				"numBytes": {
					"type": "long"
				},
				"learningDataset": {
					"type": "text"
				},
				"clone": {
					"type": "boolean"
				},
				"immutable": {
					"type": "boolean"
				},
				"variables": {
					"properties": {
						"varDescription": {
							"type": "text"
						},
						"varName": {
							"type": "text",
							"analyzer": "search_analyzer",
							"term_vector": "yes"
						},
						"colName": {
							"type": "text",
							"analyzer": "search_analyzer",
							"term_vector": "yes"
						},
						"varRole": {
							"type": "text"
						},
						"varType": {
							"type": "text"
						},
						"varOriginalType": {
							"type": "text"
						},
						"varOriginalName": {
							"type": "text"
						},
						"varDisplayName": {
							"type": "text"
						},
						"importance": {
							"type": "integer"
						},
						"immutable": {
							"type": "boolean"
						}
					}
				}
			}
		}
	}`

	// create index
	created, err := s.client.
		CreateIndex(s.datasetIndex).
		BodyString(body).
		Do(context.Background())
	if err != nil {
		return errors.Wrapf(err, "failed to create index %s", s.datasetIndex)
	}
	if !created.Acknowledged {
		return fmt.Errorf("Failed to create new index %s", s.datasetIndex)
	}
	return nil
}

// InitializeModelStorage creates a new ElasticSearch index for the models.
func (s *Storage) InitializeModelStorage(overwrite bool) error {
	// check if it already exists
	exists, err := s.client.IndexExists(s.modelIndex).Do(context.Background())
	if err != nil {
		return errors.Wrapf(err, "failed to complete check for existence of index %s", s.modelIndex)
	}

	// delete the index if it already exists
	if exists {
		if overwrite {
			deleted, err := s.client.
				DeleteIndex(s.modelIndex).
				Do(context.Background())
			if err != nil {
				return errors.Wrapf(err, "failed to delete index %s", s.modelIndex)
			}
			if !deleted.Acknowledged {
				return fmt.Errorf("failed to create index `%s`, index could not be deleted", s.modelIndex)
			}
		} else {
			return nil
		}
	}

	// create body
	body := `{
		"settings": {
			"max_ngram_diff": 20,
			"analysis": {
				"filter": {
					"ngram_filter": {
						"type": "ngram",
						"min_gram": 4,
						"max_gram": 20
					},
					"search_filter": {
						"type": "edge_ngram",
						"min_gram": 1,
						"max_gram": 20
					}
				},
				"tokenizer": {
					"search_tokenizer": {
						"type": "edge_ngram",
						"min_gram": 1,
						"max_gram": 20,
						"token_chars": [
							"letter",
							"digit"
						]
					}
				},
				"analyzer": {
					"ngram_analyzer": {
						"type": "custom",
						"tokenizer": "standard",
						"filter": [
							"lowercase",
							"ngram_filter"
						]
					},
					"search_analyzer": {
						"type": "custom",
						"tokenizer": "search_tokenizer",
						"filter": [
							"lowercase",
							"search_filter"
						]
					},
					"id_analyzer": {
						"type":	  "pattern",
						"pattern":   "\\W|_",
						"lowercase": true
					}
				}
			}
		},
		"mappings": {
			"properties": {
				"modelName": {
					"type": "text",
					"analyzer": "search_analyzer"
				},
				"modelDescription": {
					"type": "text",
					"analyzer": "search_analyzer"
				},
				"filepath": {
					"type": "text"
				},
				"fittedSolutionId": {
					"type": "text"
				},
				"datasetId": {
					"type": "text",
					"analyzer": "search_analyzer"
				},
				"datasetName": {
					"type": "text",
					"analyzer": "search_analyzer",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"variables": {
					"type": "text",
					"analyzer": "search_analyzer",
					"term_vector": "yes"
				},
				"variableDetails": {
					"properties": {
						"name": {
							"type": "text",
							"analyzer": "search_analyzer",
							"term_vector": "yes"
						},
						"rank": {
							"type": "integer"
						},
						"varType": {
							"type": "text"
						}
					}
				}
			}
		}
	}`

	// create index
	created, err := s.client.
		CreateIndex(s.modelIndex).
		BodyString(body).
		Do(context.Background())
	if err != nil {
		return errors.Wrapf(err, "failed to create index %s", s.modelIndex)
	}
	if !created.Acknowledged {
		return fmt.Errorf("Failed to create new index %s", s.modelIndex)
	}
	return nil
}
