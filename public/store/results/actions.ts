import _ from 'lodash';
import axios from 'axios';
import { ActionContext } from 'vuex';
import { DistilState } from '../store';
import { INCLUDE_FILTER, EXCLUDE_FILTER } from '../../util/filters';
import { getSolutionsByRequestIds, getSolutionById } from '../../util/solutions';
import { getSummaries, getSummary, updateCorrectnessSummary, getCorrectnessCol } from '../../util/data';
import { Variable, Extrema, ES_INDEX } from '../dataset/index';
import { HighlightRoot } from '../highlights/index';
import { Solution, SOLUTION_ERRORED } from '../solutions/index';
import { mutations } from './module'
import { ResultsState } from './index'
import { addHighlightToFilterParams } from '../../util/highlights';
import { getPredictedCol, getErrorCol, getVarFromTarget, createPendingSummary, createErrorSummary, createEmptyTableData} from '../../util/data';

export type ResultsContext = ActionContext<ResultsState, DistilState>;

export const actions = {

	// fetches variable summary data for the given dataset and variables
	fetchTrainingResultSummaries(context: ResultsContext, args: { dataset: string, variables: Variable[], solutionId: string, extrema: Extrema }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.variables) {
			console.warn('`variables` argument is missing');
			return null;
		}
		if (!args.solutionId) {
			console.warn('`solutionId` argument is missing');
			return null;
		}
		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		// commit empty place holders, if there is no data
		const promises = [];
		args.variables.forEach(variable => {
			const summary = _.find(context.state.resultSummaries, v => {
				return v.name === variable.name;
			});

			const name = variable.name;
			const label = variable.name;
			const dataset = args.dataset;

			if (solution.progress === SOLUTION_ERRORED) {
				mutations.updateResultSummaries(context, createErrorSummary(name, label, dataset, `No data available due to error`));
				return;
			}
			// update if none exists, or doesn't match latest resultId
			if (!summary || summary.resultId !== solution.resultId) {
				// add placeholder
				const solutionId = args.solutionId;
				mutations.updateResultSummaries(context, createPendingSummary(name, label, dataset, solutionId));
				// fetch summary
				promises.push(context.dispatch('fetchResultSummary', {
					dataset: args.dataset,
					solutionId: args.solutionId,
					variable: variable.name,
					extrema: args.extrema
				}));
			}
		});
		// fill them in asynchronously
		return Promise.all(promises);
	},

	fetchResultSummary(context: ResultsContext, args: { dataset: string, variable: string, solutionId: string, extrema: Extrema }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.variable) {
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
		// only use extrema if this is the feature variable
		let extremaMin = null;
		let extremaMax = null;
		if (args.variable === solution.feature && args.extrema) {
			extremaMin = args.extrema.min;
			extremaMax = args.extrema.max;
		}
		return axios.post(`/distil/results-variable-summary/${ES_INDEX}/${args.dataset}/${args.variable}/${extremaMin}/${extremaMax}/${solution.resultId}`, {})
			.then(response => {
				mutations.updateResultSummaries(context, response.data.histogram);
			})
			.catch(error => {
				console.error(error);
				const name = args.variable;
				const label = args.variable;
				const dataset = args.dataset;
				mutations.updateResultSummaries(context,  createErrorSummary(name, label, dataset, error));
			});
	},

	fetchResultExtrema(context: ResultsContext, args: { dataset: string, variable: string, solutionId: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.variable) {
			console.warn('`variable` argument is missing');
			return null;
		}
		if (!args.solutionId) {
			console.warn('`solutionId` argument is missing');
			return null;
		}

		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		if (!solution.resultId) {
			console.warn(`No 'resultId' exists for solution '${args.solutionId}'`);
			return null;
		}

		return axios.get(`/distil/results-variable-extrema/${ES_INDEX}/${args.dataset}/${args.variable}/${solution.resultId}`)
			.then(response => {
				mutations.updateResultExtrema(context, {
					extrema: response.data.extrema
				});
			})
			.catch(error => {
				console.error(error);
			});
	},

	fetchIncludedResultTableData(context: ResultsContext, args: { solutionId: string, dataset: string, highlightRoot: HighlightRoot }) {
		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		if (!solution.resultId) {
			// no results ready to pull
			console.warn(`No 'resultId' exists for solution '${args.solutionId}'`);
			return null;
		}

		let filterParams = {
			variables: [],
			filters: []
		};
		filterParams = addHighlightToFilterParams(context, filterParams, args.highlightRoot, INCLUDE_FILTER, getVarFromTarget);

		return axios.post(`/distil/results/${ES_INDEX}/${args.dataset}/${encodeURIComponent(args.solutionId)}`, filterParams)
			.then(response => {
				mutations.setIncludedResultTableData(context, response.data);
			})
			.catch(error => {
				console.error(`Failed to fetch results from ${args.solutionId} with error ${error}`);
				mutations.setIncludedResultTableData(context, createEmptyTableData(args.dataset));
			});
	},

	fetchExcludedResultTableData(context: ResultsContext, args: { solutionId: string, dataset: string, highlightRoot: HighlightRoot }) {
		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		if (!solution.resultId) {
			// no results ready to pull
			console.warn(`No 'resultId' exists for solution '${args.solutionId}'`);
			return null;
		}

		let filterParams = {
			variables: [],
			filters: []
		};
		filterParams = addHighlightToFilterParams(context, filterParams, args.highlightRoot, EXCLUDE_FILTER, getVarFromTarget);

		return axios.post(`/distil/results/${ES_INDEX}/${args.dataset}/${encodeURIComponent(args.solutionId)}`, filterParams)
			.then(response => {
				mutations.setExcludedResultTableData(context, response.data);
			})
			.catch(error => {
				console.error(`Failed to fetch results from ${args.solutionId} with error ${error}`);
				mutations.setExcludedResultTableData(context, createEmptyTableData(args.dataset));
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

	fetchPredictedExtrema(context: ResultsContext, args: { dataset: string, solutionId: string }) {
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
			console.warn(`No 'resultId' exists for solution '${args.solutionId}'`);
			return null;
		}

		return axios.get(`/distil/results-extrema/${ES_INDEX}/${args.dataset}/${solution.resultId}`)
			.then(response => {
				mutations.updatePredictedExtremas(context, {
					solutionId: args.solutionId,
					extrema: response.data.extrema
				});
			})
			.catch(error => {
				console.error(error);
			});
	},

	fetchPredictedExtremas(context: ResultsContext, args: { dataset: string, requestIds: string[] }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.requestIds) {
			console.warn('`requestIds` argument is missing');
			return null;
		}

		const solutions = getSolutionsByRequestIds(context.rootState.solutionModule, args.requestIds);
		return Promise.all(solutions.map(solution => {
			return context.dispatch('fetchPredictedExtrema', {
				dataset: args.dataset,
				solutionId: solution.solutionId
			});
		}));
	},

	fetchResidualsExtrema(context: ResultsContext, args: { dataset: string, solutionId: string }) {
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
			console.warn(`No 'resultId' exists for solution '${args.solutionId}'`);
			return null;
		}

		return axios.get(`/distil/residuals-extrema/${ES_INDEX}/${args.dataset}/${solution.resultId}`)
			.then(response => {
				mutations.updateResidualsExtremas(context, {
					solutionId: args.solutionId,
					extrema: response.data.extrema
				});
			})
			.catch(error => {
				console.error(error);
			});
	},

	fetchResidualsExtremas(context: ResultsContext, args: { dataset: string, requestIds: string[] }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.requestIds) {
			console.warn('`requestIds` argument is missing');
			return null;
		}

		const solutions = getSolutionsByRequestIds(context.rootState.solutionModule, args.requestIds);
		return Promise.all(solutions.map(solution => {
			return context.dispatch('fetchResidualsExtrema', {
				dataset: args.dataset,
				solutionId: solution.solutionId
			});
		}));
	},

	// fetches result summary for a given solution id.
	fetchPredictedSummary(context: ResultsContext, args: { dataset: string, solutionId: string, extrema: Extrema }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.solutionId) {
			console.warn('`solutionId` argument is missing');
			return null;
		}

		// only use extrema if this is the feature variable
		let extremaMin = null;
		let extremaMax = null;
		if (args.extrema) {
			extremaMin = args.extrema.min;
			extremaMax = args.extrema.max;
		}
		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		const endPoint = `/distil/results-summary/${ES_INDEX}/${args.dataset}/${extremaMin}/${extremaMax}`
		const nameFunc = (p: Solution) => getPredictedCol(p.feature);
		const labelFunc = (p: Solution) => 'Predicted';

		getSummary(context, endPoint, solution, nameFunc, labelFunc, mutations.updatePredictedSummaries, null);
	},

	// fetches result summaries for a given solution create request
	fetchPredictedSummaries(context: ResultsContext, args: { dataset: string, requestIds: string[], extrema: Extrema }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.requestIds) {
			console.warn('`requestIds` argument is missing');
			return null;
		}
		// only use extrema if this is the feature variable
		let extremaMin = null;
		let extremaMax = null;
		if (args.extrema) {
			extremaMin = args.extrema.min;
			extremaMax = args.extrema.max;
		}
		const solutions = getSolutionsByRequestIds(context.rootState.solutionModule, args.requestIds);
		const endPoint = `/distil/results-summary/${ES_INDEX}/${args.dataset}/${extremaMin}/${extremaMax}`
		const nameFunc = (p: Solution) => getPredictedCol(p.feature);
		const labelFunc = (p: Solution) => 'Predicted';
		getSummaries(context, endPoint, solutions, nameFunc, labelFunc, mutations.updatePredictedSummaries, null);
	},

	// fetches result summary for a given solution id.
	fetchResidualsSummary(context: ResultsContext, args: { dataset: string, solutionId: string, extrema: Extrema }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.solutionId) {
			console.warn('`solutionId` argument is missing');
			return null;
		}
		if (!args.extrema || (!args.extrema.min && !args.extrema.max)) {
			console.warn('`extrema` argument is missing');
			return null;
		}
		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		const endPoint = `/distil/residuals-summary/${ES_INDEX}/${args.dataset}/${args.extrema.min}/${args.extrema.max}`
		const nameFunc = (p: Solution) => getErrorCol(p.feature);
		const labelFunc = (p: Solution) => 'Error';
		getSummary(context, endPoint, solution, nameFunc, labelFunc, mutations.updateResidualsSummaries, null);
	},

	// fetches result summaries for a given solution create request
	fetchResidualsSummaries(context: ResultsContext, args: { dataset: string, requestIds: string[], extrema: Extrema }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.requestIds) {
			console.warn('`requestIds` argument is missing');
			return null;
		}
		if (!args.extrema || (!args.extrema.min && !args.extrema.max)) {
			console.warn('`extrema` argument is missing');
			return null;
		}
		const solutions = getSolutionsByRequestIds(context.rootState.solutionModule, args.requestIds);
		const endPoint = `/distil/residuals-summary/${ES_INDEX}/${args.dataset}/${args.extrema.min}/${args.extrema.max}`
		const nameFunc = (p: Solution) => getErrorCol(p.feature);
		const labelFunc = (p: Solution) => 'Error';
		getSummaries(context, endPoint, solutions, nameFunc, labelFunc, mutations.updateResidualsSummaries, null);
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

		// only use extrema if this is the feature variable
		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		const endPoint = `/distil/results-summary/${ES_INDEX}/${args.dataset}/null/null`;
		const nameFunc = (p: Solution) => getCorrectnessCol(p.feature);
		const labelFunc = (p: Solution) => 'Error Summary';

		getSummary(context, endPoint, solution, nameFunc, labelFunc, updateCorrectnessSummary, null);
	},

	// fetches result summaries for a given pipeline create request
	fetchCorrectnessSummaries(context: ResultsContext, args: { dataset: string, requestIds: string[]}) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.requestIds) {
			console.warn('`requestIds` argument is missing');
			return null;
		}
		// only use extrema if this is the feature variable
		const solutions = getSolutionsByRequestIds(context.rootState.solutionModule, args.requestIds);
		const endPoint = `/distil/results-summary/${ES_INDEX}/${args.dataset}/null/null`
		const nameFunc = (p: Solution) => getCorrectnessCol(p.feature);
		const labelFunc = (p: Solution) => 'Error Summary';
		getSummaries(context, endPoint, solutions, nameFunc, labelFunc, updateCorrectnessSummary, null);
	}

}
