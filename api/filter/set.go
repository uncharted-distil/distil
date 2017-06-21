package filter

import (
	"gopkg.in/olivere/elastic.v3"
)

// Set represents a cohesive set of filters to be applied to a search.
type Set struct {
	Size    int
	Filters map[string]Filter
}

// NewSet returns a new filter set.
func NewSet(size int) *Set {
	return &Set{
		Size:    size,
		Filters: make(map[string]Filter),
	}
}

// Search executes the elasticsearch search for the set and returns the
// response.
func (s *Set) Search(client *elastic.Client, dataset string) (*elastic.SearchResult, error) {
	// construct an ES query that fetches documents from the dataset with the
	// supplied variable filters applied
	search := client.Search()
	root := elastic.NewBoolQuery()
	var includes []string
	// for each filter
	for name, filter := range s.Filters {
		// get variable field
		field := name + ".value"
		// apply filter
		query, err := filter.Query(field)
		if err != nil {
			return nil, err
		}
		if query != nil {
			// add query to root bool query
			root.Filter(query)
		}
		// append name to includes
		includes = append(includes, name)
	}
	// get fetch context
	fetchContext := elastic.NewFetchSourceContext(true).Include(includes...)
	// execute the ES query
	return search.
		Query(root).
		Index(dataset).
		Size(s.Size).
		FetchSource(true).
		FetchSourceContext(fetchContext).
		Do()
}
