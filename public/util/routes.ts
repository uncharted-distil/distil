import _ from 'lodash';
import { Route, Location } from 'vue-router';
import { Dictionary } from './dict';
import { JOINED_VARS_INSTANCE_PAGE, AVAILABLE_TRAINING_VARS_INSTANCE_PAGE,
	TRAINING_VARS_INSTANCE_PAGE, RESULT_TRAINING_VARS_INSTANCE_PAGE } from '../store/route/index';

export interface RouteArgs {
	dataset?: string;
	terms?: string;
	filters?: string;
	training?: string;
	target?: string;
	solutionId?: string;
	highlights?: string;
	row?: string;
	residualThresholdMin?: string;
	residualThresholdMax?: string;
	geo?: string;
	joinDatasets?: string;
	joinColumnA?: string;
	joinColumnB?: string;
	joinAccuracy?: string;

	// we currently don't have a way to add these to the interface
	//
	// JOINED_VARS_INSTANCE_PAGE?: string;
	// AVAILABLE_TRAINING_VARS_INSTANCE_PAGE?: string;
	// AVAILABLE_TARGET_VARS_INSTANCE_PAGE?: string;
	// TRAINING_VARS_INSTANCE_PAGE?: string;
	// RESULT_TRAINING_VARS_INSTANCE_PAGE?: string;
}

export interface Something {

}


/**
 * Builds a route entry object that can be directly pushed onto the stack
 * via  call to route.push(). This holds all the app view state to support
 * nav bar navigation.
 *
 * @param {string} path - route path
 * @param {RouteArgs} args - the arguments for the route.
 */
export function createRouteEntry(path: string, args: RouteArgs = {}): Location {
	const routeEntry: Location = {
		path: path,
		query: validateQueryArgs(args) as Dictionary<string>
	};

	return routeEntry;
}

export function overlayRouteEntry(route: Route, args: RouteArgs): Location {
	// initialize a new object from the supplied route
	const routeEntry: Location = {
		path: route.path,
		query: _.merge({}, route.query, validateQueryArgs(args))
	};
	return routeEntry;
}

export function getRouteFacetPage(key: string, route: Route): number {
	const page = route.query[key] as string;
	return page ? parseInt(page) : 1;
}

function validateQueryArgs(args: RouteArgs): RouteArgs {
	const query: RouteArgs = {};

	if (args.dataset) { query.dataset = args.dataset; }
	if (args.terms) { query.terms = args.terms; }
	if (args.target) { query.target = args.target; }
	if (args.solutionId) { query.solutionId = args.solutionId; }
	if (!_.isEmpty(args.filters)) { query.filters = args.filters; }
	if (!_.isEmpty(args.training)) { query.training = args.training; }
	if (args.residualThresholdMin) { query.residualThresholdMin = args.residualThresholdMin; }
	if (args.residualThresholdMax) { query.residualThresholdMax = args.residualThresholdMax; }
	if (!_.isEmpty(args.highlights)) { query.highlights = args.highlights; }
	if (args.geo) { query.geo = args.geo; }
	if (args.joinDatasets) { query.joinDatasets = args.joinDatasets; }
	if (args.joinColumnA) { query.joinColumnA = args.joinColumnA; }
	if (args.joinColumnB) { query.joinColumnB = args.joinColumnB; }
	if (args.joinAccuracy) { query.joinAccuracy = args.joinAccuracy; }

	if (args[JOINED_VARS_INSTANCE_PAGE]) { query[JOINED_VARS_INSTANCE_PAGE] = args[JOINED_VARS_INSTANCE_PAGE]; }
	if (args[AVAILABLE_TRAINING_VARS_INSTANCE_PAGE]) { query[AVAILABLE_TRAINING_VARS_INSTANCE_PAGE] = args[AVAILABLE_TRAINING_VARS_INSTANCE_PAGE]; }
	if (args[TRAINING_VARS_INSTANCE_PAGE]) { query[TRAINING_VARS_INSTANCE_PAGE] = args[TRAINING_VARS_INSTANCE_PAGE]; }
	if (args[RESULT_TRAINING_VARS_INSTANCE_PAGE]) { query[RESULT_TRAINING_VARS_INSTANCE_PAGE] = args[RESULT_TRAINING_VARS_INSTANCE_PAGE]; }

	return query;
}
