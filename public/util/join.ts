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

import { JOIN_DATASETS_ROUTE } from "../store/route/index";
import { getters as routeGetters } from "../store/route/module";
import { getters as datasetGetters } from "../store/dataset/module";
import { createRouteEntry } from "./routes";
import { addRecentDataset, minimumRouteKey } from "./data";
import store from "../store/store";
import VueRouter from "vue-router";
import { VariableSummary } from "../store/dataset";

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
  const target = routeGetters.getRouteTargetVariable(store);
  const entry = createRouteEntry(JOIN_DATASETS_ROUTE, {
    joinDatasets: datasetA + "," + datasetB,
    priorRoute: sourceRoute,
    target: target,
  });
  router.push(entry).catch((err) => console.warn(err));
}

export function getVariableSummaries(context): VariableSummary[] {
  const variables = routeGetters.getJoinDatasetsVariables(context);
  const summaries = datasetGetters.getVariableSummariesDictionary(context);
  const routeKey = minimumRouteKey();
  const result = [];
  variables.forEach((v) => {
    if (summaries[v.key]) result.push(summaries[v.key][routeKey]);
  });
  return result;
}
