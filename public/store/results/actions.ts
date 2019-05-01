import axios from 'axios';
import { ActionContext } from 'vuex';
import { DistilState } from '../store';
import { INCLUDE_FILTER, EXCLUDE_FILTER } from '../../util/filters';
import { getSolutionsByRequestIds, getSolutionById } from '../../util/solutions';
import { Variable } from '../dataset/index';
import { HighlightRoot } from '../highlights/index';
import { mutations } from './module';
import { ResultsState } from './index';
import { addHighlightToFilterParams } from '../../util/highlights';
import { getSummary, createPendingSummary, createErrorSummary, createEmptyTableData, fetchHistogramExemplars, getTimeseriesAnalysisIntervals } from '../../util/data';

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

		const promises = [];
		args.training.forEach(variable => {
			const key = variable.colName;
			const label = variable.colDisplayName;
			// add placeholder
			mutations.updateTrainingSummary(context, createPendingSummary(key, label, dataset, solutionId));
			// fetch summary
			promises.push(context.dispatch('fetchTrainingSummary', {
				dataset: dataset,
				variable: variable,
				resultID: solution.resultId
			}));
		});
		return Promise.all(promises);
	},

	fetchTrainingSummary(context: ResultsContext, args: { dataset: string, variable: Variable, resultID: string }): Promise<void>  {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.variable) {
			console.warn('`variable` argument is missing');
			return null;
		}
		if (!args.resultID) {
			console.warn('`resultID` argument is missing');
			return null;
		}

		const timeseries = context.getters.getRouteTimeseriesAnalysis;
		let interval = context.getters.getRouteTimeseriesBinningInterval;

		if (timeseries) {

			if (!interval) {
				const timeVar = context.getters.getTimeseriesAnalysisVariable;
				const range = context.getters.getTimeseriesAnalysisRange;
				const intervals = getTimeseriesAnalysisIntervals(timeVar, range);
				interval = intervals[0].value;
			}

			return axios.post(`distil/training-timeseries-summary/${args.dataset}/${timeseries}/${args.variable.colName}/${interval}/${args.resultID}`, {})
				.then(response => {
					const histogram = response.data.histogram;
					mutations.updateTrainingSummary(context, histogram);
				})
				.catch(error => {
					console.error(error);
					mutations.updateTrainingSummary(context, createErrorSummary(args.variable.colName, args.variable.colDisplayName, args.dataset, error));
				});
		}

		return axios.post(`/distil/training-summary/${args.dataset}/${args.variable.colName}/${args.resultID}`, {})
			.then(response => {
				const histogram = response.data.histogram;
				return fetchHistogramExemplars(args.dataset, args.variable.colName, histogram)
					.then(() => {
						mutations.updateTrainingSummary(context, histogram);
					});
			})
			.catch(error => {
				console.error(error);
				mutations.updateTrainingSummary(context, createErrorSummary(args.variable.colName, args.variable.colDisplayName, args.dataset, error));
			});
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

		const timeseries = context.getters.getRouteTimeseriesAnalysis;
		let interval = context.getters.getRouteTimeseriesBinningInterval;

		if (timeseries) {

			if (!interval) {
				const timeVar = context.getters.getTimeseriesAnalysisVariable;
				const range = context.getters.getTimeseriesAnalysisRange;
				const intervals = getTimeseriesAnalysisIntervals(timeVar, range);
				interval = intervals[0].value;
			}

			return axios.post(`distil/target-timeseries-summary/${args.dataset}/${timeseries}/${args.target}/${interval}/${solution.resultId}`, {})
				.then(response => {
					const histogram = response.data.histogram;
					mutations.updateTargetSummary(context, histogram);
				})
				.catch(error => {
					console.error(error);
					mutations.updateTargetSummary(context,  createErrorSummary(key, label, dataset, error));
				});
		}

		return axios.post(`/distil/target-summary/${args.dataset}/${args.target}/${solution.resultId}`, {})
			.then(response => {
				const histogram = response.data.histogram;
				return fetchHistogramExemplars(args.dataset, args.target, histogram)
					.then(() => {
						mutations.updateTargetSummary(context, histogram);
					});
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

		const timeseries = context.getters.getRouteTimeseriesAnalysis;
		let interval = context.getters.getRouteTimeseriesBinningInterval;

		if (timeseries) {

			if (!interval) {
				const timeVar = context.getters.getTimeseriesAnalysisVariable;
				const range = context.getters.getTimeseriesAnalysisRange;
				const intervals = getTimeseriesAnalysisIntervals(timeVar, range);
				interval = intervals[0].value;
			}

			const endPoint = `distil/forecasting-summary/${args.dataset}/${timeseries}/${args.target}/${interval}`;
			const key = solution.predictedKey;
			const label = 'Forecasted';
			return getSummary(context, endPoint, solution, key, label, mutations.updatePredictedSummaries, null);
		}

		const endpoint = `/distil/predicted-summary/${args.dataset}/${args.target}`;
		const key = solution.predictedKey;
		const label = 'Predicted';
		return getSummary(context, endpoint, solution, key, label, mutations.updatePredictedSummaries, null);
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

		const endPoint = `/distil/residuals-summary/${args.dataset}/${args.target}`;
		const key = solution.errorKey;
		const label = 'Error';
		return getSummary(context, endPoint, solution, key, label, mutations.updateResidualsSummaries, null);
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
		return getSummary(context, endPoint, solution, key, label, mutations.updateCorrectnessSummaries, null);
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

};
