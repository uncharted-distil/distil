import { HighlightRoot, RowSelection } from '../data/index';
import { decodeFilters, Filter, FilterParams } from '../../util/filters';
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

	getRouteResultTrainingVarsPage(state: Route): number {
		return state.query.resultTrainingVarsPage ? _.toNumber(state.query.resultTrainingVarsPage) : 1
	},

	getRouteTargetVariable(state: Route): string {
		return state.query.target ? state.query.target : null;
	},

	getRouteSolutionId(state: Route): string {
		return state.query.solutionId ? state.query.solutionId : null;
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

	getDecodedFilters(state: Route, getters: any): Filter[] {
		return decodeFilters(state.query.filters).slice();
	},

	getDecodedFilterParams(state: Route, getters: any): FilterParams {
		const filters = getters.getDecodedFilters;
		const filterParams = {
			filters: filters,
			variables: []
		};
		// add training vars
		const training = getters.getRouteTrainingVariables as string;
		if (training) {
			filterParams.variables = filterParams.variables.concat(training.split(','));
		}
		// add target vars
		const target = getters.getRouteTargetVariable as string;
		if (target) {
			filterParams.variables.push(target);
		}
		return filterParams;
	},

	getDecodedHighlightRoot(state: Route): HighlightRoot {
		return decodeHighlights(state.query.highlights);
	},

	getDecodedRowSelection(state: Route): RowSelection {
		return decodeRowSelection(state.query.row);
	},
}
