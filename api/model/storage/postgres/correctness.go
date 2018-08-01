package postgres

import (
	"fmt"
	"math"
	"strings"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

// FetchCorrectnessSummary fetches a histogram of the residuals associated with a set of numerical predictions.
func (s *Storage) FetchCorrectnessSummary(dataset string, resultURI string, filterParams *model.FilterParams) (*model.Histogram, error) {
	datasetResult := s.getResultTable(dataset)
	targetName, err := s.getResultTargetName(datasetResult, resultURI)
	if err != nil {
		return nil, err
	}

	variable, err := s.getResultTargetVariable(dataset, targetName)
	if err != nil {
		return nil, err
	}

	// get filter where / params
	wheres, params, err := s.buildResultQueryFilters(dataset, resultURI, filterParams)
	if err != nil {
		return nil, err
	}

	wheres = append(wheres, fmt.Sprintf("result.result_id = $%d AND result.target = $%d ", len(params)+1, len(params)+2))
	params = append(params, resultURI, targetName)

	query := fmt.Sprintf(
		`SELECT data."%s", result.value, COUNT(*) AS count
		 FROM %s AS result INNER JOIN %s AS data ON result.index = data."%s"
		 WHERE %s
		 GROUP BY result.value, data."%s"
		 ORDER BY count desc;`,
		targetName, datasetResult, dataset, model.D3MIndexFieldName, strings.Join(wheres, " AND "), targetName)

	// execute the postgres query
	res, err := s.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result summaries from postgres")
	}
	defer res.Close()

	return s.parseHistogram(res, variable)
}

func (s *Storage) parseHistogram(rows *pgx.Rows, variable *model.Variable) (*model.Histogram, error) {

	termsAggName := model.TermsAggPrefix + variable.Key

	// extract the counts
	countMap := map[string]map[string]int64{}
	if rows != nil {
		for rows.Next() {
			var predictedTerm string
			var targetTerm string
			var bucketCount int64
			err := rows.Scan(&targetTerm, &predictedTerm, &bucketCount)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("no %s histogram aggregation found", termsAggName))
			}
			if len(countMap[predictedTerm]) == 0 {
				countMap[predictedTerm] = map[string]int64{}
			}
			countMap[predictedTerm][targetTerm] = bucketCount
		}
	}

	correctBucket := &model.Bucket{
		Key: "Correct",
	}
	incorrectBucket := &model.Bucket{
		Key: "Incorrect",
	}

	for predictedKey, targetCounts := range countMap {
		for targetKey, count := range targetCounts {
			if predictedKey == targetKey {
				correctBucket.Count += count
			} else {
				incorrectBucket.Count += count
			}
		}
	}

	min := int64(math.MaxInt32)
	max := int64(-math.MaxInt32)
	if incorrectBucket.Count < correctBucket.Count {
		min = incorrectBucket.Count
		max = correctBucket.Count
	} else {
		min = correctBucket.Count
		max = incorrectBucket.Count
	}

	// assign histogram attributes
	return &model.Histogram{
		Label:   variable.Label,
		Key:     variable.Key,
		VarType: variable.Type,
		Type:    model.CategoricalType,
		Buckets: []*model.Bucket{
			correctBucket,
			incorrectBucket,
		},
		Extrema: &model.Extrema{
			Min: float64(min),
			Max: float64(max),
		},
	}, nil
}
