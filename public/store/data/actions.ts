import _ from 'lodash';
import axios from 'axios';
import { encodeQueryParams, FilterMap } from '../../util/filters';
import { getPipelineResults } from '../../util/pipelines';
import { DataState, Variable } from './index';
import { DistilState } from '../store';
import { ActionContext } from 'vuex';

const ES_INDEX = 'datasets';

export type DataContext = ActionContext<DataState, DistilState>;

export const actions = {

	// searches dataset descriptions and column names for supplied terms
	searchDatasets(context: DataContext, terms: string) {
		const params = !_.isEmpty(terms) ? `?search=${terms}` : '';
		return axios.get(`/distil/datasets/${ES_INDEX}${params}`)
			.then(response => {
				context.commit('setDatasets', response.data.datasets);
			})
			.catch(error => {
				console.error(error);
				context.commit('setDatasets', []);
			});
	},

	// fetches all variables for a single dataset.
	getVariables(context: DataContext, args: { dataset: string }) {
		const dataset = args.dataset;
		return axios.get(`/distil/variables/${ES_INDEX}/${dataset}`)
			.then(response => {
				context.commit('setVariables', response.data.variables);
			})
			.catch(error => {
				console.error(error);
				context.commit('setVariables', []);
			});
	},

	setVariableType(context: DataContext, args: { dataset: string, field: string, type: string }) {
		return axios.post(`/distil/variables/${ES_INDEX}/${args.dataset}`,
			{
				field: args.field,
				type: args.type
			})
			.then(() => {
				context.commit('updateVariableType', args);
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
		context.commit('setResultsSummaries', histograms);
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
					context.commit('updateVariableSummaries', {
						name: variable,
						err: 'No analysis available'
					});
					return;
				}
				// ensure buckets is not nil
				context.commit('updateVariableSummaries', histogram);
			})
			.catch(error => {
				console.error(error);
				context.commit('updateVariableSummaries', {
					name: variable,
					err: error
				});
			});
	},

	// update filtered data based on the  current filter state
	updateFilteredData(context: DataContext, args: { dataset: string, filters: FilterMap }) {
		const dataset = args.dataset;
		const filters = args.filters;
		const queryParams = encodeQueryParams(filters);
		const url = `distil/filtered-data/${ES_INDEX}/${dataset}/inclusive${queryParams}`;
		// request filtered data from server - no data is valid given filter settings
		return axios.get(url)
			.then(response => {
				context.commit('setFilteredData', response.data);
			})
			.catch(error => {
				console.error(error);
				context.commit('setFilteredData', []);
			});
	},

	// update filtered data based on the  current filter state
	updateSelectedData(context: DataContext, args: { dataset: string, filters: FilterMap }) {
		const dataset = args.dataset;
		const filters = args.filters;
		const queryParams = encodeQueryParams(filters);
		const url = `distil/filtered-data/${ES_INDEX}/${dataset}/exclusive${queryParams}`;
		// request filtered data from server - no data is valid given filter settings
		return axios.get(url)
			.then(response => {
				context.commit('setSelectedData', response.data);
			})
			.catch(error => {
				console.error(error);
				context.commit('setSelectedData', []);
			});
	},

	// fetches result summaries for a given pipeline create request
	getResultsSummaries(context: DataContext, args: { dataset: string, requestId: string }) {
		const dataset = args.dataset;
		const requestId = args.requestId;
		const results = getPipelineResults(context.rootState.pipelineModule, requestId);

		// save a placeholder histogram
		const pendingHistograms = _.map(results, r => {
			return {
				name: r.name,
				pending: true
			};
		});
		context.commit('setResultsSummaries', pendingHistograms);

		// fetch the results for each pipeline
		for (var result of results) {
			const name = result.name;
			const feature = result.feature;
			const pipelineId = result.pipelineId;
			const res = encodeURIComponent(result.pipeline.resultId);
			axios.get(`/distil/results-summary/${ES_INDEX}/${dataset}/${res}`)
				.then(response => {
					// save the histogram data
					const histogram = response.data.histogram;
					if (!histogram) {
						context.commit('setResultsSummaries', [
							{
								name: name,
								feature: feature,
								pipelineId: pipelineId,
								err: 'No analysis available'
							}
						]);
						return;
					}
					// ensure buckets is not nil
					histogram.buckets = histogram.buckets ? histogram.buckets : [];
					histogram.name = name;
					histogram.feature = feature;
					histogram.pipelineId = pipelineId;
					context.commit('updateResultsSummaries', histogram);
				})
				.catch(error => {
					context.commit('setResultsSummaries', [
						{
							name: name,
							feature: feature,
							pipelineId: pipelineId,
							err: error
						}
					]);
					return;
				});
		}
	},

	// fetches result data for created pipeline
	updateResults(context: DataContext, args: { resultId: string, dataset: string, filters: FilterMap }) {
		const encodedResultId = encodeURIComponent(args.resultId);
		const filters = args.filters;
		const queryParams = encodeQueryParams(filters);
		return axios.get(`/distil/results/${ES_INDEX}/${args.dataset}/${encodedResultId}/inclusive${queryParams}`)
			.then(response => {
				context.commit('setResultData', response.data);
			})
			.catch(error => {
				console.error(`Failed to fetch results from ${args.resultId} with error ${error}`);
			});
	},

	highlightFeatureRange(context: DataContext, highlight: Range) {
		context.commit('highlightFeatureRange', highlight);
	},

	clearFeatureHighlightRange(context: DataContext, varName: string) {
		context.commit('clearFeatureHighlightRange', varName);
	},

	highlightFeatureValues(context: DataContext, highlight: { [name: string]: any }) {
		context.commit('highlightFeatureValues', highlight);
	},

	clearFeatureHighlightValues(context: DataContext) {
		context.commit('clearFeatureHighlightValues');
	}
}
