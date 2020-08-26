import VueRouter from "vue-router";
import store from "../store/store";
import { createRouteEntry } from "../util/routes";
import { restoreView } from "../util/view";
import {
  APPLY_MODEL_ROUTE,
  // HOME_ROUTE,
  SEARCH_ROUTE,
  // GROUPING_ROUTE,
  JOIN_DATASETS_ROUTE,
  SELECT_TARGET_ROUTE,
  SELECT_TRAINING_ROUTE,
  RESULTS_ROUTE,
  PREDICTION_ROUTE,
} from "../store/route/index";
import { getters as routeGetters } from "../store/route/module";

export function gotoView(router: VueRouter, view: string) {
  const key =
    routeGetters.getRouteJoinDatasetsHash(store) ||
    routeGetters.getRouteDataset(store);
  const prev = restoreView(view, key);
  console.log(`Restoring view: ${view} for key ${key}`);
  const entry = createRouteEntry(view, prev ? prev.query : {});
  router.push(entry);
}

// export function gotoHome(router: VueRouter) {
//   gotoView(router, HOME_ROUTE);
// }

export function gotoSearch(router: VueRouter) {
  gotoView(router, SEARCH_ROUTE);
}

export function gotoJoinDatasets(router: VueRouter) {
  gotoView(router, JOIN_DATASETS_ROUTE);
}

// export function gotoVariableGrouping(router: VueRouter) {
// 	gotoView(router, GROUPING_ROUTE);
// }

export function gotoSelectTarget(router: VueRouter) {
  gotoView(router, SELECT_TARGET_ROUTE);
}

export function gotoSelectData(router: VueRouter) {
  gotoView(router, SELECT_TRAINING_ROUTE);
}

export function gotoResults(router: VueRouter) {
  gotoView(router, RESULTS_ROUTE);
}

export function gotoApplyModel(router: VueRouter) {
  gotoView(router, APPLY_MODEL_ROUTE);
}

export function gotoPredictions(route: VueRouter) {
  gotoView(route, PREDICTION_ROUTE);
}
