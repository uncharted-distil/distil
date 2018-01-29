import { decodeFilters, Filter } from '../../util/filters';
import { Highlights, decodeHighlights } from '../../util/highlights'
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

	getRoutePipelineId(state: Route): string {
		return state.query.pipelineId ? state.query.pipelineId : null;
	},

	getRouteFilters(state: Route): string {
		return state.query.filters ? state.query.filters : null
	},

	getRouteResultFilters(state: Route): string {
		return state.query.results ? state.query.results : null;
	},

	getRouteResidualThresholdMin(state: Route): string {
		return state.query.residualThresholdMin;
	},

	getRouteResidualThresholdMax(state: Route): string {
		return state.query.residualThresholdMax;
	},

	getDecodedFilters(state: Route): Filter[] {
		return decodeFilters(state.query.filters ? state.query.filters : {} as any);
	},

	getDecodedResultsFilters(state: Route): Filter[] {
		return decodeFilters(state.query.results ? state.query.results : {} as any);
	},

	getDecodedHighlightedFeatureValues(state: Route): Highlights {
		return decodeHighlights(state.query.highlights);
	}
}
