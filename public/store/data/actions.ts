import _ from 'lodash';
import axios from 'axios';
import { AxiosPromise } from 'axios';
import { FilterParams, INCLUDE_FILTER, EXCLUDE_FILTER } from '../../util/filters';
import { getPipelinesByRequestIds, getPipelineById } from '../../util/pipelines';
import { getSummaries, getSummary, updatePredictedSummary, updatePredictedHighlightSummary } from '../../util/data';
import { Variable, Data, Extrema } from './index';
import { PipelineInfo, PIPELINE_ERRORED } from '../pipelines/index';
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
	fetchTrainingResultSummaries(context: DataContext, args: { dataset: string, variables: Variable[], pipelineId: string, extrema: Extrema }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.variables) {
			console.warn('`variables` argument is missing');
			return null;
		}
		if (!args.pipelineId) {
			console.warn('`pipelineId` argument is missing');
			return null;
		}
		const pipeline = getPipelineById(context.rootState.pipelineModule, args.pipelineId);
		// commit empty place holders, if there is no data
		const promises = [];
		args.variables.forEach(variable => {
			const summary = _.find(context.state.resultSummaries, v => {
				return v.name === variable.name;
			});

			const name = variable.name;
			const label = variable.name;
			const dataset = args.dataset;

			if (pipeline.progress === PIPELINE_ERRORED) {
				mutations.updateResultSummaries(context, createErrorSummary(name, label, dataset, `No data available due to error`));
				return;
			}
			// update if none exists, or doesn't match latest resultId
			if (!summary || summary.resultId !== pipeline.resultId) {
				// add placeholder
				const pipelineId = args.pipelineId;
				mutations.updateResultSummaries(context, createPendingSummary(name, label, dataset, pipelineId));
				// fetch summary
				promises.push(context.dispatch('fetchResultSummary', {
					dataset: args.dataset,
					pipelineId: args.pipelineId,
					variable: variable.name,
					extrema: args.extrema
				}));
			}
		});
		// fill them in asynchronously
		return Promise.all(promises);
	},

	fetchResultSummary(context: DataContext, args: { dataset: string, variable: string, pipelineId: string, extrema: Extrema }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.variable) {
			console.warn('`variable` argument is missing');
			return null;
		}
		if (!args.pipelineId) {
			console.warn('`pipelineId` argument is missing');
			return null;
		}
		const pipeline = getPipelineById(context.rootState.pipelineModule, args.pipelineId);
		if (!pipeline.resultId) {
			// no results ready to pull
			return null;
		}
		// only use extrema if this is the feature variable
		let extremaMin = null;
		let extremaMax = null;
		if (args.variable === pipeline.feature && args.extrema) {
			extremaMin = args.extrema.min;
			extremaMax = args.extrema.max;
		}
		return axios.post(`/distil/results-variable-summary/${ES_INDEX}/${args.dataset}/${args.variable}/${extremaMin}/${extremaMax}/${pipeline.resultId}`, {})
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

	fetchTargetResultExtrema(context: DataContext, args: { dataset: string, variable: string, pipelineId: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.variable) {
			console.warn('`variable` argument is missing');
			return null;
		}
		if (!args.pipelineId) {
			console.warn('`pipelineId` argument is missing');
			return null;
		}

		mutations.clearTargetResultExtrema(context);

		const pipeline = getPipelineById(context.rootState.pipelineModule, args.pipelineId);
		if (!pipeline.resultId) {
			console.warn(`No 'resultId' exists for pipeline '${args.pipelineId}'`);
			return null;
		}

		return axios.get(`/distil/results-variable-extrema/${ES_INDEX}/${args.dataset}/${args.variable}/${pipeline.resultId}`)
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

	fetchPredictedExtrema(context: DataContext, args: { dataset: string, pipelineId: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.pipelineId) {
			console.warn('`pipelineId` argument is missing');
			return null;
		}

		// clear extrema
		mutations.clearPredictedExtrema(context, args.pipelineId);

		const pipeline = getPipelineById(context.rootState.pipelineModule, args.pipelineId);
		if (!pipeline.resultId) {
			console.warn(`No 'resultId' exists for pipeline '${args.pipelineId}'`);
			return null;
		}

		return axios.get(`/distil/results-extrema/${ES_INDEX}/${args.dataset}/${pipeline.resultId}`)
			.then(response => {
				mutations.updatePredictedExtremas(context, {
					pipelineId: args.pipelineId,
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

		const pipelines = getPipelinesByRequestIds(context.rootState.pipelineModule, args.requestIds);
		return Promise.all(pipelines.map(pipeline => {
			return context.dispatch('fetchPredictedExtrema', {
				dataset: args.dataset,
				pipelineId: pipeline.pipelineId
			});
		}));
	},

	fetchResidualsExtrema(context: DataContext, args: { dataset: string, pipelineId: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.pipelineId) {
			console.warn('`pipelineId` argument is missing');
			return null;
		}

		// clear extrema
		mutations.clearResidualsExtrema(context, args.pipelineId);

		const pipeline = getPipelineById(context.rootState.pipelineModule, args.pipelineId);
		if (!pipeline.resultId) {
			console.warn(`No 'resultId' exists for pipeline '${args.pipelineId}'`);
			return null;
		}

		return axios.get(`/distil/residuals-extrema/${ES_INDEX}/${args.dataset}/${pipeline.resultId}`)
			.then(response => {
				mutations.updateResidualsExtremas(context, {
					pipelineId: args.pipelineId,
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

		const pipelines = getPipelinesByRequestIds(context.rootState.pipelineModule, args.requestIds);
		return Promise.all(pipelines.map(pipeline => {
			return context.dispatch('fetchResidualsExtrema', {
				dataset: args.dataset,
				pipelineId: pipeline.pipelineId
			});
		}));
	},

	// fetches result summary for a given pipeline id.
	fetchPredictedSummary(context: DataContext, args: { dataset: string, pipelineId: string, extrema: Extrema }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.pipelineId) {
			console.warn('`pipelineId` argument is missing');
			return null;
		}



		// only use extrema if this is the feature variable
		let extremaMin = null;
		let extremaMax = null;
		if (args.extrema) {
			extremaMin = args.extrema.min;
			extremaMax = args.extrema.max;
		}
		const pipeline = getPipelineById(context.rootState.pipelineModule, args.pipelineId);
		const endPoint = `/distil/results-summary/${ES_INDEX}/${args.dataset}/${extremaMin}/${extremaMax}`
		const nameFunc = (p: PipelineInfo) => getPredictedCol(p.feature);
		const labelFunc = (p: PipelineInfo) => 'Predicted';

		getSummary(context, endPoint, pipeline, nameFunc, labelFunc, updatePredictedSummary, null);
	},

	// fetches result summaries for a given pipeline create request
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
		const pipelines = getPipelinesByRequestIds(context.rootState.pipelineModule, args.requestIds);
		const endPoint = `/distil/results-summary/${ES_INDEX}/${args.dataset}/${extremaMin}/${extremaMax}`
		const nameFunc = (p: PipelineInfo) => getPredictedCol(p.feature);
		const labelFunc = (p: PipelineInfo) => 'Predicted';
		getSummaries(context, endPoint, pipelines, nameFunc, labelFunc, updatePredictedSummary, null);
	},

	// fetches result summary for a given pipeline id.
	fetchResidualsSummary(context: DataContext, args: { dataset: string, pipelineId: string, extrema: Extrema }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.pipelineId) {
			console.warn('`pipelineId` argument is missing');
			return null;
		}
		if (!args.extrema || (!args.extrema.min && !args.extrema.max)) {
			console.warn('`extrema` argument is missing');
			return null;
		}
		const pipeline = getPipelineById(context.rootState.pipelineModule, args.pipelineId);
		const endPoint = `/distil/residuals-summary/${ES_INDEX}/${args.dataset}/${args.extrema.min}/${args.extrema.max}`
		const nameFunc = (p: PipelineInfo) => getErrorCol(p.feature);
		const labelFunc = (p: PipelineInfo) => 'Error';
		getSummary(context, endPoint, pipeline, nameFunc, labelFunc, mutations.updateResidualsSummaries, null);
	},

	// fetches result summaries for a given pipeline create request
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
		const pipelines = getPipelinesByRequestIds(context.rootState.pipelineModule, args.requestIds);
		const endPoint = `/distil/residuals-summary/${ES_INDEX}/${args.dataset}/${args.extrema.min}/${args.extrema.max}`
		const nameFunc = (p: PipelineInfo) => getErrorCol(p.feature);
		const labelFunc = (p: PipelineInfo) => 'Error';
		getSummaries(context, endPoint, pipelines, nameFunc, labelFunc, mutations.updateResidualsSummaries, null);
	},

	// fetches result data for created pipeline
	fetchResultTableData(context: DataContext, args: { pipelineId: string, dataset: string, highlightRoot: HighlightRoot, filters?: FilterParams }) {
		return Promise.all([
			context.dispatch('fetchHighlightedResultTableData', {
				dataset: args.dataset,
				pipelineId: args.pipelineId,
				highlightRoot: args.highlightRoot
			}),
			context.dispatch('fetchUnhighlightedResultTableData', {
				dataset: args.dataset,
				pipelineId: args.pipelineId,
				highlightRoot: args.highlightRoot
			})
		]);
	},

	fetchHighlightedResultTableData(context: DataContext, args: { pipelineId: string, dataset: string, highlightRoot: HighlightRoot, filters?: FilterParams }) {
		if (!args.filters) {
			args.filters = {
				variables: [],
				filters: []
			};
		}

		const pipeline = getPipelineById(context.rootState.pipelineModule, args.pipelineId);
		if (!pipeline.resultId) {
			// no results ready to pull
			console.warn(`No 'resultId' exists for pipeline '${args.pipelineId}'`);
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
				console.error(`Failed to fetch results from ${args.pipelineId} with error ${error}`);
				mutations.setHighlightedResultData(context, createEmptyData(args.dataset));
			});
	},

	fetchUnhighlightedResultTableData(context: DataContext, args: { pipelineId: string, dataset: string, highlightRoot: HighlightRoot, filters?: FilterParams }) {
		if (!args.filters) {
			args.filters = {
				variables: [],
				filters: []
			};
		}

		const pipeline = getPipelineById(context.rootState.pipelineModule, args.pipelineId);
		if (!pipeline.resultId) {
			// no results ready to pull
			console.warn(`No 'resultId' exists for pipeline '${args.pipelineId}'`);
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
				console.error(`Failed to fetch results from ${args.pipelineId} with error ${error}`);
				mutations.setUnhighlightedResultData(context, createEmptyData(args.dataset));
			});
	},

	fetchResults(context: DataContext, args: { pipelineId: string, dataset: string, filters: FilterParams }): AxiosPromise<Data> {
		const encodedPipelineId = encodeURIComponent(args.pipelineId);
		return axios.post(`/distil/results/${ES_INDEX}/${args.dataset}/${encodedPipelineId}`, args.filters);
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

	fetchResultHighlightSummaries(context: DataContext, args: { highlightRoot: HighlightRoot, dataset: string, pipelineId: string, variables: Variable[], extrema: Extrema }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.variables) {
			console.warn('`variables` argument is missing');
			return null;
		}

		const pipeline = getPipelineById(context.rootState.pipelineModule, args.pipelineId);
		if (!pipeline.resultId) {
			// no results ready to pull
			console.warn(`No 'resultId' exists for pipeline '${args.pipelineId}'`);
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
			if (variable.name === pipeline.feature) {
				extremaMin = args.extrema.min;
				extremaMax = args.extrema.max;
			}
			return axios.post(`/distil/results-variable-summary/${ES_INDEX}/${args.dataset}/${variable.name}/${extremaMin}/${extremaMax}/${pipeline.resultId}`, filters)
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

		const pipelines = getPipelinesByRequestIds(context.rootState.pipelineModule, args.requestIds);

		const endPoint = `/distil/results-summary/${ES_INDEX}/${args.dataset}/${args.extrema.min}/${args.extrema.max}`
		const nameFunc = (p: PipelineInfo) => getPredictedCol(p.feature);
		const labelFunc = (p: PipelineInfo) => '';
		getSummaries(context, endPoint, pipelines, nameFunc, labelFunc, updatePredictedHighlightSummary, filters);
	},

	fetchResultHighlightSamples(context: DataContext, args: { highlightRoot: HighlightRoot, dataset: string, pipelineId: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.pipelineId) {
			console.warn('`pipelineId` argument is missing');
			return null;
		}

		const pipeline = getPipelineById(context.rootState.pipelineModule, args.pipelineId);
		if (!pipeline.resultId) {
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
				pipelineId: args.pipelineId,
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

	fetchResultHighlightValues(context: DataContext, args: { highlightRoot: HighlightRoot, dataset: string, variables: Variable[], pipelineId: string, requestIds: string[], extrema: Extrema }) {

		// clear existing values
		mutations.clearHighlightSummaries(context);
		mutations.updateHighlightSamples(context, null);

		return Promise.all([
			context.dispatch('fetchResultHighlightSamples', {
				highlightRoot: args.highlightRoot,
				dataset: args.dataset,
				pipelineId: args.pipelineId
			}),
			context.dispatch('fetchResultHighlightSummaries', {
				highlightRoot: args.highlightRoot,
				dataset: args.dataset,
				variables: args.variables,
				pipelineId: args.pipelineId,
				extrema: args.extrema
			}),
			context.dispatch('fetchPredictedHighlightSummaries', {
				highlightRoot: args.highlightRoot,
				dataset: args.dataset,
				requestIds: args.requestIds,
				extrema: args.extrema
			})
		]);
	}
}
