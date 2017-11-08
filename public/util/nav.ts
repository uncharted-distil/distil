import VueRouter from 'vue-router';
import { Store } from 'vuex';
import { createRouteEntry } from '../util/routes';
import { getters } from '../store/route/module';

export function gotoHome(store: Store<any>, router: VueRouter) {
	const entry = createRouteEntry('/home', {
		terms: getters.getRouteTerms(store)
	});
	router.push(entry);
}

export function gotoSearch(store: Store<any>, router: VueRouter) {
	const entry = createRouteEntry('/search', {
		terms: getters.getRouteTerms(store)
	});
	router.push(entry);
}

export function gotoExplore(store: Store<any>, router: VueRouter) {
	const entry = createRouteEntry('/explore', {
		terms: getters.getRouteTerms(store),
		dataset: getters.getRouteDataset(store),
		filters: getters.getRouteFilters(store),
		target: getters.getRouteTargetVariable(store),
		training: getters.getRouteTrainingVariables(store)
	});
	router.push(entry);
}

export function gotoSelect(store: Store<any>, router: VueRouter) {
	const entry = createRouteEntry('/select', {
		terms: getters.getRouteTerms(store),
		dataset: getters.getRouteDataset(store),
		filters: getters.getRouteFilters(store),
		target: getters.getRouteTargetVariable(store),
		training: getters.getRouteTrainingVariables(store)
	});
	router.push(entry);
}

export function gotoPipelines(store: Store<any>, router: VueRouter) {
	const entry = createRouteEntry('/pipelines', {
		terms: getters.getRouteTerms(store),
		dataset: getters.getRouteDataset(store),
		filters: getters.getRouteFilters(store),
		target: getters.getRouteTargetVariable(store),
		training: getters.getRouteTrainingVariables(store)
	});
	router.push(entry);
}

export function gotoResults(store: Store<any>, router: VueRouter) {
	const entry = createRouteEntry('/results', {
		terms: getters.getRouteTerms(store),
		dataset: getters.getRouteDataset(store),
		filters: getters.getRouteFilters(store),
		target: getters.getRouteTargetVariable(store),
		training: getters.getRouteTrainingVariables(store),
	});
	router.push(entry);
}
