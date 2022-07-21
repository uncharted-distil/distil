/**
 *
 *    Copyright © 2021 Uncharted Software Inc.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

import _ from "lodash";
import { Location, Route } from "vue-router";
import { SummaryMode } from "../store/dataset";
import { ColorScaleNames } from "./color";
import { Dictionary } from "./dict";
// TODO: should really have a separate definition for each route
export interface RouteArgs {
  clustering?: string;
  dataset?: string;
  terms?: string;
  filters?: string;
  training?: string;
  target?: string;
  explore?: string;
  solutionId?: string;
  highlights?: string;
  hasGeoData?: boolean;
  row?: string;
  residualThresholdMin?: string;
  residualThresholdMax?: string;
  joinDatasets?: string;
  joinColumnA?: string;
  joinColumnB?: string;
  joinAccuracy?: string;
  joinPairs?: string[];
  joinInfo?: string;
  dataExplorerState?: string;
  toggledActions?: string;
  baseColumnSuggestions?: string[]; // suggested base join columns
  joinColumnSuggestions?: string[]; // suggested target join columns
  groupingType?: string;
  // added page & search args directly since we can't use the consts as names
  availableTargetVarsPage?: number;
  availableTrainingVarsPage?: number;
  joinedVarsPage?: number;
  resultTrainingVarsPage?: number;
  trainingVarsPage?: number;
  availableTargetVarsSearch?: string;
  availableTrainingVarsSearch?: string;
  topVarsSearch?: string;
  bottomVarsSearch?: string;
  resultTrainingVarsSearch?: string;
  trainingVarsSearch?: string;
  task?: string;
  selectedTask?: string;
  dataMode?: string;
  varModes?: string;
  varRanked?: string;
  produceRequestId?: string;
  fittedSolutionId?: string;
  openSolutions?: string;
  previousTarget?: string;
  singleSolution?: string;
  colorScale?: ColorScaleNames;
  colorScaleVariable?: string;
  imageLayerScale?: ColorScaleNames;
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
  label?: string;
  // orderBy contains variable names that will order the dataset
  orderBy?: string;
  outlier?: string;
  priorRoute?: string;
  positiveLabel?: string;
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
export function overlayRouteReplace(
  route: Route,
  args: RouteArgs = {}
): Location {
  const path = route.path;
  const keys = Object.keys(args);
  const query = _.cloneDeep(route.query);
  keys.forEach((key) => {
    query[key] = args[key];
  });
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
