package model

// FetchResults returns the set of test predictions made by a given pipeline.
func FetchResults(client Storage, resultsURI string, index string, dataset string) (*FilteredData, error) {
	return client.FetchResults(dataset, resultsURI, index)
}

// FetchResultsSummary returns a histogram summarizing prediction results
func FetchResultsSummary(client Storage, resultsURI string, index string, dataset string) (*Histogram, error) {
	return client.FetchResultsSummary(dataset, resultsURI, index)
}
