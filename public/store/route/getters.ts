import { decodeFilters, Filter } from '../../util/filters';
import { HighlightRoot } from '../data/index';
import { decodeHighlights } from '../../util/highlights'
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

	getRouteHighlightRoot(state: Route): string {
		return state.query.highlights ? state.query.highlights : null
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
		return decodeFilters(state.query.filters);
	},

	getDecodedHighlightRoot(state: Route): HighlightRoot {
		return decodeHighlights(state.query.highlights);
	}
}
