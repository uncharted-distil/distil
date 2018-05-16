import axios from 'axios';
import { ActionContext } from 'vuex';
import { HighlightState } from './index';
import { DistilState } from '../store';
import { FilterParams, INCLUDE_FILTER } from '../../util/filters';
import { getSolutionsByRequestIds, getSolutionById } from '../../util/solutions';
import { getSummaries, updateCorrectnessHighlightSummary, getCorrectnessCol } from '../../util/data';
import { Variable, Extrema, ES_INDEX } from '../dataset/index';
import { mutations } from './module'
import { SolutionInfo } from '../solutions/index';
import { HighlightRoot } from './index';
import { createFilterFromHighlightRoot, parseHighlightSamples } from '../../util/highlights';
import { getPredictedCol, getVarFromTarget } from '../../util/data';

export type HighlightsContext = ActionContext<HighlightState, DistilState>;

export const actions = {

	fetchDataHighlightSamples(context: HighlightsContext, args: { highlightRoot: HighlightRoot, filters: FilterParams, dataset: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.filters) {
			console.warn('`filters` argument is missing');
			return null;
		}

		const highlightFilter = createFilterFromHighlightRoot(args.highlightRoot, INCLUDE_FILTER);
		if (highlightFilter) {
			args.filters.filters.push(highlightFilter);
		}

		// fetch the data using the supplied filtered
		return axios.post(`distil/data/${ES_INDEX}/${args.dataset}/false`, args.filters)
			.then(res => {
				mutations.updateHighlightSamples(context, parseHighlightSamples(res.data));
			})
			.catch(error => {
				console.error(error);
				mutations.updateHighlightSamples(context, null);
			});
	},

	fetchDataHighlightSummaries(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, filters: FilterParams, variables: Variable[] }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.variables) {
			console.warn('`variables` argument is missing');
			return null;
		}

		const highlightFilter = createFilterFromHighlightRoot(args.highlightRoot, INCLUDE_FILTER);
		if (highlightFilter) {
			args.filters.filters.push(highlightFilter);
		}

		// commit empty place holders, if there is no data
		return Promise.all(args.variables.map(variable => {
			return axios.post(`/distil/variable-summary/${ES_INDEX}/${args.dataset}/${variable.name}`, args.filters)
				.then(response => {
					mutations.updateHighlightSummaries(context, response.data.histogram);
				})
				.catch(error => {
					console.error(error);
				});
		}));
	},

	fetchDataHighlightValues(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, filters: FilterParams, variables: Variable[] }) {
		return Promise.all([
			context.dispatch('fetchDataHighlightSamples', {
				highlightRoot: args.highlightRoot,
				dataset: args.dataset,
				filters: args.filters,
			}),
			context.dispatch('fetchDataHighlightSummaries', {
				highlightRoot: args.highlightRoot,
				dataset: args.dataset,
				variables: args.variables,
				filters: args.filters
			})
		]);
	},

	fetchResultHighlightSummaries(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, solutionId: string, variables: Variable[], extrema: Extrema }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.variables) {
			console.warn('`variables` argument is missing');
			return null;
		}

		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		if (!solution.resultId) {
			// no results ready to pull
			console.warn(`No 'resultId' exists for solution '${args.solutionId}'`);
			return null;
		}

		const filters = {
			variables: [],
			filters: []
		}

		const highlightFilter = createFilterFromHighlightRoot(args.highlightRoot, INCLUDE_FILTER);
		if (highlightFilter) {
			highlightFilter.name = getVarFromTarget(highlightFilter.name);
			filters.filters.push(highlightFilter);
		}

		// commit empty place holders, if there is no data
		return Promise.all(args.variables.map(variable => {
			// only use extrema if this is the feature variable
			let extremaMin = null;
			let extremaMax = null;
			if (variable.name === solution.feature) {
				extremaMin = args.extrema.min;
				extremaMax = args.extrema.max;
			}
			return axios.post(`/distil/results-variable-summary/${ES_INDEX}/${args.dataset}/${variable.name}/${extremaMin}/${extremaMax}/${solution.resultId}`, filters)
				.then(response => {
					mutations.updateHighlightSummaries(context, response.data.histogram);
				})
				.catch(error => {
					console.error(error);
				});
		}));
	},

	fetchPredictedHighlightSummaries(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, requestIds: string[], extrema: Extrema }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}

		const filters = {
			variables: [],
			filters: []
		}

		const highlightFilter = createFilterFromHighlightRoot(args.highlightRoot, INCLUDE_FILTER);
		if (highlightFilter) {
			highlightFilter.name = getVarFromTarget(highlightFilter.name);
			filters.filters.push(highlightFilter);
		}

		const solutions = getSolutionsByRequestIds(context.rootState.solutionModule, args.requestIds);

		const endPoint = `/distil/results-summary/${ES_INDEX}/${args.dataset}/${args.extrema.min}/${args.extrema.max}`
		const nameFunc = (p: SolutionInfo) => getPredictedCol(p.feature);
		const labelFunc = (p: SolutionInfo) => '';
		getSummaries(context, endPoint, solutions, nameFunc, labelFunc, mutations.updatePredictedHighlightSummaries, filters);
	},

	fetchCorrectnessHighlightSummaries(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, requestIds: string[], extrema: Extrema }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}

		const filters = {
			variables: [],
			filters: []
		}

		const highlightFilter = createFilterFromHighlightRoot(args.highlightRoot, INCLUDE_FILTER);
		if (highlightFilter) {
			highlightFilter.name = getVarFromTarget(highlightFilter.name);
			filters.filters.push(highlightFilter);
		}

		const solutions = getSolutionsByRequestIds(context.rootState.solutionModule, args.requestIds);

		const endPoint = `/distil/results-summary/${ES_INDEX}/${args.dataset}/${args.extrema.min}/${args.extrema.max}`
		const nameFunc = (p: SolutionInfo) => getCorrectnessCol(p.feature);
		const labelFunc = (p: SolutionInfo) => '';
		getSummaries(context, endPoint, solutions, nameFunc, labelFunc, updateCorrectnessHighlightSummary, filters);
	},

	fetchResultHighlightSamples(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, solutionId: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.solutionId) {
			console.warn('`solutionId` argument is missing');
			return null;
		}

		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		if (!solution.resultId) {
			// no results ready to pull
			return null;
		}

		const filters = {
			variables: [],
			filters: []
		}

		const highlightFilter = createFilterFromHighlightRoot(args.highlightRoot, INCLUDE_FILTER);
		if (highlightFilter) {
			highlightFilter.name = getVarFromTarget(highlightFilter.name);
			filters.filters.push(highlightFilter);
		}

		// fetch the data using the supplied filtered
		return context.dispatch('fetchResultTableData', {
				solutionId: args.solutionId,
				dataset: args.dataset,
				filters: filters
			})
			.then(res => {
				mutations.updateHighlightSamples(context, parseHighlightSamples(res.data));
			})
			.catch(error => {
				console.error(error);
				mutations.updateHighlightSamples(context, null);
			});
	},

	fetchResultHighlightValues(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, variables: Variable[], solutionId: string, requestIds: string[], extrema: Extrema }) {
		return Promise.all([
			context.dispatch('fetchResultHighlightSamples', {
				highlightRoot: args.highlightRoot,
				dataset: args.dataset,
				solutionId: args.solutionId
			}),
			context.dispatch('fetchResultHighlightSummaries', {
				highlightRoot: args.highlightRoot,
				dataset: args.dataset,
				variables: args.variables,
				solutionId: args.solutionId,
				extrema: args.extrema
			}),
			context.dispatch('fetchPredictedHighlightSummaries', {
				highlightRoot: args.highlightRoot,
				dataset: args.dataset,
				requestIds: args.requestIds,
				extrema: args.extrema
			}),
			context.dispatch('fetchCorrectnessHighlightSummaries', {
				highlightRoot: args.highlightRoot,
				dataset: args.dataset,
				requestIds: args.requestIds,
				extrema: args.extrema
			})
		]);
	}
}
