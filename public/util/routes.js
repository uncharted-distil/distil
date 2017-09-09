import _ from 'lodash';

/**
 * Builds a route entry object that can be directly pushed onto the stack
 * via  call to route.push(). This holds all the app view state to support
 * nav bar navigation.
 *
 * @param {string} path - route path
 * @param {Object} args - the arguments for the route.
 * @param {string} args.terms - search terms from the route query string
 * @param {string} args.dataset - dataset name from the route query string
 * @param {Object} args.filters - filters - The list filters from the route query string.
 * @param {string} args.pipelineID - pipelineID
 */
export function createRouteEntry(path, args = {}) {
	const query = {};
	if (args.dataset) { query.dataset = args.dataset; }
	if (args.terms) { query.terms = args.terms; }
	if (!_.isEmpty(args.training)) { query.training = args.training.join(','); }
	if (args.target) { query.target = args.target; }
	if (args.createRequestId) { query.createRequestId = args.createRequestId; }
	if (!_.isEmpty(args.filters)) { query.filters = args.filters; }
	if (!_.isEmpty(args.results)) { query.results = args.results; }
	return {
		path: path,
		query: query
	};
}
