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

import VueRouter from "vue-router";
import { VariableSummary } from "../store/dataset";
import { getters as datasetGetters } from "../store/dataset/module";
import { JOIN_DATASETS_ROUTE } from "../store/route/index";
import { getters as routeGetters } from "../store/route/module";
import store from "../store/store";
import { addRecentDataset, minimumRouteKey } from "./data";
import { createRouteEntry, overlayRouteEntry } from "./routes";

export function loadJoinedDataset(
  router: VueRouter,
  datasetID: string,
  target: string
) {
  const priorPath = routeGetters.getPriorPath(store);
  const returnPath = priorPath ? priorPath : routeGetters.getRoutePath(store);
  const datasets = datasetGetters.getDatasets(store);
  const targetDatasetID = datasets.reduce((a, d) => {
    if (d.id.includes(datasetID) && d.id > a) {
      a = d.id;
    }
    return a;
  }, "");
  const entry = createRouteEntry(returnPath, {
    dataset: targetDatasetID,
    target: target,
    task: routeGetters.getRouteTask(store),
  });
  router.push(entry).catch((err) => console.warn(err));
  addRecentDataset(targetDatasetID);
}

export function loadJoinView(
  router: VueRouter,
  datasetA: string,
  datasetB: string
) {
  const sourceRoute = routeGetters.getRoutePath(store);
  const entry = createRouteEntry(JOIN_DATASETS_ROUTE, {
    joinDatasets: datasetA + "," + datasetB,
    priorRoute: sourceRoute,
    previousTarget: routeGetters.getRouteTargetVariable(store),
  });
  router.push(entry).catch((err) => console.warn(err));
}

export function swapJoinView(router: VueRouter) {
  const joinedDataset = routeGetters.getRouteJoinDatasets(store);
  // swap the datasets themselves
  const tmp = joinedDataset[1];
  joinedDataset[1] = joinedDataset[0];
  joinedDataset[0] = tmp;

  const joinPairs = routeGetters.getJoinPairs(store);
  const entry = overlayRouteEntry(routeGetters.getRoute(store), {
    joinPairs: [
      ...joinPairs.map((jp) => {
        // swap the join pairs
        const first = jp.first;
        jp.first = jp.second;
        jp.second = first;
        return JSON.stringify(jp);
      }),
    ],
    joinDatasets: joinedDataset.join(),
  });
  router.push(entry).catch((err) => console.warn(err));
}

export function getVariableSummaries(
  context,
  dataset?: string
): VariableSummary[] {
  let variables = routeGetters.getJoinDatasetsVariables(context);
  if (dataset) {
    variables = variables.filter((v) => {
      return v.datasetName === dataset;
    });
  }
  const summaries = datasetGetters.getVariableSummariesDictionary(context);
  const routeKey = minimumRouteKey();
  const result = [] as VariableSummary[];
  variables.forEach((v) => {
    if (summaries[v.key + dataset])
      if (summaries[v.key + dataset][routeKey])
        result.push(summaries[v.key + dataset][routeKey]);
  });
  return result;
}
