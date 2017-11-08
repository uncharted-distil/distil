import { decodeFilters, FilterMap } from '../../util/filters';
import { Route } from 'vue-router';

export const getters = {
	getRoute(state: Route): Route {
		return state;
	},

	getRoutePath(state: Route): string {
		return state.path;
	},

	getRouteTerms(state: Route): string {
		return state.query.terms;
	},

	getRouteDataset(state: Route): string {
		return state.query.dataset;
	},

	getRouteTrainingVariables(state: Route): string {
		return state.query.training ? state.query.training : null
	},

	getRouteTargetVariable(state: Route): string {
		return state.query.target ? state.query.target : null;
	},

	getRouteCreateRequestId(state: Route): string {
		return state.query.createRequestId;
	},

	getRouteResultId(state: Route): string {
		return state.query.resultId;
	},

	getRouteFilters(state: Route): string {
		return state.query.filters ? state.query.filters : null
	},

	getRouteResultFilters(state: Route): string {
		return state.query.results ? state.query.results : null;
	},

	getRouteResidualThreshold(state: Route): string {
		return state.query.residualThreshold;
	},

	getDecodedFilters(state: Route): FilterMap {
		return decodeFilters(state.query.filters ? state.query.filters : {} as any);
	},

	getDecodedResultsFilters(state: Route): FilterMap {
		return decodeFilters(state.query.results ? state.query.results : {} as any);
	}
}
