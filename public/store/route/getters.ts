import { decodeFilters, FilterMap } from '../../util/filters';
import { Route } from 'vue-router';
import { GetterTree } from 'vuex';
import { DistilState } from '../index';

export const getters: GetterTree<Route, DistilState> = {
	getRoute(state: Route) {
		return () => state;
	},

	getRoutePath(state: Route) {
		return () => state.path;
	},

	getRouteTerms(state: Route) {
		return () => state.query.terms;
	},

	getRouteDataset(state: Route) {
		return () => state.query.dataset;
	},

	getRouteTrainingVariables(state: Route) {
		return () => state.query.training ? state.query.training : null;
	},

	getRouteTargetVariable(state: Route) {
		return () => state.query.target ? state.query.target : null;
	},

	getRouteCreateRequestId(state: Route) {
		return () => state.query.createRequestId;
	},

	getRouteResultId(state: Route) {
		return () => state.query.resultId;
	},

	getRouteFilters(state: Route) {
		return () => state.query.filters ? state.query.filters : [];
	},

	getRouteResultFilters(state: Route) {
		return () => state.query.results ? state.query.results : [];
	},

	getRouteFacetsPage(state: Route) {
		return (pageKey: string) => state.query[pageKey];
	},

	getRouteResidualThreshold(state: Route) {
		return () => state.query.residualThreshold;
	},

	getFilters(state: Route) {
		return () => decodeFilters(state.query.filters ? state.query.filters : "") as FilterMap;
	},

	getResultsFilters(state: Route) {
		return decodeFilters(state.query.results ? state.query.results : "") as FilterMap;
	},

	getTrainingVariables(state: Route) {
		return () => state.query.training ? state.query.training.split(',') : [];
	},

	getTargetVariable(state: Route) {
		return () => {
			return state.query.target ? state.query.target : null;
		};
	}
}
