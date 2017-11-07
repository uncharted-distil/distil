import VueRouter from 'vue-router';
import { Store } from 'vuex';
import { createRouteEntry } from '../util/routes';

export function gotoHome(store: Store<any>, router: VueRouter) {
	const entry = createRouteEntry('/home', {
		terms: store.getters.getRouteTerms()
	});
	router.push(entry);
}

export function gotoSearch(store: Store<any>, router: VueRouter) {
	const entry = createRouteEntry('/search', {
		terms: store.getters.getRouteTerms()
	});
	router.push(entry);
}

export function gotoExplore(store: Store<any>, router: VueRouter) {
	const entry = createRouteEntry('/explore', {
		terms: store.getters.getRouteTerms(),
		dataset: store.getters.getRouteDataset(),
		filters: store.getters.getRouteFilters(),
		target: store.getters.getRouteTargetVariable(),
		training: store.getters.getRouteTrainingVariables()
	});
	router.push(entry);
}

export function gotoSelect(store: Store<any>, router: VueRouter) {
	const entry = createRouteEntry('/select', {
		terms: store.getters.getRouteTerms(),
		dataset: store.getters.getRouteDataset(),
		filters: store.getters.getRouteFilters(),
		target: store.getters.getRouteTargetVariable(),
		training: store.getters.getRouteTrainingVariables()
	});
	router.push(entry);
}

export function gotoPipelines(store: Store<any>, router: VueRouter) {
	const entry = createRouteEntry('/pipelines', {
		terms: store.getters.getRouteTerms(),
		dataset: store.getters.getRouteDataset(),
		filters: store.getters.getRouteFilters(),
		target: store.getters.getRouteTargetVariable(),
		training: store.getters.getRouteTrainingVariables()
	});
	router.push(entry);
}

export function gotoResults(store: Store<any>, router: VueRouter) {
	const entry = createRouteEntry('/results', {
		terms: store.getters.getRouteTerms(),
		dataset: store.getters.getRouteDataset(),
		filters: store.getters.getRouteFilters(),
		target: store.getters.getRouteTargetVariable(),
		training: store.getters.getRouteTrainingVariables(),
	});
	router.push(entry);
}
