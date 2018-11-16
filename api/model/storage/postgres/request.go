package postgres

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-compute/model"
	api "github.com/unchartedsoftware/distil/api/model"
)

// PersistRequest persists a request to Postgres.
func (s *Storage) PersistRequest(requestID string, dataset string, progress string, createdTime time.Time) error {
	sql := fmt.Sprintf("INSERT INTO %s (request_id, dataset, progress, created_time, last_updated_time) VALUES ($1, $2, $3, $4, $4);", requestTableName)

	_, err := s.client.Exec(sql, requestID, dataset, progress, createdTime)

	return err
}

// UpdateRequest updates a request in Postgres.
func (s *Storage) UpdateRequest(requestID string, progress string, updatedTime time.Time) error {
	sql := fmt.Sprintf("UPDATE %s SET progress = $1, last_updated_time = $2 WHERE request_id = $3;", requestTableName)

	_, err := s.client.Exec(sql, progress, updatedTime, requestID)

	return err
}

// PersistRequestFeature persists request feature information to Postgres.
func (s *Storage) PersistRequestFeature(requestID string, featureName string, featureType string) error {
	sql := fmt.Sprintf("INSERT INTO %s (request_id, feature_name, feature_type) VALUES ($1, $2, $3);", featureTableName)

	_, err := s.client.Exec(sql, requestID, featureName, featureType)

	return err
}

// PersistRequestFilters persists request filters information to Postgres.
func (s *Storage) PersistRequestFilters(requestID string, filters *api.FilterParams) error {
	sql := fmt.Sprintf("INSERT INTO %s (request_id, feature_name, filter_type, filter_mode, filter_min, filter_max, filter_categories, filter_indices) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);", filterTableName)

	for _, filter := range filters.Filters {
		switch filter.Type {
		case model.NumericalFilter:
			_, err := s.client.Exec(sql, requestID, filter.Key, model.NumericalFilter, filter.Mode, filter.Min, filter.Max, "", "")
			if err != nil {
				return err
			}
		case model.CategoricalFilter, model.FeatureFilter, model.TextFilter:
			_, err := s.client.Exec(sql, requestID, filter.Key, filter.Type, filter.Mode, 0, 0, strings.Join(filter.Categories, ","), "")
			if err != nil {
				return err
			}
		case model.RowFilter:
			_, err := s.client.Exec(sql, requestID, "", model.RowFilter, filter.Mode, 0, 0, "", strings.Join(filter.D3mIndices, ","))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// FetchRequest pulls request information from Postgres.
func (s *Storage) FetchRequest(requestID string) (*api.Request, error) {
	sql := fmt.Sprintf("SELECT request_id, dataset, progress, created_time, last_updated_time FROM %s WHERE request_id = $1 ORDER BY created_time desc LIMIT 1;", requestTableName)

	rows, err := s.client.Query(sql, requestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}
	rows.Next()

	return s.loadRequest(rows)
}

// FetchRequestBySolutionID pulls request information from Postgres using
// a solution ID.
func (s *Storage) FetchRequestBySolutionID(solutionID string) (*api.Request, error) {
	sql := fmt.Sprintf("SELECT req.request_id, req.dataset, req.progress, req.created_time, req.last_updated_time "+
		"FROM %s as req INNER JOIN %s as sol ON req.request_id = sol.request_id "+
		"WHERE sol.solution_id = $1;", requestTableName, solutionTableName)

	rows, err := s.client.Query(sql, solutionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}
	rows.Next()

	return s.loadRequest(rows)
}

func (s *Storage) loadRequest(rows *pgx.Rows) (*api.Request, error) {
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
	sql := fmt.Sprintf("SELECT request_id, feature_name, feature_type FROM %s WHERE request_id = $1;", featureTableName)

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

	return results, nil
}

// FetchRequestFilters pulls request filter information from Postgres.
func (s *Storage) FetchRequestFilters(requestID string, features []*api.Feature) (*api.FilterParams, error) {
	sql := fmt.Sprintf("SELECT request_id, feature_name, filter_type, filter_mode, filter_min, filter_max, filter_categories, filter_indices FROM %s WHERE request_id = $1;", filterTableName)

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
		var filterCategories string
		var filterIndices string

		err = rows.Scan(&requestID, &featureName, &filterType, &filterMode, &filterMin, &filterMax, &filterCategories, &filterIndices)
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
		case model.FeatureFilter:
			filters.Filters = append(filters.Filters, model.NewFeatureFilter(
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
		case model.RowFilter:
			filters.Filters = append(filters.Filters, model.NewRowFilter(
				filterMode,
				strings.Split(filterIndices, ","),
			))
		}
	}

	for _, feature := range features {
		filters.Variables = append(filters.Variables, feature.FeatureName)
	}

	return filters, nil
}

func (s *Storage) loadRequestFromSolutionID(solutionID string) (*api.Request, error) {
	solution, err := s.FetchSolution(solutionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to fetch solution from Postgres")
	}

	request, err := s.FetchRequest(solution.RequestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to fetch request from Postgres")
	}
	request.Solutions = []*api.Solution{solution}

	return request, nil
}

// FetchRequestByDatasetTarget pulls a request with solution
// result information from Postgres. Only the latest result for each
// solution is fetched.
func (s *Storage) FetchRequestByDatasetTarget(dataset string, target string, solutionID string) ([]*api.Request, error) {

	// get the solution ids
	sql := fmt.Sprintf("SELECT DISTINCT solution.solution_id "+
		"FROM %s request INNER JOIN %s rf ON request.request_id = rf.request_id "+
		"INNER JOIN %s solution ON request.request_id = solution.request_id ",
		requestTableName, featureTableName, solutionTableName)
	params := make([]interface{}, 0)

	if dataset != "" {
		sql = fmt.Sprintf("%s AND request.dataset = $%d", sql, len(params)+1)
		params = append(params, dataset)
	}
	if target != "" {
		sql = fmt.Sprintf("%s AND rf.feature_name = $%d AND rf.feature_type = $%d", sql, len(params)+1, len(params)+2)
		params = append(params, target)
		params = append(params, model.FeatureTypeTarget)
	}
	if solutionID != "" {
		sql = fmt.Sprintf("%s AND solution.solution_id = $%d", sql, len(params)+1)
		params = append(params, solutionID)
	}

	sql = fmt.Sprintf("%s;", sql)
	rows, err := s.client.Query(sql, params...)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull solution ids from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	// TODO: code should be changed to not have a request / result built.
	// Would need to lookup to see if the request was already loaded.
	// Then would need to see if the solution was loaded.
	requests := make([]*api.Request, 0)
	for rows.Next() {
		var solutionID string

		err = rows.Scan(&solutionID)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse solution id from Postgres")
		}

		request, err := s.loadRequestFromSolutionID(solutionID)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to load request from Postgres")
		}

		result, err := s.FetchSolutionResult(solutionID)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse solution result from Postgres")
		}

		if result != nil {
			request.Solutions[0].Result = result
		}

		scores, err := s.FetchSolutionScores(solutionID)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse solution result from Postgres")
		}

		if scores != nil {
			request.Solutions[0].Scores = scores
		}

		requests = append(requests, request)
	}

	return requests, nil
}
