import { Variable, VariableSummary } from '../dataset/index';
import { HighlightRoot, RowSelection } from '../highlights/index';
import { decodeFilters, Filter, FilterParams } from '../../util/filters';
import { decodeHighlights } from '../../util/highlights'
import { decodeRowSelection } from '../../util/row';
import { Dictionary } from '../../util/dict';
import { Route } from 'vue-router';
import _ from 'lodash';

function buildLookup(strs: any[]): Dictionary<boolean> {
	const lookup = {};
	strs.forEach(str => {
		lookup[str] = true;
		lookup[str.toLowerCase()] = true;
	});
	return lookup;
}

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

	getDecodedTrainingVariableNames(state: Route, getters: any): string[] {
		const training = getters.getRouteTrainingVariables;
		return training ? training.split(',') : [];
	},

	getDecodedFilters(state: Route, getters: any): Filter[] {
		return decodeFilters(state.query.filters);
	},

	getDecodedFilterParams(state: Route, getters: any): FilterParams {
		const filters = getters.getDecodedFilters;
		const filterParams = _.cloneDeep({
			filters: filters,
			variables: []
		});
		// add training vars
		const training = getters.getDecodedTrainingVariableNames;
		filterParams.variables = filterParams.variables.concat(training);
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

	getTrainingVariables(state: Route, getters: any): Variable[] {
		const training = getters.getDecodedTrainingVariableNames;
		const lookup = buildLookup(training);
		const variables = getters.getVariables;
		return variables.filter(variable => lookup[variable.name.toLowerCase()]);
	},

	getTrainingVariableSummaries(state: Route, getters: any): VariableSummary[] {
		const training = getters.getDecodedTrainingVariableNames;
		const lookup = buildLookup(training);
		const summaries = getters.getVariableSummaries;
		return summaries.filter(summary => lookup[summary.name.toLowerCase()]);
	},

	getTargetVariable(state: Route, getters: any): Variable {
		const target = getters.getRouteTargetVariable;
		if (target) {
			const variables = getters.getVariables;
			const found = variables.filter(summary => target.toLowerCase() === summary.name.toLowerCase());
			if (found) {
				return found[0];
			}
		}
		return null;
	},

	getTargetVariableSummaries(state: Route, getters: any): VariableSummary[] {
		const target = getters.getRouteTargetVariable;
		if (target) {
			const summaries = getters.getVariableSummaries;
			return summaries.filter(summary => target.toLowerCase() === summary.name.toLowerCase());
		}
		return [];
	},

	getAvailableVariables(state: Route, getters: any): Variable[] {
		const training = getters.getDecodedTrainingVariableNames;
		const target = getters.getRouteTargetVariable;
		const lookup = buildLookup(training.concat([ target ]));
		const variables = getters.getVariables;
		return variables.filter(variable => !lookup[variable.name.toLowerCase()]);
	},

	getAvailableVariableSummaries(state: Route, getters: any): VariableSummary[] {
		const training = getters.getDecodedTrainingVariableNames;
		const target = getters.getRouteTargetVariable;
		const lookup = buildLookup(training.concat([ target ]));
		const summaries = getters.getVariableSummaries;
		return summaries.filter(summary => !lookup[summary.name.toLowerCase()]);
	},

	getSolutionIndex(state: Route, getters: any): number {
		const solutionId = getters.getRouteSolutionId;
		const solutions = getters.getSolutions;
		return _.findIndex(solutions, (solution: any) => {
			return solution.solutionId === solutionId;
		});
	}
}
