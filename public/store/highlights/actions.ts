import axios from 'axios';
import { ActionContext } from 'vuex';
import { HighlightState } from './index';
import { DistilState } from '../store';
import { FilterParams, INCLUDE_FILTER } from '../../util/filters';
import { getSolutionsByRequestIds, getSolutionById } from '../../util/solutions';
import { getSummary } from '../../util/data';
import { Variable, Extrema, ES_INDEX } from '../dataset/index';
import { mutations } from './module'
import { HighlightRoot } from './index';
import { addHighlightToFilterParams, parseHighlightSamples } from '../../util/highlights';

export type HighlightsContext = ActionContext<HighlightState, DistilState>;

export const actions = {

	fetchDataHighlightSamples(context: HighlightsContext, args: { highlightRoot: HighlightRoot, filterParams: FilterParams, dataset: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.filterParams) {
			console.warn('`filters` argument is missing');
			return null;
		}

		const filterParams = addHighlightToFilterParams(context, args.filterParams, args.highlightRoot, INCLUDE_FILTER);

		// fetch the data using the supplied filtered
		return axios.post(`distil/data/${ES_INDEX}/${args.dataset}/false`, filterParams)
			.then(res => {
				mutations.updateHighlightSamples(context, parseHighlightSamples(res.data));
			})
			.catch(error => {
				console.error(error);
				mutations.updateHighlightSamples(context, null);
			});
	},

	fetchDataHighlightSummaries(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, filterParams: FilterParams, variables: Variable[] }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.variables) {
			console.warn('`variables` argument is missing');
			return null;
		}

		const filterParams = addHighlightToFilterParams(context, args.filterParams, args.highlightRoot, INCLUDE_FILTER);

		// commit empty place holders, if there is no data
		return Promise.all(args.variables.map(variable => {
			return axios.post(`/distil/variable-summary/${ES_INDEX}/${args.dataset}/${variable.key}`, filterParams)
				.then(response => {
					mutations.updateHighlightSummaries(context, response.data.histogram);
				})
				.catch(error => {
					console.error(error);
				});
		}));
	},

	fetchDataHighlightValues(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, filterParams: FilterParams, variables: Variable[] }) {
		return Promise.all([
			context.dispatch('fetchDataHighlightSamples', {
				highlightRoot: args.highlightRoot,
				dataset: args.dataset,
				filterParams: args.filterParams,
			}),
			context.dispatch('fetchDataHighlightSummaries', {
				highlightRoot: args.highlightRoot,
				dataset: args.dataset,
				variables: args.variables,
				filterParams: args.filterParams
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

		let filterParams = {
			variables: [],
			filters: []
		}
		filterParams = addHighlightToFilterParams(context, filterParams, args.highlightRoot, INCLUDE_FILTER);

		// commit empty place holders, if there is no data
		return Promise.all(args.variables.map(variable => {
			// only use extrema if this is the feature variable
			let extremaMin = null;
			let extremaMax = null;
			if (variable.key === solution.feature) {
				extremaMin = args.extrema.min;
				extremaMax = args.extrema.max;
			}
			return axios.post(`/distil/results-variable-summary/${ES_INDEX}/${args.dataset}/${variable.key}/${extremaMin}/${extremaMax}/${solution.resultId}`, filterParams)
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

		let filterParams = {
			variables: [],
			filters: []
		}
		filterParams = addHighlightToFilterParams(context, filterParams, args.highlightRoot, INCLUDE_FILTER);

		const solutions = getSolutionsByRequestIds(context.rootState.solutionModule, args.requestIds);
		const endpoint = `/distil/predicted-summary/${ES_INDEX}/${args.dataset}/${args.extrema.min}/${args.extrema.max}`

		return Promise.all(solutions.map(solution => {
			const key = solution.predictedKey;
			const label = 'Predicted';
			return getSummary(context, endpoint, solution, key, label, mutations.updateHighlightSummaries, filterParams);
		}));
	},

	fetchResidualHighlightSummaries(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, requestIds: string[], extrema: Extrema }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}

		let filterParams = {
			variables: [],
			filters: []
		}
		filterParams = addHighlightToFilterParams(context, filterParams, args.highlightRoot, INCLUDE_FILTER);

		const solutions = getSolutionsByRequestIds(context.rootState.solutionModule, args.requestIds);
		const endpoint = `/distil/residuals-summary/${ES_INDEX}/${args.dataset}/${args.extrema.min}/${args.extrema.max}`;

		return Promise.all(solutions.map(solution => {
			const key = solution.errorKey;
			const label = 'Error';
			return getSummary(context, endpoint, solution, key, label, mutations.updateHighlightSummaries, filterParams);
		}));
	},

	fetchCorrectnessHighlightSummaries(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, requestIds: string[]}) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}

		let filterParams = {
			variables: [],
			filters: []
		}
		filterParams = addHighlightToFilterParams(context, filterParams, args.highlightRoot, INCLUDE_FILTER);

		const solutions = getSolutionsByRequestIds(context.rootState.solutionModule, args.requestIds);
		const endpoint = `/distil/correctness-summary/${ES_INDEX}/${args.dataset}`

		return Promise.all(solutions.map(solution => {
			const key = solution.errorKey;
			const label = 'Error';
			return getSummary(context, endpoint, solution, key, label, mutations.updateHighlightSummaries, filterParams);
		}));
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

		let filterParams = {
			variables: [],
			filters: []
		}
		filterParams = addHighlightToFilterParams(context, filterParams, args.highlightRoot, INCLUDE_FILTER);

		// fetch the data using the supplied filtered
		return context.dispatch('fetchResultTableData', {
				solutionId: args.solutionId,
				dataset: args.dataset,
				filterParams: filterParams
			})
			.then(res => {
				mutations.updateHighlightSamples(context, parseHighlightSamples(res.data));
			})
			.catch(error => {
				console.error(error);
				mutations.updateHighlightSamples(context, null);
			});
	},

	fetchResultHighlightValues(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, variables: Variable[], solutionId: string, requestIds: string[], extrema: Extrema, includeCorrectness: boolean, includeResidual: boolean }) {
		const ps = [
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
			})
		];
		if (args.includeCorrectness) {
			ps.push(
				context.dispatch('fetchCorrectnessHighlightSummaries', {
					highlightRoot: args.highlightRoot,
					dataset: args.dataset,
					requestIds: args.requestIds
				}));
		}
		if (args.includeResidual) {
			ps.push(
				context.dispatch('fetchResidualHighlightSummaries', {
					highlightRoot: args.highlightRoot,
					dataset: args.dataset,
					requestIds: args.requestIds
				}));
		}
		return Promise.all(ps);
	}
}
