package model

// FetchResults returns the set of test predictions made by a given pipeline with requested filtering applied.
func FetchResults(client Storage, dataset string, index string, resultsURI string, filterParams *FilterParams) (*FilteredData, error) {
	return client.FetchResults(dataset, index, resultsURI, filterParams)
}

// FetchResultsSummary returns a histogram summarizing prediction results
func FetchResultsSummary(client Storage, resultsURI string, index string, dataset string) (*Histogram, error) {
	return client.FetchResultsSummary(dataset, resultsURI, index)
}
