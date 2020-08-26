import _ from "lodash";
import { Route, Location, RouteRecord } from "vue-router";
import { Dictionary } from "./dict";
import {
  JOINED_VARS_INSTANCE_PAGE,
  AVAILABLE_TARGET_VARS_INSTANCE_PAGE,
  AVAILABLE_TRAINING_VARS_INSTANCE_PAGE,
  TRAINING_VARS_INSTANCE_PAGE,
  RESULT_TRAINING_VARS_INSTANCE_PAGE,
} from "../store/route/index";
import { SummaryMode } from "../store/dataset";

// TODO: should really have a separate definintion for each route
export interface RouteArgs {
  clustering?: string;
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
  dataMode?: string;
  varModes?: string;
  varRanked?: string;
  produceRequestId?: string;
  fittedSolutionId?: string;
  singleSolution?: string;
  predictionsDataset?: string;
  bandCombinationId?: string;
  modelTimeLimit?: number;
  modelLimit?: number;
  modelQuality?: string;
  dataSize?: number;

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
    query: validateQueryArgs(args) as Dictionary<string>,
  };

  return routeEntry;
}

export function overlayRouteEntry(route: Route, args: RouteArgs): Location {
  // initialize a new object from the supplied route
  const routeEntry: Location = {
    path: route.path,
    query: _.merge({}, route.query, validateQueryArgs(args)),
  };
  return routeEntry;
}

export function getRouteFacetPage(key: string, route: Route): number {
  const page = route.query[key] as string;
  return page ? parseInt(page) : 1;
}

function validateQueryArgs(args: RouteArgs): RouteArgs {
  return _.reduce(
    args,
    (query, value, arg) => {
      if (!_.isUndefined(value)) {
        query[arg] = value;
      }
      return query;
    },
    {} as RouteArgs,
  );
}

export function varModesToString(varModes: Map<string, SummaryMode>): string {
  // serialize the modes map into a string and add to the route
  return Array.from(varModes)
    .reduce((acc, curr) => {
      acc.push(`${curr[0]}:${curr[1]}`);
      return acc;
    }, [] as String[])
    .join(",");
}
