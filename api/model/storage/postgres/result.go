package postgres

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
)

// PersistResult stores the pipeline result to Postgres.
func (s *Storage) PersistResult(dataset string, pipelineID string, resultURI string) error {
	// Read the results file.
	file, err := os.Open(resultURI)
	if err != nil {
		return errors.Wrap(err, "unable open pipeline result file")
	}
	csvReader := csv.NewReader(bufio.NewReader(file))
	csvReader.TrimLeadingSpace = true
	records, err := csvReader.ReadAll()
	if err != nil {
		return errors.Wrap(err, "unable load pipeline result as csv")
	}
	if len(records) <= 0 || len(records[0]) <= 0 {
		return errors.Wrap(err, "pipeline csv empty")
	}

	// currently only support a single result column.
	if len(records[0]) > 2 {
		log.Warnf("Result contains %s columns, expected 2.  Additional columns will be ignored.", len(records[0]))
	}

	// Header row will have the target.
	targetName := records[0][1]

	// store all results to the storage
	for i := 1; i < len(records); i++ {
		// Each data row is index, target.
		err = nil

		// handle the parsed result/error
		if err != nil {
			return errors.Wrap(err, "failed csv value parsing")
		}
		parsedVal, err := strconv.ParseInt(records[i][0], 10, 64)
		if err != nil {
			return errors.Wrap(err, "failed csv index parsing")
		}

		// store the result to the storage
		s.executeInsertResultStatement(dataset, pipelineID, resultURI, parsedVal, targetName, records[i][1])
	}

	return nil
}

func (s *Storage) executeInsertResultStatement(dataset string, pipelineID string, resultID string, index int64, target string, value string) error {
	statement := fmt.Sprintf("INSERT INTO %s (pipeline_id, result_id, index, target, value) VALUES (?, ?, ?, ?, ?);", dataset)

	_, err := s.client.Exec(statement, pipelineID, resultID, index, target, value)

	return err
}
