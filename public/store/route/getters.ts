import { HighlightRoot, RowSelection } from '../data/index';
import { decodeFilters, FilterParams } from '../../util/filters';
import { decodeHighlights } from '../../util/highlights'
import { decodeRowSelection } from '../../util/row';
import { Route } from 'vue-router';
import _ from 'lodash';

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

	getRouteTrainingVarsPage(state: Route): number {
		return state.query.trainingVarsPage ? _.toNumber(state.query.trainingVarsPage) : 1
	},

	getRouteAvailableVarsPage(state: Route): number {
		return state.query.availableVarsPage ? _.toNumber(state.query.availableVarsPage) : 1
	},

	getRouteTargetVariable(state: Route): string {
		return state.query.target ? state.query.target : null;
	},

	getRoutePipelineId(state: Route): string {
		return state.query.pipelineId ? state.query.pipelineId : null;
	},

	getRouteResultId(state: Route): string {
		return state.query.resultId ? state.query.resultId : null;
	},

	getRouteFilters(state: Route): string {
		return state.query.filters ? state.query.filters : null
	},

	getRouteHighlightRoot(state: Route): string {
		return state.query.highlights ? state.query.highlights : null
	},

	getRouteRowSelection(state: Route): string {
		return state.query.row ? state.query.row : null
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

	getDecodedFilterParams(state: Route): FilterParams {
		return decodeFilters(state.query.filters);
	},

	getDecodedHighlightRoot(state: Route): HighlightRoot {
		return decodeHighlights(state.query.highlights);
	},

	getDecodedRowSelection(state: Route): RowSelection {
		return decodeRowSelection(state.query.row);
	},
}
