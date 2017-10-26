import _ from 'lodash';

/**
 * Builds a route entry object that can be directly pushed onto the stack
 * via  call to route.push(). This holds all the app view state to support
 * nav bar navigation.
 *
 * @param {string} path - route path
 * @param {Object} args - the arguments for the route.
 */
export function createRouteEntry(path, args = {}) {
	const query = {};
	if (args.dataset) { query.dataset = args.dataset; }
	if (args.terms) { query.terms = args.terms; }
	if (!_.isEmpty(args.training)) { query.training = args.training; }
	if (args.target) { query.target = args.target; }
	if (args.createRequestId) { query.createRequestId = args.createRequestId; }
	if (!_.isEmpty(args.filters)) { query.filters = args.filters; }
	if (!_.isEmpty(args.results)) { query.results = args.results; }
	if (!_.isEmpty(args.resultId)) { query.resultId = args.resultId; }
	return {
		path: path,
		query: query
	};
}

export function createRouteEntryFromRoute(route, args = {}) {
	// initialize a new object from the supplied route
	const routeEntry = {
		path: route.path,
		query: _.cloneDeep(route.query)
	};

	// merge in the supplied arguments
	_.merge(routeEntry.query, args);

	return routeEntry;
}
