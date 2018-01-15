import VueRouter from 'vue-router';
import { Store } from 'vuex';
import { createRouteEntry } from '../util/routes';
import { restoreView } from '../util/view';
import { getters as routeGetters } from '../store/route/module';

export function gotoView(store: Store<any>, router: VueRouter, view: string, overrides: any) {
	const dataset = routeGetters.getRouteDataset(store);
	const prev = restoreView(store, view, dataset);
	const entry = createRouteEntry(view, prev ? prev.query : overrides);
	router.push(entry);
}

export function gotoHome(store: Store<any>, router: VueRouter) {
	gotoView(store, router, '/home', {
		terms: routeGetters.getRouteTerms(store)
	});
}

export function gotoSearch(store: Store<any>, router: VueRouter) {
	gotoView(store, router, '/search', {
		terms: routeGetters.getRouteTerms(store)
	});
}

export function gotoSelect(store: Store<any>, router: VueRouter) {
	gotoView(store, router, '/select', {
		terms: routeGetters.getRouteTerms(store),
		dataset: routeGetters.getRouteDataset(store),
		filters: routeGetters.getRouteFilters(store),
		target: routeGetters.getRouteTargetVariable(store),
		training: routeGetters.getRouteTrainingVariables(store)
	});
}

export function gotoResults(store: Store<any>, router: VueRouter) {
	gotoView(store, router, '/results', {
		terms: routeGetters.getRouteTerms(store),
		dataset: routeGetters.getRouteDataset(store),
		target: routeGetters.getRouteTargetVariable(store)
	});
}
