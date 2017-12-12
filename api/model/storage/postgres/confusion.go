package postgres

import (
	"fmt"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

// FetchConfusionSummary fetches a histogram of the residuals associated with a set of numerical predictions.
func (s *Storage) FetchConfusionSummary(dataset string, resultURI string, index string) (*model.Histogram, error) {
	datasetResult := s.getResultTable(dataset)
	targetName, err := s.getResultTargetName(datasetResult, resultURI, index)
	if err != nil {
		return nil, err
	}

	variable, err := s.getResultTargetVariable(dataset, index, targetName)
	if err != nil {
		return nil, err
	}

	// Just return a nil in the case where we were asked to return a confusion matrix for a non-numeric variable.
	if model.IsCategorical(variable.Type) {
		// fetch numeric histograms
		residuals, err := s.fetchConfusionHistogram(resultURI, dataset, variable)
		if err != nil {
			return nil, err
		}
		return residuals, nil
	}
	return nil, nil
}

func (s *Storage) fetchConfusionHistogram(resultURI string, dataset string, variable *model.Variable) (*model.Histogram, error) {
	targetName := variable.Name

	query := fmt.Sprintf("SELECT base.%s, result.value, COUNT(*) AS count "+
		"FROM %s_result AS result INNER JOIN %s AS base ON result.index = base.\"d3mIndex\" "+
		"WHERE result.result_id = $1 and result.target = $2 "+
		"GROUP BY result.value, base.%s "+
		"ORDER BY count desc;", targetName, dataset, dataset, targetName)

	// execute the postgres query
	res, err := s.client.Query(query, resultURI, targetName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result summaries from postgres")
	}
	defer res.Close()

	return parseConfusionHistogram(res, variable)
}

func parseConfusionHistogram(rows *pgx.Rows, variable *model.Variable) (*model.Histogram, error) {
	// get terms agg name
	termsAggName := model.TermsAggPrefix + variable.Name

	// extract the counts
	countMap := map[string]map[string]int64{}
	if rows != nil {
		for rows.Next() {
			var predictedTerm string
			var targetTerm string
			var bucketCount int64
			err := rows.Scan(&predictedTerm, &targetTerm, &bucketCount)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("no %s histogram aggregation found", termsAggName))
			}
			if len(countMap[predictedTerm]) == 0 {
				countMap[predictedTerm] = map[string]int64{}
			}
			countMap[predictedTerm][targetTerm] = bucketCount
		}
	}

	// convert the extracted counts into buckets
	buckets := make([]*model.Bucket, 0)
	for predictedKey, targetCounts := range countMap {
		bucket := model.Bucket{
			Key:     predictedKey,
			Count:   0,
			Buckets: []*model.Bucket{},
		}
		for targetKey, count := range targetCounts {
			targetBucket := model.Bucket{
				Key:   targetKey,
				Count: count,
			}
			bucket.Count = bucket.Count + count
			bucket.Buckets = append(bucket.Buckets, &targetBucket)
		}
		buckets = append(buckets, &bucket)
	}

	// assign histogram attributes
	return &model.Histogram{
		Name:    variable.Name,
		Type:    model.CategoricalType,
		Buckets: buckets,
	}, nil
}
