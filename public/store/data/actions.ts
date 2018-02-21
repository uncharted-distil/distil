import _ from 'lodash';
import axios from 'axios';
import { AxiosPromise } from 'axios';
import { encodeQueryParams, Filter } from '../../util/filters';
import { getPipelinesByRequestIds, getPipelineById } from '../../util/pipelines';
import { getSummaries, getSummary } from '../../util/data';
import { Variable, Data, Extrema } from './index';
import { PipelineInfo, PIPELINE_ERRORED } from '../pipelines/index';
import { mutations } from './module'
import { HighlightRoot, createFilterFromHighlightRoot, parseHighlightValues } from '../../util/highlights';
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

	exportProblem(context: DataContext, args: { dataset: string, target: string, filters: Filter[] }) {
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
		const filters = args.filters;
		const queryParams = encodeQueryParams(filters);
		return axios.post(`/distil/discovery/${ES_INDEX}/${args.dataset}/${args.target}${queryParams}`)
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
			context.dispatch('fetchVariableSummaries', {
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
		return axios.get(`/distil/variable-summaries/${ES_INDEX}/${args.dataset}/${args.variable}`)
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
	fetchResultSummaries(context: DataContext, args: { dataset: string, variables: Variable[], pipelineId: string, extrema: Extrema }) {
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
		if (args.variable === pipeline.feature) {
			extremaMin = args.extrema.min;
			extremaMax = args.extrema.max;
		}
		return axios.get(`/distil/results-variable-summary/${ES_INDEX}/${args.dataset}/${args.variable}/${extremaMin}/${extremaMax}/${pipeline.resultId}`)
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

	// update filtered data based on the  current filter state
	fetchFilteredTableData(context: DataContext, args: { dataset: string, filters: Filter[] }) {
		//mutations.setFilteredData(context, null);
		context.dispatch('fetchData', { dataset: args.dataset, filters: args.filters, inclusive: true })
			.then(response => {
				mutations.setFilteredData(context, response.data);
			})
			.catch(error => {
				console.error(error);
				mutations.setFilteredData(context, createEmptyData(args.dataset));
			});
	},

	// update filtered data based on the  current filter state
	fetchSelectedTableData(context: DataContext, args: { dataset: string, filters: Filter[] }) {
		//mutations.setSelectedData(context, null);
		context.dispatch('fetchData', { dataset: args.dataset, filters: args.filters, inclusive: false })
			.then(response => {
				mutations.setSelectedData(context, response.data);
			})
			.catch(error => {
				console.error(error);
				mutations.setSelectedData(context, createEmptyData(args.dataset));
			});
	},

	fetchData(context: DataContext, args: { dataset: string, filters: Filter[], inclusive: boolean }): AxiosPromise<Data> {
		const dataset = args.dataset;
		const filters = args.filters;
		const queryParams = encodeQueryParams(filters);
		const inclusiveStr = args.inclusive ? 'inclusive' : 'exclusive';
		const url = `distil/filtered-data/${ES_INDEX}/${dataset}/${inclusiveStr}${queryParams}`;
		// request filtered data from server - no data is valid given filter settings
		return axios.get<Data>(url);
	},

	fetchPredictedExtrema(context: DataContext, args: { dataset: string, pipelineId: string }) {
		const pipeline = getPipelineById(context.rootState.pipelineModule, args.pipelineId);
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
		const pipelines = getPipelinesByRequestIds(context.rootState.pipelineModule, args.requestIds);
		mutations.clearPredictedExtremas(context);
		return Promise.all(pipelines.map(pipeline => {
			return context.dispatch('fetchPredictedExtrema', {
				dataset: args.dataset,
				pipelineId: pipeline.pipelineId
			});
		}));
	},

	fetchResidualsExtrema(context: DataContext, args: { dataset: string, pipelineId: string }) {
		const pipeline = getPipelineById(context.rootState.pipelineModule, args.pipelineId);
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
		const pipelines = getPipelinesByRequestIds(context.rootState.pipelineModule, args.requestIds);
		mutations.clearResidualsExtremas(context);
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
		if (!args.extrema) {
			console.warn('`extrema` argument is missing');
			return null;
		}
		const pipeline = getPipelineById(context.rootState.pipelineModule, args.pipelineId);
		const endPoint = `/distil/results-summary/${ES_INDEX}/${args.dataset}/${args.extrema.min}/${args.extrema.max}`
		const nameFunc = (p: PipelineInfo) => getPredictedCol(p.feature);
		const labelFunc = (p: PipelineInfo) => 'Predicted';
		getSummary(context, endPoint, pipeline, nameFunc, labelFunc, mutations.updatePredictedSummaries);
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
		if (!args.extrema) {
			console.warn('`extrema` argument is missing');
			return null;
		}
		const pipelines = getPipelinesByRequestIds(context.rootState.pipelineModule, args.requestIds);
		const endPoint = `/distil/results-summary/${ES_INDEX}/${args.dataset}/${args.extrema.min}/${args.extrema.max}`
		const nameFunc = (p: PipelineInfo) => getPredictedCol(p.feature);
		const labelFunc = (p: PipelineInfo) => 'Predicted';
		getSummaries(context, endPoint, pipelines, nameFunc, labelFunc, mutations.updatePredictedSummaries);
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
		if (!args.extrema) {
			console.warn('`extrema` argument is missing');
			return null;
		}
		const pipeline = getPipelineById(context.rootState.pipelineModule, args.pipelineId);
		const endPoint = `/distil/residuals-summary/${ES_INDEX}/${args.dataset}/${args.extrema.min}/${args.extrema.max}`
		const nameFunc = (p: PipelineInfo) => getErrorCol(p.feature);
		const labelFunc = (p: PipelineInfo) => 'Error';
		getSummary(context, endPoint, pipeline, nameFunc, labelFunc, mutations.updateResidualsSummaries);
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
		getSummaries(context, endPoint, pipelines, nameFunc, labelFunc, mutations.updateResidualsSummaries);
	},

	// fetches result data for created pipeline
	fetchResultTableData(context: DataContext, args: { pipelineId: string, dataset: string, filters: Filter[] }) {
		//mutations.setResultData(context, null);
		context.dispatch('fetchResults', args)
			.then(response => {
				mutations.setResultData(context, response.data);
			})
			.catch(error => {
				console.error(`Failed to fetch results from ${args.pipelineId} with error ${error}`);
				mutations.setResultData(context, createEmptyData(args.dataset));
			});
	},

	fetchResults(context: DataContext, args: { pipelineId: string, dataset: string, filters: Filter[] }): AxiosPromise<Data> {
		const encodedPipelineId = encodeURIComponent(args.pipelineId);
		const filters = args.filters;
		const queryParams = encodeQueryParams(filters);
		return axios.get<Data>(`/distil/results/${ES_INDEX}/${args.dataset}/${encodedPipelineId}/inclusive${queryParams}`);
	},

	fetchDataHighlightValues(context: DataContext, args: { highlightRoot: HighlightRoot, dataset: string, filters: Filter[] }) {
		if (!args.highlightRoot) {
			mutations.setHighlightedValues(context, {});
			return null;
		}
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.filters) {
			console.warn('`filters` argument is missing');
			return null;
		}

		// if root is from table row, populate here and return
		if (_.isArray(args.highlightRoot.value)) {
			const highlightValues = args.highlightRoot.value;
			const values = {};
			highlightValues.forEach(value => {
				const col = value[0];
				const vals = value[1]
				values[col] = [ vals ];
			});
			mutations.setHighlightedValues(context, values);
			return;
		}

		const filtersCopy = args.filters.slice();
		const selectFilter = createFilterFromHighlightRoot(args.highlightRoot);

		const index = _.findIndex(filtersCopy, f => f.name === args.highlightRoot.key);
		if (index < 0) {
			filtersCopy.push(selectFilter);
		} else {
			filtersCopy[index] = selectFilter;
		}

		// fetch the data using the supplied filtered
		return context.dispatch('fetchData', {
				dataset: args.dataset,
				filters: filtersCopy,
				inclusive: true
			})
			.then(res => {
				mutations.setHighlightedValues(context, parseHighlightValues(res.data));
			})
			.catch(error => {
				console.error(error);
				mutations.setHighlightedValues(context, {});
			});
	},

	fetchResultHighlightValues(context: DataContext, args: { highlightRoot: HighlightRoot, dataset: string, filters: Filter[], pipelineId: string }) {
		if (!args.highlightRoot) {
			mutations.setHighlightedValues(context, {});
			return null;
		}
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.filters) {
			console.warn('`filters` argument is missing');
			return null;
		}
		if (!args.pipelineId) {
			console.warn('`pipelineId` argument is missing');
			return null;
		}

		// if root is from table row, populate here and return
		if (_.isArray(args.highlightRoot.value)) {
			const highlightValues = args.highlightRoot.value;
			const values = {};
			highlightValues.forEach(value => {
				const col = value[0];
				const vals = value[1]
				values[col] = [ vals ];
			});
			mutations.setHighlightedValues(context, values);
			return;
		}

		const filtersCopy = args.filters.slice();
		const selectFilter = createFilterFromHighlightRoot(args.highlightRoot);
		selectFilter.name = getVarFromTarget(selectFilter.name);

		const index = _.findIndex(filtersCopy, f => f.name === args.highlightRoot.key);
		if (index < 0) {
			filtersCopy.push(selectFilter);
		} else {
			filtersCopy[index] = selectFilter;
		}

		// fetch the data using the supplied filtered
		return context.dispatch('fetchResults', {
				pipelineId: args.pipelineId,
				dataset: args.dataset,
				filters: filtersCopy
			})
			.then(res => {
				mutations.setHighlightedValues(context, parseHighlightValues(res.data));
			})
			.catch(error => {
				console.error(error);
				mutations.setHighlightedValues(context, {});
			});
	},
}
