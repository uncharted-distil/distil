package postgres

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
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
func (s *Storage) PersistRequestFilters(requestID string, filters *model.FilterParams) error {
	sql := fmt.Sprintf("INSERT INTO %s (request_id, feature_name, filter_type, filter_mode, filter_min, filter_max, filter_categories) VALUES ($1, $2, $3, $4, $5, $6, $7);", filterTableName)

	for _, filter := range filters.Filters {
		switch filter.Type {
		case model.NumericalFilter:
			_, err := s.client.Exec(sql, requestID, filter.Name, model.NumericalFilter, filter.Mode, filter.Min, filter.Max, "")
			if err != nil {
				return err
			}
		case model.CategoricalFilter:
			_, err := s.client.Exec(sql, requestID, filter.Name, model.CategoricalFilter, filter.Mode, 0, 0, strings.Join(filter.Categories, ","))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// FetchRequest pulls request information from Postgres.
func (s *Storage) FetchRequest(requestID string) (*model.Request, error) {
	sql := fmt.Sprintf("SELECT request_id, dataset, progress, created_time, last_updated_time FROM %s WHERE request_id = $1;", requestTableName)

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

func (s *Storage) loadRequest(rows *pgx.Rows) (*model.Request, error) {
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

	return &model.Request{
		RequestID:         requestID,
		Dataset:         dataset,
		Progress:        progress,
		CreatedTime:     createdTime,
		LastUpdatedTime: lastUpdatedTime,
		Features:        features,
		Filters:         filters,
	}, nil
}

// FetchRequestFeatures pulls request feature information from Postgres.
func (s *Storage) FetchRequestFeatures(requestID string) ([]*model.Feature, error) {
	sql := fmt.Sprintf("SELECT request_id, feature_name, feature_type FROM %s WHERE request_id = $1;", featureTableName)

	rows, err := s.client.Query(sql, requestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request features from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results := make([]*model.Feature, 0)
	for rows.Next() {
		var requestID string
		var featureName string
		var featureType string

		err = rows.Scan(&requestID, &featureName, &featureType)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse request features from Postgres")
		}

		results = append(results, &model.Feature{
			RequestID:     requestID,
			FeatureName: featureName,
			FeatureType: featureType,
		})
	}

	return results, nil
}

// FetchRequestFilters pulls request filter information from Postgres.
func (s *Storage) FetchRequestFilters(requestID string, features []*model.Feature) (*model.FilterParams, error) {
	sql := fmt.Sprintf("SELECT request_id, feature_name, filter_type, filter_mode, filter_min, filter_max, filter_categories FROM %s WHERE request_id = $1;", filterTableName)

	rows, err := s.client.Query(sql, requestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request filters from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	filters := &model.FilterParams{
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

		err = rows.Scan(&requestID, &featureName, &filterType, &filterMode, &filterMin, &filterMax, &filterCategories)
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
		case model.NumericalFilter:
			filters.Filters = append(filters.Filters, model.NewNumericalFilter(
				featureName,
				filterMode,
				filterMin,
				filterMax,
			))
		}
	}

	for _, feature := range features {
		filters.Variables = append(filters.Variables, feature.FeatureName)
	}

	return filters, nil
}
