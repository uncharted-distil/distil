import VueRouter from 'vue-router';
import _ from 'lodash';
import { Store } from 'vuex';
import { createRouteEntry } from '../util/routes';
import { restoreView } from '../util/view';
import { HOME_ROUTE, SEARCH_ROUTE, SELECT_ROUTE, CREATE_ROUTE, RESULTS_ROUTE } from '../store/route/index';
import { getters as routeGetters } from '../store/route/module';

export function gotoView(store: Store<any>, router: VueRouter, view: string, overrides: any) {
	const dataset = routeGetters.getRouteDataset(store);
	const prev = restoreView(store, view, dataset);
	const entry = createRouteEntry(view, prev ? _.merge({}, prev.query, pruneEmpty(overrides)) : overrides);
	router.push(entry);
}

export function gotoHome(store: Store<any>, router: VueRouter) {
	gotoView(store, router, HOME_ROUTE, {
		terms: routeGetters.getRouteTerms(store)
	});
}

export function gotoSearch(store: Store<any>, router: VueRouter) {
	gotoView(store, router, SEARCH_ROUTE, {
		terms: routeGetters.getRouteTerms(store)
	});
}

export function gotoSelectTarget(store: Store<any>, router: VueRouter) {
	gotoView(store, router, SELECT_ROUTE, {
		dataset: routeGetters.getRouteDataset(store),
		filters: routeGetters.getRouteFilters(store),
		target: routeGetters.getRouteTargetVariable(store),
		training: routeGetters.getRouteTrainingVariables(store)
	});
}

export function gotoSelectData(store: Store<any>, router: VueRouter) {
	gotoView(store, router, CREATE_ROUTE, {
		dataset: routeGetters.getRouteDataset(store),
		filters: routeGetters.getRouteFilters(store),
		target: routeGetters.getRouteTargetVariable(store),
		training: routeGetters.getRouteTrainingVariables(store)
	});
}

export function gotoResults(store: Store<any>, router: VueRouter) {
	gotoView(store, router, RESULTS_ROUTE, {
		dataset: routeGetters.getRouteDataset(store),
		target: routeGetters.getRouteTargetVariable(store)
	});
}

function prune(current) {
	_.forIn(current, (value, key) => {
		if (value === undefined ||
			value == null ||
			(_.isString(value) && _.isEmpty(value)) ||
			(_.isObject(value) && _.isEmpty(prune(value)))) {
			delete current[key];
		}
	});
	// remove any leftover undefined values from the delete
	// operation on an array
	if (_.isArray(current)) {
		_.pull(current, undefined);
	}
	return current;
}

function pruneEmpty(obj) {
	return prune(_.cloneDeep(obj));
}
