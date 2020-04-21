import _ from "lodash";
import { Route, Location } from "vue-router";
import { Dictionary } from "./dict";
import {
  JOINED_VARS_INSTANCE_PAGE,
  AVAILABLE_TARGET_VARS_INSTANCE_PAGE,
  AVAILABLE_TRAINING_VARS_INSTANCE_PAGE,
  TRAINING_VARS_INSTANCE_PAGE,
  RESULT_TRAINING_VARS_INSTANCE_PAGE
} from "../store/route/index";
import { SummaryMode } from "../store/dataset";

// TODO: should really have a separate definintion for each route
export interface RouteArgs {
  dataset?: string;
  terms?: string;
  filters?: string;
  training?: string;
  target?: string;
  include?: string;
  solutionId?: string;
  highlights?: string;
  row?: string;
  residualThresholdMin?: string;
  residualThresholdMax?: string;
  joinDatasets?: string;
  joinColumnA?: string;
  joinColumnB?: string;
  joinAccuracy?: string;
  baseColumnSuggestions?: string; // suggested base join columns
  joinColumnSuggestions?: string; // suggested target join columns
  groupingType?: string;
  availableTargetVarsPage?: number;
  task?: string;
  varModes?: string;
  varRanked?: string;
  produceRequestId?: string;
  fittedSolutionId?: string;

  // we currently don't have a way to add these to the interface
  //
  // JOINED_VARS_INSTANCE_PAGE?: string;
  // AVAILABLE_TARGET_VARS_INSTANCE_PAGE?: string;
  // AVAILABLE_TRAINING_VARS_INSTANCE_PAGE?: string;
  // TRAINING_VARS_INSTANCE_PAGE?: string;
  // RESULT_TRAINING_VARS_INSTANCE_PAGE?: string;
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

  // If `undefined` or empty array do not add property. This is to allow args
  // of `''` and `null` to overwrite existing values.

  if (!_.isUndefined(args.dataset)) {
    query.dataset = args.dataset;
  }
  if (!_.isUndefined(args.terms)) {
    query.terms = args.terms;
  }
  if (!_.isUndefined(args.target)) {
    query.target = args.target;
  }
  if (!_.isUndefined(args.include)) {
    query.include = args.include;
  }
  if (!_.isUndefined(args.solutionId)) {
    query.solutionId = args.solutionId;
  }
  if (!_.isUndefined(args.filters)) {
    query.filters = args.filters;
  }
  if (!_.isUndefined(args.training)) {
    query.training = args.training;
  }
  if (!_.isUndefined(args.residualThresholdMin)) {
    query.residualThresholdMin = args.residualThresholdMin;
  }
  if (!_.isUndefined(args.residualThresholdMax)) {
    query.residualThresholdMax = args.residualThresholdMax;
  }
  if (!_.isUndefined(args.highlights)) {
    query.highlights = args.highlights;
  }
  if (!_.isUndefined(args.row)) {
    query.row = args.row;
  }
  if (!_.isUndefined(args.joinDatasets)) {
    query.joinDatasets = args.joinDatasets;
  }
  if (!_.isUndefined(args.joinColumnA)) {
    query.joinColumnA = args.joinColumnA;
  }
  if (!_.isUndefined(args.joinColumnB)) {
    query.joinColumnB = args.joinColumnB;
  }
  if (!_.isUndefined(args.baseColumnSuggestions)) {
    query.baseColumnSuggestions = args.baseColumnSuggestions;
  }
  if (!_.isUndefined(args.joinColumnSuggestions)) {
    query.joinColumnSuggestions = args.joinColumnSuggestions;
  }
  if (!_.isUndefined(args.joinAccuracy)) {
    query.joinAccuracy = args.joinAccuracy;
  }
  if (!_.isUndefined(args.groupingType)) {
    query.groupingType = args.groupingType;
  }
  if (!_.isUndefined(args.task)) {
    query.task = args.task;
  }
  if (!_.isUndefined(args.produceRequestId)) {
    query.produceRequestId = args.produceRequestId;
  }
  if (!_.isUndefined(args.varModes)) {
    query.varModes = args.varModes;
  }
  if (!_.isUndefined(args.varRanked)) {
    query.varRanked = args.varRanked;
  }
  if (!_.isUndefined(args.fittedSolutionId)) {
    query.fittedSolutionId = args.fittedSolutionId;
  }
  if (args[JOINED_VARS_INSTANCE_PAGE]) {
    query[JOINED_VARS_INSTANCE_PAGE] = args[JOINED_VARS_INSTANCE_PAGE];
  }
  if (args[AVAILABLE_TARGET_VARS_INSTANCE_PAGE]) {
    query[AVAILABLE_TARGET_VARS_INSTANCE_PAGE] =
      args[AVAILABLE_TARGET_VARS_INSTANCE_PAGE];
  }
  if (args[AVAILABLE_TRAINING_VARS_INSTANCE_PAGE]) {
    query[AVAILABLE_TRAINING_VARS_INSTANCE_PAGE] =
      args[AVAILABLE_TRAINING_VARS_INSTANCE_PAGE];
  }
  if (args[TRAINING_VARS_INSTANCE_PAGE]) {
    query[TRAINING_VARS_INSTANCE_PAGE] = args[TRAINING_VARS_INSTANCE_PAGE];
  }
  if (args[RESULT_TRAINING_VARS_INSTANCE_PAGE]) {
    query[RESULT_TRAINING_VARS_INSTANCE_PAGE] =
      args[RESULT_TRAINING_VARS_INSTANCE_PAGE];
  }

  return query;
}

export function varModesToString(varModes: Map<string, SummaryMode>): string {
  // serialize the modes map into a string and add to the route
  return Array.from(varModes)
    .reduce(
      (acc, curr) => {
        acc.push(`${curr[0]}:${curr[1]}`);
        return acc;
      },
      [] as String[]
    )
    .join(",");
}
