import _ from 'lodash';

/**
 * Builds a route entry object that can be directly pushed onto the stack
 * via  call to route.push().  This holds all the app view state to support
 * nav bar navigation.
 *
 * @param {string} path - route path
 * @param {string} dataset - dataset name from the route query string
 * @param {string} terms - search terms from the route query string
 * @param {Object} filters - filters - The list filters from the route query string.
 */
export function createRouteEntry(path, dataset, terms, filters) {
	const query = {};
	if (dataset) { query.dataset = dataset; }
	if (terms) { query.terms = terms; }
	if (!_.isEmpty(filters)) { _.assign(query, filters);}
	return {
		path: path,
		query: query
	};
}
