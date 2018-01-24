import _ from 'lodash';
import axios from 'axios';
import { AxiosPromise } from 'axios';
import { encodeQueryParams, Filter } from '../../util/filters';
import { getPipelinesByRequestIds, getPipelineById } from '../../util/pipelines';
import { getSummaries, getSummary } from '../../util/data';
import { Variable, Data } from './index';
import { PipelineInfo } from '../pipelines/index';
import { mutations } from './module'
import { DataContext, getPredictedFacetKey, getErrorFacetKey } from '../../util/data';

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
		// commit empty place holders
		const histograms = args.variables.map(variable => {
			return {
				name: variable.name,
				feature: name,
				pending: true,
				buckets: [],
				extrema: {
					min: NaN,
					max: NaN
				}
			};
		});
		mutations.setVariableSummaries(context, histograms);
		// fill them in asynchronously
		return Promise.all(args.variables.map(variable => {
			return context.dispatch('fetchVariableSummary', {
				dataset: args.dataset,
				variable: variable.name
			});
		}));
	},

	fetchVariableSummary(context: DataContext, args: { dataset: string, variable: string }) {
		const dataset = args.dataset;
		const variable = args.variable;
		return axios.get(`/distil/variable-summaries/${ES_INDEX}/${dataset}/${variable}`)
			.then(response => {
				// save the variable summary data
				const histogram = response.data.histogram;
				if (!histogram) {
					mutations.updateVariableSummaries(context, {
						name: variable,
						feature: '',
						buckets: [],
						extrema: {} as any,
						err: 'No analysis available'
					});
					return;
				}
				// ensure buckets is not nil
				mutations.updateVariableSummaries(context, histogram);
			})
			.catch(error => {
				console.error(error);
				mutations.updateVariableSummaries(context, {
					name: variable,
					feature: '',
					buckets: [],
					extrema: {} as any,
					err: error
				});
			});
	},

	// update filtered data based on the  current filter state
	updateFilteredData(context: DataContext, args: { dataset: string, filters: Filter[] }) {
		context.dispatch('fetchData', { dataset: args.dataset, filters: args.filters, inclusive: true })
			.then(response => {
				mutations.setFilteredData(context, response.data);
			})
			.catch(error => {
				console.error(error);
				mutations.setFilteredData(context, {} as Data);
			});
	},

	// update filtered data based on the  current filter state
	updateSelectedData(context: DataContext, args: { dataset: string, filters: Filter[] }) {
		context.dispatch('fetchData', { dataset: args.dataset, filters: args.filters, inclusive: false })
			.then(response => {
				mutations.setSelectedData(context, response.data);
			})
			.catch(error => {
				console.error(error);
				mutations.setSelectedData(context, {} as Data);
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

	// fetches result summary for a given pipeline id.
	fetchResultsSummary(context: DataContext, args: { dataset: string, pipelineId: string }) {
		const pipeline = getPipelineById(context.rootState.pipelineModule, args.pipelineId);
		const endPoint = `/distil/results-summary/${ES_INDEX}/${args.dataset}`
		const nameFunc = (p: PipelineInfo) => getPredictedFacetKey(p.feature);
		getSummary(context, endPoint, pipeline, nameFunc, mutations.updateResultsSummaries);
	},

	// fetches result summaries for a given pipeline create request
	fetchResultsSummaries(context: DataContext, args: { dataset: string, requestIds: string[] }) {
		const pipelines = getPipelinesByRequestIds(context.rootState.pipelineModule, args.requestIds);
		const endPoint = `/distil/results-summary/${ES_INDEX}/${args.dataset}`
		const nameFunc = (p: PipelineInfo) => getPredictedFacetKey(p.feature);
		getSummaries(context, endPoint, pipelines, nameFunc, mutations.updateResultsSummaries);
	},

	// fetches result summary for a given pipeline id.
	fetchResidualsSummary(context: DataContext, args: { dataset: string, pipelineId: string }) {
		const pipeline = getPipelineById(context.rootState.pipelineModule, args.pipelineId);
		const endPoint = `/distil/residuals-summary/${ES_INDEX}/${args.dataset}`
		const nameFunc = (p: PipelineInfo) => getErrorFacetKey(p.feature);
		getSummary(context, endPoint, pipeline, nameFunc, mutations.updateResidualsSummaries);
	},

	// fetches result summaries for a given pipeline create request
	fetchResidualsSummaries(context: DataContext, args: { dataset: string, requestIds: string[] }) {
		const pipelines = getPipelinesByRequestIds(context.rootState.pipelineModule, args.requestIds);
		const endPoint = `/distil/residuals-summary/${ES_INDEX}/${args.dataset}`
		const nameFunc = (p: PipelineInfo) => getErrorFacetKey(p.feature);
		getSummaries(context, endPoint, pipelines, nameFunc, mutations.updateResidualsSummaries);
	},

	// fetches result data for created pipeline
	updateResults(context: DataContext, args: { pipelineId: string, dataset: string, filters: Filter[] }) {
		context.dispatch('fetchResults', args)
			.then(response => {
				mutations.setResultData(context, response.data);
			})
			.catch(error => {
				console.error(`Failed to fetch results from ${args.pipelineId} with error ${error}`);
			});
	},

	fetchResults(context: DataContext, args: { pipelineId: string, dataset: string, filters: Filter[] }): AxiosPromise<Data> {
		const encodedPipelineId = encodeURIComponent(args.pipelineId);
		const filters = args.filters;
		const queryParams = encodeQueryParams(filters);
		return axios.get<Data>(`/distil/results/${ES_INDEX}/${args.dataset}/${encodedPipelineId}/inclusive${queryParams}`);
	}
}
