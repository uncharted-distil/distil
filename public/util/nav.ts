import VueRouter from 'vue-router';
import store from '../store/store';
import { createRouteEntry } from '../util/routes';
import { restoreView } from '../util/view';
import { HOME_ROUTE, SEARCH_ROUTE, SELECT_TARGET_ROUTE, SELECT_TRAINING_ROUTE, RESULTS_ROUTE } from '../store/route/index';
import { getters as routeGetters } from '../store/route/module';

export function gotoView(router: VueRouter, view: string) {
	const dataset = routeGetters.getRouteDataset(store);
	const prev = restoreView(store, view, dataset);
	const entry = createRouteEntry(view, prev ? prev.query : {});
	router.push(entry);
}

export function gotoHome(router: VueRouter) {
	gotoView(router, HOME_ROUTE);
}

export function gotoSearch(router: VueRouter) {
	gotoView(router, SEARCH_ROUTE);
}

export function gotoSelectTarget(router: VueRouter) {
	gotoView(router, SELECT_TARGET_ROUTE);
}

export function gotoSelectData(router: VueRouter) {
	gotoView(router, SELECT_TRAINING_ROUTE);
}

export function gotoResults(router: VueRouter) {
	gotoView(router, RESULTS_ROUTE);
}
