package postgres

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

// PersistModel persists a model to Postgres.
func (s *Storage) PersistModel(modelID string, dataset string, progress string, createdTime time.Time) error {
	sql := fmt.Sprintf("INSERT INTO %s (model_id, dataset, progress, created_time, last_updated_time) VALUES ($1, $2, $3, $4, $4);", modelTableName)

	_, err := s.client.Exec(sql, modelID, dataset, progress, createdTime)

	return err
}

// UpdateModel updates a model in Postgres.
func (s *Storage) UpdateModel(modelID string, progress string, updatedTime time.Time) error {
	sql := fmt.Sprintf("UPDATE %s SET progress = $1, last_updated_time = $2 WHERE model_id = $3;", modelTableName)

	_, err := s.client.Exec(sql, progress, updatedTime, modelID)

	return err
}

// PersistModelFeature persists model feature information to Postgres.
func (s *Storage) PersistModelFeature(modelID string, featureName string, featureType string) error {
	sql := fmt.Sprintf("INSERT INTO %s (model_id, feature_name, feature_type) VALUES ($1, $2, $3);", featureTableName)

	_, err := s.client.Exec(sql, modelID, featureName, featureType)

	return err
}

// PersistModelFilters persists model filters information to Postgres.
func (s *Storage) PersistModelFilters(modelID string, filters *model.FilterParams) error {
	sql := fmt.Sprintf("INSERT INTO %s (model_id, feature_name, filter_type, filter_mode, filter_min, filter_max, filter_categories) VALUES ($1, $2, $3, $4, $5, $6, $7);", filterTableName)

	for _, filter := range filters.Filters {
		switch filter.Type {
		case model.NumericalFilter:
			_, err := s.client.Exec(sql, modelID, filter.Name, model.NumericalFilter, filter.Mode, filter.Min, filter.Max, "")
			if err != nil {
				return err
			}
		case model.CategoricalFilter:
			_, err := s.client.Exec(sql, modelID, filter.Name, model.CategoricalFilter, filter.Mode, 0, 0, strings.Join(filter.Categories, ","))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// FetchModel pulls model information from Postgres.
func (s *Storage) FetchModel(modelID string) (*model.Model, error) {
	sql := fmt.Sprintf("SELECT model_id, dataset, progress, created_time, last_updated_time FROM %s WHERE model_id = $1;", modelTableName)

	rows, err := s.client.Query(sql, modelID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull model from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}
	rows.Next()

	return s.loadModel(rows)
}

func (s *Storage) loadModel(rows *pgx.Rows) (*model.Model, error) {
	var modelID string
	var dataset string
	var progress string
	var createdTime time.Time
	var lastUpdatedTime time.Time

	err := rows.Scan(&modelID, &dataset, &progress, &createdTime, &lastUpdatedTime)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse model from Postgres")
	}

	features, err := s.FetchModelFeatures(modelID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get model features from Postgres")
	}

	filters, err := s.FetchModelFilters(modelID, features)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get model filters from Postgres")
	}

	return &model.Model{
		ModelID:         modelID,
		Dataset:         dataset,
		Progress:        progress,
		CreatedTime:     createdTime,
		LastUpdatedTime: lastUpdatedTime,
		Features:        features,
		Filters:         filters,
	}, nil
}

// FetchModelFeatures pulls model feature information from Postgres.
func (s *Storage) FetchModelFeatures(modelID string) ([]*model.Feature, error) {
	sql := fmt.Sprintf("SELECT model_id, feature_name, feature_type FROM %s WHERE model_id = $1;", featureTableName)

	rows, err := s.client.Query(sql, modelID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull model features from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results := make([]*model.Feature, 0)
	for rows.Next() {
		var modelID string
		var featureName string
		var featureType string

		err = rows.Scan(&modelID, &featureName, &featureType)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse model features from Postgres")
		}

		results = append(results, &model.Feature{
			ModelID:     modelID,
			FeatureName: featureName,
			FeatureType: featureType,
		})
	}

	return results, nil
}

// FetchModelFilters pulls model filter information from Postgres.
func (s *Storage) FetchModelFilters(modelID string, features []*model.Feature) (*model.FilterParams, error) {
	sql := fmt.Sprintf("SELECT model_id, feature_name, filter_type, filter_mode, filter_min, filter_max, filter_categories FROM %s WHERE model_id = $1;", filterTableName)

	rows, err := s.client.Query(sql, modelID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull model filters from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	filters := &model.FilterParams{
		Size: model.DefaultFilterSize,
	}

	for rows.Next() {
		var modelID string
		var featureName string
		var filterType string
		var filterMode string
		var filterMin float64
		var filterMax float64
		var filterCategories string

		err = rows.Scan(&modelID, &featureName, &filterType, &filterMode, &filterMin, &filterMax, &filterCategories)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse model filters from Postgres")
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
