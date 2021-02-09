import _ from "lodash";
import { Route, Location } from "vue-router";
import { Dictionary } from "./dict";
import { SummaryMode } from "../store/dataset";
import { ColorScaleNames } from "./data";
// TODO: should really have a separate definition for each route
export interface RouteArgs {
  clustering?: string;
  dataset?: string;
  terms?: string;
  filters?: string;
  training?: string;
  target?: string;
  explore?: string;
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
  // added page & search args directly since we can't use the consts as names
  availableTargetVarsPage?: number;
  availableTrainingVarsPage?: number;
  joinedVarsPage?: number;
  resultTrainingVarsPage?: number;
  trainingVarsPage?: number;
  availableTargetVarsSearch?: string;
  availableTrainingVarsSearch?: string;
  joinedVarsSearch?: string;
  resultTrainingVarsSearch?: string;
  trainingVarsSearch?: string;
  task?: string;
  dataMode?: string;
  varModes?: string;
  varRanked?: string;
  produceRequestId?: string;
  fittedSolutionId?: string;
  singleSolution?: string;
  colorScale?: ColorScaleNames;
  predictionsDataset?: string;
  bandCombinationId?: string;
  imageAttention?: boolean;
  modelTimeLimit?: number;
  modelLimit?: number;
  modelQuality?: string;
  dataSize?: number;
  metrics?: string;
  trainTestSplit?: number;
  timestampSplit?: number;
  annotationHasChanged?: boolean;
  // orderBy contains variable names that will order the dataset
  orderBy?: string;
  outlier?: string;
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
    {} as RouteArgs
  );
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
  const query = validateQueryArgs(args) as Dictionary<string>;
  return { path, query };
}

/* Initialize a new object from the supplied route. */
export function overlayRouteEntry(
  route: Route,
  args: RouteArgs = {}
): Location {
  const path = route.path;
  const query = _.merge({}, route.query, validateQueryArgs(args));
  return { path, query };
}

export function getRouteFacetPage(key: string, route: Route): number {
  const page = route.query[key] as string;
  return page ? parseInt(page) : 1;
}

export function getRouteFacetSearch(key: string, route: Route): string {
  const searchQuery = route.query[key] as string;
  return searchQuery ? searchQuery : "";
}

export function varModesToString(varModes: Map<string, SummaryMode>): string {
  // serialize the modes map into a string and add to the route
  return Array.from(varModes)
    .reduce((acc, curr) => {
      acc.push(`${curr[0]}:${curr[1]}`);
      return acc;
    }, [] as string[])
    .join(",");
}
