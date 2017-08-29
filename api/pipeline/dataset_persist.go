package pipeline

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/mitchellh/hashstructure"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/plog"
)

// FilteredDataProvider defines a function that will fetch data from a back end source given
// a set of filter parameters.
type FilteredDataProvider func(dataset string, filters *model.FilterParams) (*model.FilteredData, error)

// Hash the filter set
func getFilteredDatasetHash(dataset string, filterParams *model.FilterParams) (uint64, error) {
	hash, err := hashstructure.Hash([]interface{}{dataset, *filterParams}, nil)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to generate hashcode for %s", dataset)
	}
	return hash, nil
}

// PersistFilteredData creates a hash code from the combination of the dataset name and its filter
// state, and saves the filtered data to disk if it hasn't been previously.  The path to the data
// is returned.
func PersistFilteredData(fetchData FilteredDataProvider, datasetDir string, dataset string, filters *model.FilterParams) (string, error) {

	// parse the dataset and its filter state and generate a hashcode from both
	hash, err := getFilteredDatasetHash(dataset, filters)
	if err != nil {
		return "", err
	}

	// check to see if we already have this filtered dataset saved - return the path
	// if so
	path := path.Join(datasetDir, fmt.Sprintf("%d.csv", hash))
	if fileExists(path) {
		log.Infof("Found cached data for %s with hash %d", dataset, hash)
		return path, nil
	}

	// get the filtered dataset from elastic search
	start := time.Now()
	filteredData, err := fetchData(dataset, filters)
	if err != nil {
		return "", err
	}
	if len(filteredData.Values) <= 0 {
		log.Infof("No data available for %s after filter application", dataset)
		return "", nil
	}

	// write it to a csv file
	if err := os.Mkdir(datasetDir, 0700); err != nil && !os.IsExist(err) {
		return "", errors.Wrapf(err, "unable to create dataset dir %s", datasetDir)
	}

	file, err := os.Create(path)
	if err != nil {
		return "", errors.Wrapf(err, "unable to persist data to %s", datasetDir)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	strVals := make([]string, len(filteredData.Values[0]))
	for _, row := range filteredData.Values {
		// convert vals in row to string
		for i, value := range row {
			strVals[i] = fmt.Sprintf("%v", value)
		}
		err := writer.Write(strVals)
		if err != nil {
			log.Errorf("%v", errors.Wrapf(err, "unable to persist %v", strVals))
		}
	}

	log.Infof("Persisted data for %s with hash %d to %s in %v", dataset, hash, datasetDir, time.Since(start))
	return path, nil
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
