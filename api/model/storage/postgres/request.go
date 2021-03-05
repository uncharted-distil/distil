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

package postgres

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	postgres "github.com/uncharted-distil/distil/api/postgres"
)

// PersistRequest persists a request to Postgres.
func (s *Storage) PersistRequest(requestID string, dataset string, progress string, createdTime time.Time) error {
	sql := fmt.Sprintf("INSERT INTO %s (request_id, dataset, progress, created_time, last_updated_time) VALUES ($1, $2, $3, $4, $4);", postgres.RequestTableName)

	_, err := s.client.Exec(sql, requestID, dataset, progress, createdTime)

	return errors.Wrapf(err, "failed to persist request to PostGres")
}

// UpdateRequest updates a request in Postgres.
func (s *Storage) UpdateRequest(requestID string, progress string, updatedTime time.Time) error {
	sql := fmt.Sprintf("UPDATE %s SET progress = $1, last_updated_time = $2 WHERE request_id = $3;", postgres.RequestTableName)

	_, err := s.client.Exec(sql, progress, updatedTime, requestID)

	return errors.Wrapf(err, "failed to update request in PostGres")
}

// PersistRequestFeature persists request feature information to Postgres.
func (s *Storage) PersistRequestFeature(requestID string, featureName string, featureType string) error {
	sql := fmt.Sprintf("INSERT INTO %s (request_id, feature_name, feature_type) VALUES ($1, $2, $3);", postgres.RequestFeatureTableName)

	_, err := s.client.Exec(sql, requestID, featureName, featureType)

	return errors.Wrapf(err, "failed to persist request freature in PostGres")
}

// PersistRequestFilters persists request filters information to Postgres.
func (s *Storage) PersistRequestFilters(requestID string, filters *api.FilterParams) error {
	sql := fmt.Sprintf(
		"INSERT INTO %s (request_id, feature_name, filter_type, filter_mode, filter_min, filter_max, filter_min_x, filter_max_x, filter_min_y, filter_max_y, filter_categories, filter_indices) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);", postgres.RequestFilterTableName)

	for _, filter := range filters.Filters {
		switch filter.Type {
		case model.NumericalFilter:
			_, err := s.client.Exec(sql, requestID, filter.Key, model.NumericalFilter, filter.Mode, filter.Min, filter.Max, 0, 0, 0, 0, "", "")
			if err != nil {
				return errors.Wrap(err, "failed to persist numerical filter")
			}
		case model.BivariateFilter, model.GeoBoundsFilter:
			_, err := s.client.Exec(sql, requestID, filter.Key, filter.Type, filter.Mode, 0, 0, filter.Bounds.MinX, filter.Bounds.MaxX, filter.Bounds.MinY, filter.Bounds.MaxY, "", "")
			if err != nil {
				return errors.Wrap(err, "failed to persist bivariate filter")
			}
		case model.CategoricalFilter, model.TextFilter:
			_, err := s.client.Exec(sql, requestID, filter.Key, filter.Type, filter.Mode, 0, 0, 0, 0, 0, 0, strings.Join(filter.Categories, ","), "")
			if err != nil {
				return errors.Wrap(err, "failed to persist categorical filter")
			}
		case model.RowFilter:
			_, err := s.client.Exec(sql, requestID, "", model.RowFilter, filter.Mode, 0, 0, 0, 0, 0, 0, "", strings.Join(filter.D3mIndices, ","))
			if err != nil {
				return errors.Wrap(err, "failed to persist row filter")
			}
		}
	}
	return nil
}

// FetchRequest pulls request information from Postgres.
func (s *Storage) FetchRequest(requestID string) (*api.Request, error) {
	sql := fmt.Sprintf("SELECT request_id, dataset, progress, created_time, last_updated_time FROM %s WHERE request_id = $1 ORDER BY created_time desc LIMIT 1;", postgres.RequestTableName)

	rows, err := s.client.Query(sql, requestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}
	rows.Next()
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	return s.loadRequest(rows)
}

// FetchRequestBySolutionID pulls request information from Postgres using
// a solution ID.
func (s *Storage) FetchRequestBySolutionID(solutionID string) (*api.Request, error) {
	sql := fmt.Sprintf("SELECT req.request_id, req.dataset, req.progress, req.created_time, req.last_updated_time "+
		"FROM %s as req INNER JOIN %s as sol ON req.request_id = sol.request_id "+
		"WHERE sol.solution_id = $1;", postgres.RequestTableName, postgres.SolutionTableName)

	rows, err := s.client.Query(sql, solutionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}
	if !rows.Next() {
		return nil, errors.Errorf("no request for solution %s", solutionID)
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	return s.loadRequest(rows)
}

// FetchRequestByFittedSolutionID pulls request information from Postgres using
// a fitted solution ID.
func (s *Storage) FetchRequestByFittedSolutionID(fittedSolutionID string) (*api.Request, error) {
	sql := fmt.Sprintf("SELECT req.request_id, req.dataset, req.progress, req.created_time, req.last_updated_time "+
		"FROM %s as req INNER JOIN %s as sol ON req.request_id = sol.request_id INNER JOIN %s sr on sr.solution_id = sol.solution_id "+
		"WHERE sr.fitted_solution_id = $1;", postgres.RequestTableName, postgres.SolutionTableName, postgres.SolutionResultTableName)

	rows, err := s.client.Query(sql, fittedSolutionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}
	if !rows.Next() {
		return nil, errors.Errorf("no request for solution %s", fittedSolutionID)
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	return s.loadRequest(rows)
}

func (s *Storage) loadRequest(rows pgx.Rows) (*api.Request, error) {
	var requestID string
	var dataset string
	var progress string
	var createdTime time.Time
	var lastUpdatedTime time.Time

	err := rows.Scan(&requestID, &dataset, &progress, &createdTime, &lastUpdatedTime)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse request from Postgres")
	}

	features, err := s.FetchRequestFeatures(requestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get request features from Postgres")
	}

	filters, err := s.FetchRequestFilters(requestID, features)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get request filters from Postgres")
	}

	return &api.Request{
		RequestID:       requestID,
		Dataset:         dataset,
		Progress:        progress,
		CreatedTime:     createdTime,
		LastUpdatedTime: lastUpdatedTime,
		Features:        features,
		Filters:         filters,
	}, nil
}

// FetchRequestFeatures pulls request feature information from Postgres.
func (s *Storage) FetchRequestFeatures(requestID string) ([]*api.Feature, error) {
	sql := fmt.Sprintf("SELECT request_id, feature_name, feature_type FROM %s WHERE request_id = $1;", postgres.RequestFeatureTableName)

	rows, err := s.client.Query(sql, requestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request features from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results := make([]*api.Feature, 0)
	for rows.Next() {
		var requestID string
		var featureName string
		var featureType string

		err = rows.Scan(&requestID, &featureName, &featureType)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse request features from Postgres")
		}

		results = append(results, &api.Feature{
			RequestID:   requestID,
			FeatureName: featureName,
			FeatureType: featureType,
		})
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	return results, nil
}

// FetchRequestFilters pulls request filter information from Postgres.
func (s *Storage) FetchRequestFilters(requestID string, features []*api.Feature) (*api.FilterParams, error) {
	sql := fmt.Sprintf("SELECT request_id, feature_name, filter_type, filter_mode, filter_min, filter_max, filter_min_x, filter_max_x, filter_min_y, filter_max_y, filter_categories, filter_indices FROM %s WHERE request_id = $1;", postgres.RequestFilterTableName)

	rows, err := s.client.Query(sql, requestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request filters from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	filters := &api.FilterParams{
		Size: model.DefaultFilterSize,
	}

	for rows.Next() {
		var requestID string
		var featureName string
		var filterType string
		var filterMode string
		var filterMin float64
		var filterMax float64
		var filterMinX float64
		var filterMaxX float64
		var filterMinY float64
		var filterMaxY float64
		var filterCategories string
		var filterIndices string

		err = rows.Scan(&requestID, &featureName, &filterType, &filterMode, &filterMin, &filterMax, &filterMinX, &filterMaxX, &filterMinY, &filterMaxY, &filterCategories, &filterIndices)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse request filters from Postgres")
		}

		switch filterType {
		case model.CategoricalFilter:
			filters.Filters = append(filters.Filters, model.NewCategoricalFilter(
				featureName,
				filterMode,
				strings.Split(filterCategories, ","),
			))
		case model.TextFilter:
			filters.Filters = append(filters.Filters, model.NewTextFilter(
				featureName,
				filterMode,
				strings.Split(filterCategories, ","),
			))
		case model.NumericalFilter:
			filters.Filters = append(filters.Filters, model.NewNumericalFilter(
				featureName,
				filterMode,
				filterMin,
				filterMax,
			))
		case model.BivariateFilter:
			filters.Filters = append(filters.Filters, model.NewBivariateFilter(
				featureName,
				filterMode,
				filterMinX,
				filterMaxX,
				filterMinY,
				filterMaxY,
			))
		case model.GeoBoundsFilter:
			filters.Filters = append(filters.Filters, model.NewGeoBoundsFilter(
				featureName,
				filterMode,
				filterMinX,
				filterMaxX,
				filterMinY,
				filterMaxY,
			))
		case model.RowFilter:
			filters.Filters = append(filters.Filters, model.NewRowFilter(
				filterMode,
				strings.Split(filterIndices, ","),
			))
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	for _, feature := range features {
		filters.Variables = append(filters.Variables, feature.FeatureName)
	}

	return filters, nil
}

// FetchRequestByDatasetTarget pulls requests associated with a given dataset and target from postgres.
func (s *Storage) FetchRequestByDatasetTarget(dataset string, target string) ([]*api.Request, error) {
	// get the solution ids
	sql := fmt.Sprintf("SELECT DISTINCT ON(request.request_id) request.request_id, request.dataset, request.progress, request.created_time, request.last_updated_time "+
		"FROM %s request INNER JOIN %s rf ON request.request_id = rf.request_id "+
		"INNER JOIN %s solution ON request.request_id = solution.request_id",
		postgres.RequestTableName, postgres.RequestFeatureTableName, postgres.SolutionTableName)
	params := make([]interface{}, 0)

	if dataset != "" {
		sql = fmt.Sprintf("%s AND request.dataset = $%d", sql, len(params)+1)
		params = append(params, dataset)
	}
	if target != "" {
		sql = fmt.Sprintf("%s AND rf.feature_name = $%d AND rf.feature_type = $%d ORDER BY request.request_id, request.last_updated_time DESC",
			sql, len(params)+1, len(params)+2)
		params = append(params, target)
		params = append(params, model.FeatureTypeTarget)
	}

	sql = fmt.Sprintf("%s;", sql)
	rows, err := s.client.Query(sql, params...)
	if err != nil {
		return nil, err
	}
	if rows != nil {
		defer rows.Close()
	}

	requests := []*api.Request{}
	for rows.Next() {
		request, err := s.loadRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}
	return requests, nil
}
