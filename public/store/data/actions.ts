import _ from 'lodash';
import axios from 'axios';
import { encodeQueryParams, Filter } from '../../util/filters';
import { getPipelineResultsOkay } from '../../util/pipelines';
import { getSummaries } from '../../util/data';
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
	getVariables(context: DataContext, args: { dataset: string }) {
		const dataset = args.dataset;
		return axios.get(`/distil/variables/${ES_INDEX}/${dataset}`)
			.then(response => {
				mutations.setVariables(context, response.data.variables);
			})
			.catch(error => {
				console.error(error);
				mutations.setVariables(context, []);
			});
	},

	setVariableType(context: DataContext, args: { dataset: string, field: string, type: string }) {
		return axios.post(`/distil/variables/${ES_INDEX}/${args.dataset}`,
			{
				field: args.field,
				type: args.type
			})
			.then(() => {
				mutations.updateVariableType(context, args);
				// update variable summary
				return context.dispatch('getVariableSummary', {
					dataset: args.dataset,
					variable: args.field
				});
			})
			.catch(error => {
				console.error(error);
			});
	},

	// fetches variable summary data for the given dataset and variables
	getVariableSummaries(context: DataContext, args: { dataset: string, variables: Variable[] }) {
		const dataset = args.dataset;
		const variables = args.variables;
		// commit empty place holders
		const histograms = variables.map(variable => {
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
		return Promise.all(variables.map(variable => {
			return context.dispatch('getVariableSummary', {
				dataset: dataset,
				variable: variable.name
			});
		}));
	},

	getVariableSummary(context: DataContext, args: { dataset: string, variable: string }) {
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
		const dataset = args.dataset;
		const filters = args.filters;
		const queryParams = encodeQueryParams(filters);
		const url = `distil/filtered-data/${ES_INDEX}/${dataset}/inclusive${queryParams}`;
		// request filtered data from server - no data is valid given filter settings
		return axios.get(url)
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
		const dataset = args.dataset;
		const filters = args.filters;
		const queryParams = encodeQueryParams(filters);
		const url = `distil/filtered-data/${ES_INDEX}/${dataset}/exclusive${queryParams}`;
		// request filtered data from server - no data is valid given filter settings
		return axios.get(url)
			.then(response => {
				mutations.setSelectedData(context, response.data);
			})
			.catch(error => {
				console.error(error);
				mutations.setSelectedData(context, {} as Data);
			});
	},

	// fetches result summaries for a given pipeline create request
	getResultsSummaries(context: DataContext, args: { dataset: string, requestId: string }) {
		const results = getPipelineResultsOkay(context.rootState.pipelineModule, args.requestId);
		const endPoint = `/distil/results-summary/${ES_INDEX}/${args.dataset}`
		const nameFunc = (p: PipelineInfo) => getPredictedFacetKey(p.feature);
		getSummaries(context, endPoint, results, nameFunc, mutations.setResultsSummaries, mutations.updateResultsSummaries);
	},

	// fetches result summaries for a given pipeline create request
	getResidualsSummaries(context: DataContext, args: { dataset: string, requestId: string }) {
		const results = getPipelineResultsOkay(context.rootState.pipelineModule, args.requestId);
		const endPoint = `/distil/residuals-summary/${ES_INDEX}/${args.dataset}`
		const nameFunc = (p: PipelineInfo) => getErrorFacetKey(p.feature);
		getSummaries(context, endPoint, results, nameFunc, mutations.setResidualsSummaries, mutations.updateResidualsSummaries);
	},

	// fetches result data for created pipeline
	updateResults(context: DataContext, args: { pipelineId: string, dataset: string, filters: Filter[] }) {
		const encodedPipelineId = encodeURIComponent(args.pipelineId);
		const filters = args.filters;
		const queryParams = encodeQueryParams(filters);
		return axios.get(`/distil/results/${ES_INDEX}/${args.dataset}/${encodedPipelineId}/inclusive${queryParams}`)
			.then(response => {
				mutations.setResultData(context, response.data);
			})
			.catch(error => {
				console.error(`Failed to fetch results from ${args.pipelineId} with error ${error}`);
			});
	}
}
