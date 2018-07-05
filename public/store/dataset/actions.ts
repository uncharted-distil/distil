import _ from 'lodash';
import axios from 'axios';
import { ActionContext } from 'vuex';
import { DatasetState, Variable } from './index';
import { mutations } from './module'
import { DistilState } from '../store';
import { HighlightRoot } from '../highlights/index';
import { FilterParams, INCLUDE_FILTER } from '../../util/filters';
import { createPendingSummary, createErrorSummary, createEmptyTableData } from '../../util/data';
import { addHighlightToFilterParams } from '../../util/highlights';

export type DatasetContext = ActionContext<DatasetState, DistilState>;

export const actions = {

	// searches dataset descriptions and column names for supplied terms
	searchDatasets(context: DatasetContext, terms: string): Promise<void> {
		const params = !_.isEmpty(terms) ? `?search=${terms}` : '';
		return axios.get(`/distil/datasets${params}`)
			.then(response => {
				mutations.setDatasets(context, response.data.datasets);
			})
			.catch(error => {
				console.error(error);
				mutations.setDatasets(context, []);
			});
	},

	// fetches all variables for a single dataset.
	fetchVariables(context: DatasetContext, args: { dataset: string }): Promise<void>  {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		return axios.get(`/distil/variables/${args.dataset}`)
			.then(response => {
				mutations.setVariables(context, response.data.variables);
			})
			.catch(error => {
				console.error(error);
				mutations.setVariables(context, []);
			});
	},

	setVariableType(context: DatasetContext, args: { dataset: string, field: string, type: string }): Promise<void>  {
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
		return axios.post(`/distil/variables/${args.dataset}`, {
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

	// fetches variable summary data for the given dataset and variables
	fetchVariableSummaries(context: DatasetContext, args: { dataset: string, variables: Variable[] }): Promise<void[]>  {
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
				return v.key === variable.key;
			});
			if (!exists) {
				// add placeholder
				const key = variable.key;
				const label = variable.label;
				const dataset = args.dataset;
				mutations.updateVariableSummaries(context,  createPendingSummary(key, label, dataset));
				// fetch summary
				promises.push(context.dispatch('fetchVariableSummary', {
					dataset: args.dataset,
					variable: variable.key
				}));
			}
		});
		// fill them in asynchronously
		return Promise.all(promises);
	},

	fetchVariableSummary(context: DatasetContext, args: { dataset: string, variable: string }): Promise<void>  {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.variable) {
			console.warn('`variable` argument is missing');
			return null;
		}
		return axios.post(`/distil/variable-summary/${args.dataset}/${args.variable}`, {})
			.then(response => {
				mutations.updateVariableSummaries(context, response.data.histogram);
			})
			.catch(error => {
				console.error(error);
				const key = args.variable;
				const label = args.variable;
				const dataset = args.dataset;
				mutations.updateVariableSummaries(context,  createErrorSummary(key, label, dataset, error));
			});
	},

	// update filtered data based on the current filter state
	fetchIncludedTableData(context: DatasetContext, args: { dataset: string, filterParams: FilterParams, highlightRoot: HighlightRoot }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.filterParams) {
			console.warn('`filters` argument is missing');
			return null;
		}

		const filterParams = addHighlightToFilterParams(context, args.filterParams, args.highlightRoot, INCLUDE_FILTER);

		// request filtered data from server - no data is valid given filter settings
		return axios.post(`distil/data/${args.dataset}/false`, filterParams)
			.then(response => {
				mutations.setIncludedTableData(context, response.data);
			})
			.catch(error => {
				console.error(error);
				mutations.setIncludedTableData(context, createEmptyTableData());
			});
	},

	// update filtered data based on the  current filter state
	fetchExcludedTableData(context: DatasetContext, args: { dataset: string, filterParams: FilterParams }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.filterParams) {
			console.warn('`filters` argument is missing');
			return null;
		}

		return axios.post(`distil/data/${args.dataset}/true`, args.filterParams)
			.then(response => {
				mutations.setExcludedTableData(context, response.data);
			})
			.catch(error => {
				console.error(error);
				mutations.setExcludedTableData(context, createEmptyTableData());
			});
	},

}
