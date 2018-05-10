import _ from 'lodash';
import axios from 'axios';
import { AxiosPromise } from 'axios';
import { FilterParams, INCLUDE_FILTER, EXCLUDE_FILTER } from '../../util/filters';
import { getSolutionsByRequestIds, getSolutionById } from '../../util/solutions';
import { getSummaries, getSummary, updateCorrectnessSummary, updateCorrectnessHighlightSummary, getCorrectnessCol } from '../../util/data';
import { Variable, Data, Extrema } from './index';
import { SolutionInfo, SOLUTION_ERRORED } from '../solutions/index';
import { mutations } from './module'
import { HighlightRoot } from './index';
import { createFilterFromHighlightRoot, parseHighlightSamples } from '../../util/highlights';
import { DataContext, getPredictedCol, getErrorCol, getVarFromTarget,
	createPendingSummary, createErrorSummary, createEmptyData} from '../../util/data';

export const ES_INDEX = 'datasets';

export const actions = {

	// searches dataset descriptions and column names for supplied terms
	searchDatasets(context: DataContext, terms: string) {
		const params = !_.isEmpty(terms) ? `?search=${terms}` : '';
		return axios.get(`/distil/datasets/${ES_INDEX}${params}`)
			.then(response => {
				mutations.setDatasets(context, response.data.datasets);
			})
			.catch(error => {
				console.error(error);
				mutations.setDatasets(context, []);
			});
	},

	// fetches all variables for a single dataset.
	fetchVariables(context: DataContext, args: { dataset: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		return axios.get(`/distil/variables/${ES_INDEX}/${args.dataset}`)
			.then(response => {
				mutations.setVariables(context, response.data.variables);
			})
			.catch(error => {
				console.error(error);
				mutations.setVariables(context, []);
			});
	},

	setVariableType(context: DataContext, args: { dataset: string, field: string, type: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.field) {
			console.warn('`field` argument is missing');
			return null;
		}
		if (!args.type) {
			console.warn('`type` argument is missing');
			return null;
		}
		return axios.post(`/distil/variables/${ES_INDEX}/${args.dataset}`,
			{
				field: args.field,
				type: args.type
			})
			.then(() => {
				mutations.updateVariableType(context, args);
				// update variable summary
				return context.dispatch('fetchVariableSummary', {
					dataset: args.dataset,
					variable: args.field
				});
			})
			.catch(error => {
				console.error(error);
			});
	},

	exportProblem(context: DataContext, args: { dataset: string, target: string, filters: FilterParams }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.target) {
			console.warn('`target` argument is missing');
			return null;
		}
		if (!args.filters) {
			console.warn('`filters` argument is missing');
			return null;
		}
		return axios.post(`/distil/discovery/${ES_INDEX}/${args.dataset}/${args.target}`, args.filters)
			.catch(error => {
				console.error(error);
			});
	},

	fetchVariablesAndVariableSummaries(context: DataContext, args: { dataset: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		return context.dispatch('fetchVariables', {
			dataset: args.dataset
		}).then(() => {
			return context.dispatch('fetchVariableSummaries', {
				dataset: args.dataset,
				variables: context.state.variables
			});
		});
	},

	// fetches variable summary data for the given dataset and variables
	fetchVariableSummaries(context: DataContext, args: { dataset: string, variables: Variable[] }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.variables) {
			console.warn('`variables` argument is missing');
			return null;
		}
		// commit empty place holders, if there is no data
		const promises = [];
		args.variables.forEach(variable => {
			const exists = _.find(context.state.variableSummaries, v => {
				return v.name === variable.name;
			});
			if (!exists) {
				// add placeholder
				const name = variable.name;
				const label = variable.name;
				const dataset = args.dataset;
				mutations.updateVariableSummaries(context,  createPendingSummary(name, label, dataset));
				// fetch summary
				promises.push(context.dispatch('fetchVariableSummary', {
					dataset: args.dataset,
					variable: variable.name
				}));
			}
		});
		// fill them in asynchronously
		return Promise.all(promises);
	},

	fetchVariableSummary(context: DataContext, args: { dataset: string, variable: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.variable) {
			console.warn('`variable` argument is missing');
			return null;
		}
		return axios.post(`/distil/variable-summary/${ES_INDEX}/${args.dataset}/${args.variable}`, {})
			.then(response => {
				mutations.updateVariableSummaries(context, response.data.histogram);
			})
			.catch(error => {
				console.error(error);
				const name = args.variable;
				const label = args.variable;
				const dataset = args.dataset;
				mutations.updateVariableSummaries(context,  createErrorSummary(name, label, dataset, error));
			});
	},

	// fetches variable summary data for the given dataset and variables
	fetchTrainingResultSummaries(context: DataContext, args: { dataset: string, variables: Variable[], solutionId: string, extrema: Extrema }) {
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

	fetchResultSummary(context: DataContext, args: { dataset: string, variable: string, solutionId: string, extrema: Extrema }) {
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

	fetchTargetResultExtrema(context: DataContext, args: { dataset: string, variable: string, solutionId: string }) {
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

		mutations.clearTargetResultExtrema(context);

		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		if (!solution.resultId) {
			console.warn(`No 'resultId' exists for solution '${args.solutionId}'`);
			return null;
		}

		return axios.get(`/distil/results-variable-extrema/${ES_INDEX}/${args.dataset}/${args.variable}/${solution.resultId}`)
			.then(response => {
				mutations.updateTargetResultExtrema(context, {
					extrema: response.data.extrema
				});
			})
			.catch(error => {
				console.error(error);
			});
	},

	// update filtered data based on the  current filter state
	fetchSelectedTableData(context: DataContext, args: { dataset: string, filters: FilterParams, highlightRoot: HighlightRoot }) {
		if (!args.filters) {
			console.warn('`filters` argument is missing');
			return null;
		}

		const highlightFilter = createFilterFromHighlightRoot(args.highlightRoot, INCLUDE_FILTER);
		if (highlightFilter) {
			args.filters.filters.push(highlightFilter);
		}

		mutations.setSelectedData(context, null);
		return context.dispatch('fetchData', { dataset: args.dataset, filters: args.filters, invert: false })
			.then(response => {
				mutations.setSelectedData(context, response.data);
			})
			.catch(error => {
				console.error(error);
				mutations.setSelectedData(context, createEmptyData(args.dataset));
			});
	},

	// update filtered data based on the  current filter state
	fetchExcludedTableData(context: DataContext, args: { dataset: string, filters: FilterParams, highlightRoot: HighlightRoot }) {
		if (!args.filters) {
			console.warn('`filters` argument is missing');
			return null;
		}

		const highlightFilter = createFilterFromHighlightRoot(args.highlightRoot, INCLUDE_FILTER);
		if (highlightFilter) {
			args.filters.filters.push(highlightFilter);
		}

		mutations.setExcludedData(context, null);

		return context.dispatch('fetchData', { dataset: args.dataset, filters: args.filters, invert: true })
			.then(response => {
				mutations.setExcludedData(context, response.data);
			})
			.catch(error => {
				console.error(error);
				mutations.setExcludedData(context, createEmptyData(args.dataset));
			});
	},


	fetchData(context: DataContext, args: { dataset: string, filters: FilterParams, invert: boolean }): AxiosPromise<Data> {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.filters) {
			console.warn('`filters` argument is missing');
			return null;
		}
		const invertStr = args.invert ? 'true' : 'false';
		// request filtered data from server - no data is valid given filter settings
		return axios.post(`distil/data/${ES_INDEX}/${args.dataset}/${invertStr}`, args.filters);
	},

	fetchPredictedExtrema(context: DataContext, args: { dataset: string, solutionId: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.solutionId) {
			console.warn('`solutionId` argument is missing');
			return null;
		}

		// clear extrema
		mutations.clearPredictedExtrema(context, args.solutionId);

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

	fetchPredictedExtremas(context: DataContext, args: { dataset: string, requestIds: string[] }) {
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

	fetchResidualsExtrema(context: DataContext, args: { dataset: string, solutionId: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.solutionId) {
			console.warn('`solutionId` argument is missing');
			return null;
		}

		// clear extrema
		mutations.clearResidualsExtrema(context, args.solutionId);

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

	fetchResidualsExtremas(context: DataContext, args: { dataset: string, requestIds: string[] }) {
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
	fetchPredictedSummary(context: DataContext, args: { dataset: string, solutionId: string, extrema: Extrema }) {
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
		const nameFunc = (p: SolutionInfo) => getPredictedCol(p.feature);
		const labelFunc = (p: SolutionInfo) => 'Predicted';

		getSummary(context, endPoint, solution, nameFunc, labelFunc, mutations.updatePredictedSummaries, null);
	},

	// fetches result summaries for a given solution create request
	fetchPredictedSummaries(context: DataContext, args: { dataset: string, requestIds: string[], extrema: Extrema }) {
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
		const nameFunc = (p: SolutionInfo) => getPredictedCol(p.feature);
		const labelFunc = (p: SolutionInfo) => 'Predicted';
		getSummaries(context, endPoint, solutions, nameFunc, labelFunc, mutations.updatePredictedSummaries, null);
	},

	// fetches result summary for a given solution id.
	fetchResidualsSummary(context: DataContext, args: { dataset: string, solutionId: string, extrema: Extrema }) {
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
		const nameFunc = (p: SolutionInfo) => getErrorCol(p.feature);
		const labelFunc = (p: SolutionInfo) => 'Error';
		getSummary(context, endPoint, solution, nameFunc, labelFunc, mutations.updateResidualsSummaries, null);
	},

	// fetches result summaries for a given solution create request
	fetchResidualsSummaries(context: DataContext, args: { dataset: string, requestIds: string[], extrema: Extrema }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.requestIds) {
			console.warn('`requestIds` argument is missing');
			return null;
		}
		if (!args.extrema) {
			console.warn('`extrema` argument is missing');
			return null;
		}
		const solutions = getSolutionsByRequestIds(context.rootState.solutionModule, args.requestIds);
		const endPoint = `/distil/residuals-summary/${ES_INDEX}/${args.dataset}/${args.extrema.min}/${args.extrema.max}`
		const nameFunc = (p: SolutionInfo) => getErrorCol(p.feature);
		const labelFunc = (p: SolutionInfo) => 'Error';
		getSummaries(context, endPoint, solutions, nameFunc, labelFunc, mutations.updateResidualsSummaries, null);
	},

	// fetches result summary for a given pipeline id.
	fetchCorrectnessSummary(context: DataContext, args: { dataset: string, solutionId: string}) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.solutionId) {
			console.warn('`pipelineId` argument is missing');
			return null;
		}

		// only use extrema if this is the feature variable
		const extremaMin = NaN;
		const extremaMax = NaN;
		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		const endPoint = `/distil/results-summary/${ES_INDEX}/${args.dataset}/${extremaMin}/${extremaMax}`
		const nameFunc = (p: SolutionInfo) => getCorrectnessCol(p.feature);
		const labelFunc = (p: SolutionInfo) => 'Error Summary';

		getSummary(context, endPoint, solution, nameFunc, labelFunc, updateCorrectnessSummary, null);
	},

	// fetches result summaries for a given pipeline create request
	fetchCorrectnessSummaries(context: DataContext, args: { dataset: string, requestIds: string[]}) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.requestIds) {
			console.warn('`requestIds` argument is missing');
			return null;
		}
		// only use extrema if this is the feature variable
		const extremaMin = NaN;
		const extremaMax = NaN;
		const solutions = getSolutionsByRequestIds(context.rootState.solutionModule, args.requestIds);
		const endPoint = `/distil/results-summary/${ES_INDEX}/${args.dataset}/${extremaMin}/${extremaMax}`
		const nameFunc = (p: SolutionInfo) => getCorrectnessCol(p.feature);
		const labelFunc = (p: SolutionInfo) => 'Error Summary';
		getSummaries(context, endPoint, solutions, nameFunc, labelFunc, updateCorrectnessSummary, null);
	},

	// fetches result data for created pipeline
	fetchResultTableData(context: DataContext, args: { solutionId: string, dataset: string, highlightRoot: HighlightRoot, filters?: FilterParams }) {
		return Promise.all([
			context.dispatch('fetchHighlightedResultTableData', {
				dataset: args.dataset,
				solutionId: args.solutionId,
				highlightRoot: args.highlightRoot
			}),
			context.dispatch('fetchUnhighlightedResultTableData', {
				dataset: args.dataset,
				solutionId: args.solutionId,
				highlightRoot: args.highlightRoot
			})
		]);
	},

	fetchHighlightedResultTableData(context: DataContext, args: { solutionId: string, dataset: string, highlightRoot: HighlightRoot, filters?: FilterParams }) {
		if (!args.filters) {
			args.filters = {
				variables: [],
				filters: []
			};
		}

		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		if (!solution.resultId) {
			// no results ready to pull
			console.warn(`No 'resultId' exists for solution '${args.solutionId}'`);
			return null;
		}

		const highlightFilter = createFilterFromHighlightRoot(args.highlightRoot, INCLUDE_FILTER);
		if (highlightFilter) {
			highlightFilter.name = getVarFromTarget(highlightFilter.name);
			args.filters.filters.push(highlightFilter);
		}

		mutations.setHighlightedResultData(context, null);
		return context.dispatch('fetchResults', args)
			.then(response => {
				mutations.setHighlightedResultData(context, response.data);
			})
			.catch(error => {
				console.error(`Failed to fetch results from ${args.solutionId} with error ${error}`);
				mutations.setHighlightedResultData(context, createEmptyData(args.dataset));
			});
	},

	fetchUnhighlightedResultTableData(context: DataContext, args: { solutionId: string, dataset: string, highlightRoot: HighlightRoot, filters?: FilterParams }) {
		if (!args.filters) {
			args.filters = {
				variables: [],
				filters: []
			};
		}

		const solution = getSolutionById(context.rootState.solutionModule, args.solutionId);
		if (!solution.resultId) {
			// no results ready to pull
			console.warn(`No 'resultId' exists for solution '${args.solutionId}'`);
			return null;
		}

		const highlightFilter = createFilterFromHighlightRoot(args.highlightRoot, EXCLUDE_FILTER);
		if (highlightFilter) {
			highlightFilter.name = getVarFromTarget(highlightFilter.name);
			args.filters.filters.push(highlightFilter);
		}

		mutations.setUnhighlightedResultData(context, null);
		return context.dispatch('fetchResults', args)
			.then(response => {
				mutations.setUnhighlightedResultData(context, response.data);
			})
			.catch(error => {
				console.error(`Failed to fetch results from ${args.solutionId} with error ${error}`);
				mutations.setUnhighlightedResultData(context, createEmptyData(args.dataset));
			});
	},

	fetchResults(context: DataContext, args: { solutionId: string, dataset: string, filters: FilterParams }): AxiosPromise<Data> {
		const encodedSolutionId = encodeURIComponent(args.solutionId);
		return axios.post(`/distil/results/${ES_INDEX}/${args.dataset}/${encodedSolutionId}`, args.filters);
	},

	fetchDataHighlightSamples(context: DataContext, args: { highlightRoot: HighlightRoot, filters: FilterParams, dataset: string }) {
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
		return context.dispatch('fetchData', {
				dataset: args.dataset,
				filters: args.filters,
				invert: false
			})
			.then(res => {
				mutations.updateHighlightSamples(context, parseHighlightSamples(res.data));
			})
			.catch(error => {
				console.error(error);
				mutations.updateHighlightSamples(context, null);
			});
	},

	fetchDataHighlightSummaries(context: DataContext, args: { highlightRoot: HighlightRoot, dataset: string, filters: FilterParams, variables: Variable[] }) {
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

	fetchDataHighlightValues(context: DataContext, args: { highlightRoot: HighlightRoot, dataset: string, filters: FilterParams, variables: Variable[] }) {

		// clear existing values
		mutations.clearHighlightSummaries(context);
		mutations.updateHighlightSamples(context, null);

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

	fetchResultHighlightSummaries(context: DataContext, args: { highlightRoot: HighlightRoot, dataset: string, solutionId: string, variables: Variable[], extrema: Extrema }) {
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

	fetchPredictedHighlightSummaries(context: DataContext, args: { highlightRoot: HighlightRoot, dataset: string, requestIds: string[], extrema: Extrema }) {
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

	fetchCorrectnessHighlightSummaries(context: DataContext, args: { highlightRoot: HighlightRoot, dataset: string, requestIds: string[], extrema: Extrema }) {
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

	fetchResultHighlightSamples(context: DataContext, args: { highlightRoot: HighlightRoot, dataset: string, solutionId: string }) {
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
		return context.dispatch('fetchResults', {
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

	fetchResultHighlightValues(context: DataContext, args: { highlightRoot: HighlightRoot, dataset: string, variables: Variable[], solutionId: string, requestIds: string[], extrema: Extrema }) {

		// clear existing values
		mutations.clearHighlightSummaries(context);
		mutations.updateHighlightSamples(context, null);

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
	},

	fetchImage(context: DataContext, args: { url: string }) {
		const IMAGES = [
			'a.jpeg',
			'b.jpeg',
			'c.jpeg'
		];
		return new Promise((resolve, reject) => {
			const image = new Image();
			image.onload = () => {
				mutations.setImage(context, { url: args.url, image: image });
				resolve(image);
			};
			image.onerror = (event: any) => {
				const err = new Error(`Unable to load image from URL: \`${event.path[0].currentSrc}\``);
				mutations.setImage(context, { url: args.url, err: err });
				reject(err);
			};
			image.crossOrigin = 'anonymous';
			image.src = `images/${IMAGES[Math.floor(Math.random() * IMAGES.length)]}`;
			//image.src = `images/${args.url}`;
		});

	}
}
