import VueRouter from 'vue-router';
import { Store } from 'vuex';
import { createRouteEntry } from '../util/routes';
import { restoreView } from '../util/view';
import { HOME_ROUTE, SEARCH_ROUTE, SELECT_ROUTE, CREATE_ROUTE, RESULTS_ROUTE } from '../store/route/index';
import { getters as routeGetters } from '../store/route/module';

export function gotoView(store: Store<any>, router: VueRouter, view: string) {
	const dataset = routeGetters.getRouteDataset(store);
	const prev = restoreView(store, view, dataset);
	const entry = createRouteEntry(view, prev ? prev.query : {});
	router.push(entry);
}

export function gotoHome(store: Store<any>, router: VueRouter) {
	gotoView(store, router, HOME_ROUTE);
}

export function gotoSearch(store: Store<any>, router: VueRouter) {
	gotoView(store, router, SEARCH_ROUTE);
}

export function gotoSelectTarget(store: Store<any>, router: VueRouter) {
	gotoView(store, router, SELECT_ROUTE);
}

export function gotoSelectData(store: Store<any>, router: VueRouter) {
	gotoView(store, router, CREATE_ROUTE);
}

export function gotoResults(store: Store<any>, router: VueRouter) {
	gotoView(store, router, RESULTS_ROUTE);
}
