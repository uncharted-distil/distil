import axios from 'axios';
import { ActionContext } from 'vuex';
import { DistilState } from '../store';
import { INCLUDE_FILTER, EXCLUDE_FILTER } from '../../util/filters';
import { getSolutionsByRequestIds, getSolutionById } from '../../util/solutions';
import { Variable } from '../dataset/index';
import { HighlightRoot } from '../highlights/index';
import { mutations } from './module'
import { ResultsState } from './index'
import { addHighlightToFilterParams } from '../../util/highlights';
import { getSummary, createPendingSummary, createErrorSummary, createEmptyTableData} from '../../util/data';

export type ResultsContext = ActionContext<ResultsState, DistilState>;

export const actions = {

	// fetches variable summary data for the given dataset and variables
	fetchTrainingSummaries(context: ResultsContext, args: { dataset: string, training: Variable[], solutionId: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.training) {
			console.warn('`training` argument is missing');
			return null;
		}
		if (!args.solutionId) {
			console.warn('`solutionId` argument is missing');
			return null;
		}
		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		if (!solution.resultId) {
			// no results ready to pull
			return;
		}

		const dataset = args.dataset;
		const solutionId = args.solutionId;

		return Promise.all(args.training.map(variable => {
			const key = variable.colName;
			const label = variable.colDisplayName;

			mutations.updateTrainingSummary(context, createPendingSummary(key, label, dataset, solutionId));

			return axios.post(`/distil/training-summary/${dataset}/${key}/${solution.resultId}`, {})
				.then(response => {
					mutations.updateTrainingSummary(context, response.data.histogram);
				})
				.catch(error => {
					console.error(error);
					mutations.updateTrainingSummary(context,  createErrorSummary(key, label, dataset, error));
				});
		}));
	},

	fetchTargetSummary(context: ResultsContext, args: { dataset: string, target: string, solutionId: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.target) {
			console.warn('`variable` argument is missing');
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

		const key = args.target;
		const label = args.target;
		const dataset = args.dataset;

		mutations.updateTargetSummary(context, createPendingSummary(key, label, dataset, args.solutionId));

		return axios.post(`/distil/target-summary/${args.dataset}/${args.target}/${solution.resultId}`, {})
			.then(response => {
				mutations.updateTargetSummary(context, response.data.histogram);
			})
			.catch(error => {
				console.error(error);
				mutations.updateTargetSummary(context,  createErrorSummary(key, label, dataset, error));
			});
	},

	fetchIncludedResultTableData(context: ResultsContext, args: { solutionId: string, dataset: string, highlightRoot: HighlightRoot }) {
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

		return axios.post(`/distil/results/${args.dataset}/${encodeURIComponent(args.solutionId)}`, filterParams)
			.then(response => {
				mutations.setIncludedResultTableData(context, response.data);
			})
			.catch(error => {
				console.error(`Failed to fetch results from ${args.solutionId} with error ${error}`);
				mutations.setIncludedResultTableData(context, createEmptyTableData());
			});
	},

	fetchExcludedResultTableData(context: ResultsContext, args: { solutionId: string, dataset: string, highlightRoot: HighlightRoot }) {
		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		if (!solution.resultId) {
			// no results ready to pull
			return null;
		}

		let filterParams = {
			variables: [],
			filters: []
		};
		filterParams = addHighlightToFilterParams(filterParams, args.highlightRoot, EXCLUDE_FILTER);

		return axios.post(`/distil/results/${args.dataset}/${encodeURIComponent(args.solutionId)}`, filterParams)
			.then(response => {
				mutations.setExcludedResultTableData(context, response.data);
			})
			.catch(error => {
				console.error(`Failed to fetch results from ${args.solutionId} with error ${error}`);
				mutations.setExcludedResultTableData(context, createEmptyTableData());
			});
	},

	fetchResultTableData(context: ResultsContext, args: { solutionId: string, dataset: string, highlightRoot: HighlightRoot}) {
		return Promise.all([
			context.dispatch('fetchIncludedResultTableData', {
				dataset: args.dataset,
				solutionId: args.solutionId,
				highlightRoot: args.highlightRoot
			}),
			context.dispatch('fetchExcludedResultTableData', {
				dataset: args.dataset,
				solutionId: args.solutionId,
				highlightRoot: args.highlightRoot
			})
		]);
	},

	fetchResidualsExtrema(context: ResultsContext, args: { dataset: string, target: string, solutionId: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.target) {
			console.warn('`target` argument is missing');
			return null;
		}

		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		if (!solution.resultId) {
			// no results ready to pull
			return null;
		}

		return axios.get(`/distil/residuals-extrema/${args.dataset}/${args.target}`)
			.then(response => {
				mutations.updateResidualsExtrema(context, response.data.extrema);
			})
			.catch(error => {
				console.error(error);
			});
	},

	// fetches result summary for a given solution id.
	fetchPredictedSummary(context: ResultsContext, args: { dataset: string, target: string, solutionId: string }) {
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

		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		if (!solution.resultId) {
			// no results ready to pull
			return null;
		}

		const endpoint = `/distil/predicted-summary/${args.dataset}/${args.target}`
		const key = solution.predictedKey;
		const label = 'Predicted';
		getSummary(context, endpoint, solution, key, label, mutations.updatePredictedSummaries, null);
	},

	// fetches result summaries for a given solution create request
	fetchPredictedSummaries(context: ResultsContext, args: { dataset: string, target: string, requestIds: string[] }) {
		if (!args.requestIds) {
			console.warn('`requestIds` argument is missing');
			return null;
		}
		const solutions = getSolutionsByRequestIds(context.rootState.solutionModule, args.requestIds);
		return Promise.all(solutions.map(solution => {
			return context.dispatch('fetchPredictedSummary', {
				dataset: args.dataset,
				target: args.target,
				solutionId: solution.solutionId,
			});
		}));
	},

	// fetches result summary for a given solution id.
	fetchResidualsSummary(context: ResultsContext, args: { dataset: string, target: string, solutionId: string }) {
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

		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		if (!solution.resultId) {
			// no results ready to pull
			return null;
		}

		const endPoint = `/distil/residuals-summary/${args.dataset}/${args.target}`
		const key = solution.errorKey;
		const label = 'Error';
		getSummary(context, endPoint, solution, key, label, mutations.updateResidualsSummaries, null);
	},

	// fetches result summaries for a given solution create request
	fetchResidualsSummaries(context: ResultsContext, args: { dataset: string, target: string, requestIds: string[] }) {
		if (!args.requestIds) {
			console.warn('`requestIds` argument is missing');
			return null;
		}
		const solutions = getSolutionsByRequestIds(context.rootState.solutionModule, args.requestIds);
		return Promise.all(solutions.map(solution => {
			return context.dispatch('fetchResidualsSummary', {
				dataset: args.dataset,
				target: args.target,
				solutionId: solution.solutionId,
			});
		}));
	},

	// fetches result summary for a given pipeline id.
	fetchCorrectnessSummary(context: ResultsContext, args: { dataset: string, solutionId: string}) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.solutionId) {
			console.warn('`pipelineId` argument is missing');
			return null;
		}

		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		if (!solution.resultId) {
			// no results ready to pull
			return null;
		}

		const endPoint = `/distil/correctness-summary/${args.dataset}`;
		const key = solution.errorKey;
		const label = 'Error';
		getSummary(context, endPoint, solution, key, label, mutations.updateCorrectnessSummaries, null);
	},

	// fetches result summaries for a given pipeline create request
	fetchCorrectnessSummaries(context: ResultsContext, args: { dataset: string, requestIds: string[]}) {
		if (!args.requestIds) {
			console.warn('`requestIds` argument is missing');
			return null;
		}
		const solutions = getSolutionsByRequestIds(context.rootState.solutionModule, args.requestIds);
		return Promise.all(solutions.map(solution => {
			return context.dispatch('fetchCorrectnessSummary', {
				dataset: args.dataset,
				solutionId: solution.solutionId,
			});
		}));
	}

}
