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

import Vue from "vue";
import VueRouter from "vue-router";
import Search from "../views/Search.vue";
import JoinDatasets from "../views/JoinDatasets.vue";
import ExportSuccess from "../views/ExportSuccess.vue";
import VariableGrouping from "../views/VariableGrouping.vue";
import DataExplorer from "../views/DataExplorer.vue";
import store from "../store/store";
import { getters as routeGetters } from "../store/route/module";
import { saveView } from "../util/view";
import {
  ROOT_ROUTE,
  SEARCH_ROUTE,
  GROUPING_ROUTE,
  JOIN_DATASETS_ROUTE,
  EXPORT_SUCCESS_ROUTE,
  DATA_EXPLORER_ROUTE,
} from "../store/route";

Vue.use(VueRouter);

const router = new VueRouter({
  routes: [
    { path: ROOT_ROUTE, redirect: SEARCH_ROUTE },
    { path: SEARCH_ROUTE, component: Search },
    { path: JOIN_DATASETS_ROUTE, component: JoinDatasets },
    { path: GROUPING_ROUTE, component: VariableGrouping },
    { path: EXPORT_SUCCESS_ROUTE, component: ExportSuccess },
    { path: DATA_EXPLORER_ROUTE, component: DataExplorer },
  ],
});

router.afterEach((_, fromRoute) => {
  let key = routeGetters.getRouteDataset(store);
  if (key === "" || !key) {
    key = routeGetters.getRouteJoinDatasetsHash(store);
  }
  if (key) {
    console.log(`Saving view: ${fromRoute.path} for key ${key}`);
  } else {
    console.log(`Saving view: ${fromRoute.path}`);
  }
  saveView({
    view: fromRoute.path,
    key: key,
    route: fromRoute,
  });
});

export default router;
