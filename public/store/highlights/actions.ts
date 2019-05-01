import axios from 'axios';
import { ActionContext } from 'vuex';
import { HighlightState, HighlightRoot } from './index';
import { DistilState } from '../store';
import { FilterParams, INCLUDE_FILTER } from '../../util/filters';
import { getSolutionsByRequestIds, getSolutionById } from '../../util/solutions';
import { getSummary } from '../../util/data';
import { Variable, Dataset } from '../dataset/index';
import { mutations } from './module';
import { addHighlightToFilterParams } from '../../util/highlights';
import { buildLookup } from '../../util/lookup';
import { Dictionary } from '../../util/dict';

export type HighlightsContext = ActionContext<HighlightState, DistilState>;

export const actions = {

	fetchDataHighlightSummaries(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, filterParams: FilterParams, variables: Variable[] }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.variables) {
			console.warn('`variables` argument is missing');
			return null;
		}

		const filterParams = addHighlightToFilterParams(args.filterParams, args.highlightRoot, INCLUDE_FILTER);

		// commit empty place holders, if there is no data
		return Promise.all(args.variables.map(variable => {
			return axios.post(`/distil/variable-summary/${args.dataset}/${variable.colName}`, filterParams)
				.then(response => {
					mutations.updateHighlightSummaries(context, response.data.histogram);
				})
				.catch(error => {
					console.error(error);
				});
		}));
	},

	fetchJoinDatasetsHighlightValues(context: HighlightsContext, args: { highlightRoot: HighlightRoot, datasets: Dataset[], filterParams: Dictionary<FilterParams>, variables: Variable[] }) {
		if (!args.datasets) {
			console.warn('`datasets` argument is missing');
			return null;
		}
		if (!args.variables) {
			console.warn('`variables` argument is missing');
			return null;
		}

		const datasetA = args.datasets[0];
		const datasetB = args.datasets[1];

		if (!datasetA || !datasetB) {
			return null;
		}

		const datasetALookup = buildLookup(datasetA.variables.map(v => v.colName));
		const datasetBLookup = buildLookup(datasetB.variables.map(v => v.colName));
		const datasetAVariables = args.variables.filter(v => datasetALookup[v.colName]);
		const datasetBVariables = args.variables.filter(v => datasetBLookup[v.colName]);

		return Promise.all([
			context.dispatch('fetchDataHighlightSummaries', {
				highlightRoot: args.highlightRoot && args.highlightRoot.dataset === datasetA.id ? args.highlightRoot : null,
				dataset: datasetA.id,
				variables: datasetAVariables,
				filterParams: args.filterParams[datasetA.id]
			}),
			context.dispatch('fetchDataHighlightSummaries', {
				highlightRoot:  args.highlightRoot && args.highlightRoot.dataset === datasetB.id ? args.highlightRoot : null,
				dataset: datasetB.id,
				variables: datasetBVariables,
				filterParams: args.filterParams[datasetB.id]
			})
		]);
	},

	fetchDataHighlightValues(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, filterParams: FilterParams, variables: Variable[] }) {
		return context.dispatch('fetchDataHighlightSummaries', {
			highlightRoot: args.highlightRoot,
			dataset: args.dataset,
			variables: args.variables,
			filterParams: args.filterParams
		});
	},

	fetchTrainingHighlightSummaries(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, solutionId: string, training: Variable[] }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.training) {
			console.warn('`variables` argument is missing');
			return null;
		}
		if (!args.solutionId) {
			console.warn('`solutionId` argument is missing');
			return null;
		}

		const timeseries = context.getters.getRouteTimeseriesAnalysis;
		if (timeseries) {
			// TODO: highlights?
			return;
		}

		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		if (!solution.resultId) {
			// no results ready to pull
			return null;
		}

		let filterParams = {
			variables: [],
			filters: []
		};
		filterParams = addHighlightToFilterParams(filterParams, args.highlightRoot, INCLUDE_FILTER);

		return Promise.all(args.training.map(variable => {
			return axios.post(`/distil/training-summary/${args.dataset}/${variable.colName}/${solution.resultId}`, filterParams)
				.then(response => {
					mutations.updateHighlightSummaries(context, response.data.histogram);
				})
				.catch(error => {
					console.error(error);
				});
		}));
	},

	fetchTargetHighlightSummary(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, target: string, solutionId: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.target) {
			console.warn('`target` argument is missing');
			return null;
		}
		if (!args.solutionId) {
			console.warn('`solutionId` argument is missing');
			return null;
		}

		const timeseries = context.getters.getRouteTimeseriesAnalysis;
		if (timeseries) {
			// TODO: highlights?
			return;
		}

		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		if (!solution.resultId) {
			// no results ready to pull
			return null;
		}

		let filterParams = {
			variables: [],
			filters: []
		};
		filterParams = addHighlightToFilterParams(filterParams, args.highlightRoot, INCLUDE_FILTER);

		return axios.post(`/distil/target-summary/${args.dataset}/${args.target}/${solution.resultId}`, filterParams)
			.then(response => {
				mutations.updateHighlightSummaries(context, response.data.histogram);
			})
			.catch(error => {
				console.error(error);
			});
	},

	fetchPredictedHighlightSummaries(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, target: string, requestIds: string[] }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.target) {
			console.warn('`target` argument is missing');
			return null;
		}
		if (!args.requestIds) {
			console.warn('`requestIds` argument is missing');
			return null;
		}

		const timeseries = context.getters.getRouteTimeseriesAnalysis;
		if (timeseries) {
			// TODO: highlights?
			return;
		}

		let filterParams = {
			variables: [],
			filters: []
		};
		filterParams = addHighlightToFilterParams(filterParams, args.highlightRoot, INCLUDE_FILTER);

		const solutions = getSolutionsByRequestIds(context.rootState.solutionModule, args.requestIds);
		const endpoint = `/distil/predicted-summary/${args.dataset}/${args.target}`;
		return Promise.all(solutions.map(solution => {
			const key = solution.predictedKey;
			const label = 'Predicted';
			return getSummary(context, endpoint, solution, key, label, mutations.updateHighlightSummaries, filterParams);
		}));
	},

	fetchResidualHighlightSummaries(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, target: string, requestIds: string[] }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.target) {
			console.warn('`target` argument is missing');
			return null;
		}
		if (!args.requestIds) {
			console.warn('`requestIds` argument is missing');
			return null;
		}

		let filterParams = {
			variables: [],
			filters: []
		};
		filterParams = addHighlightToFilterParams(filterParams, args.highlightRoot, INCLUDE_FILTER);

		const solutions = getSolutionsByRequestIds(context.rootState.solutionModule, args.requestIds);
		const endpoint = `/distil/residuals-summary/${args.dataset}/${args.target}`;

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
		if (!args.requestIds) {
			console.warn('`requestIds` argument is missing');
			return null;
		}

		let filterParams = {
			variables: [],
			filters: []
		};
		filterParams = addHighlightToFilterParams(filterParams, args.highlightRoot, INCLUDE_FILTER);

		const solutions = getSolutionsByRequestIds(context.rootState.solutionModule, args.requestIds);
		const endpoint = `/distil/correctness-summary/${args.dataset}`;

		return Promise.all(solutions.map(solution => {
			const key = solution.errorKey;
			const label = 'Error';
			return getSummary(context, endpoint, solution, key, label, mutations.updateHighlightSummaries, filterParams);
		}));
	},

	fetchResultHighlightValues(context: HighlightsContext, args: { highlightRoot: HighlightRoot, dataset: string, target: string, training: Variable[], solutionId: string, requestIds: string[], includeCorrectness: boolean, includeResidual: boolean }) {
		const ps = [
			context.dispatch('fetchTargetHighlightSummary', {
				highlightRoot: args.highlightRoot,
				dataset: args.dataset,
				target: args.target,
				solutionId: args.solutionId
			}),
			context.dispatch('fetchTrainingHighlightSummaries', {
				highlightRoot: args.highlightRoot,
				dataset: args.dataset,
				training: args.training,
				solutionId: args.solutionId
			}),
			context.dispatch('fetchPredictedHighlightSummaries', {
				highlightRoot: args.highlightRoot,
				dataset: args.dataset,
				target: args.target,
				requestIds: args.requestIds
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
					target: args.target,
					requestIds: args.requestIds
				}));
		}
		return Promise.all(ps);
	}
};
