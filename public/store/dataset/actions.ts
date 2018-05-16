import _ from 'lodash';
import axios from 'axios';
import { ActionContext } from 'vuex';
import { DatasetState, Variable, ES_INDEX } from './index';
import { mutations } from './module'
import { DistilState } from '../store';
import { HighlightRoot } from '../highlights/index';
import { FilterParams, INCLUDE_FILTER, EXCLUDE_FILTER } from '../../util/filters';
import { createPendingSummary, createErrorSummary, createEmptyTableData } from '../../util/data';
import { createFilterFromHighlightRoot } from '../../util/highlights';

export type DatasetContext = ActionContext<DatasetState, DistilState>;

export const actions = {

	// searches dataset descriptions and column names for supplied terms
	searchDatasets(context: DatasetContext, terms: string): Promise<void> {
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
	fetchVariables(context: DatasetContext, args: { dataset: string }): Promise<void>  {
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
		return axios.post(`/distil/variables/${ES_INDEX}/${args.dataset}`, {
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

	fetchVariableSummary(context: DatasetContext, args: { dataset: string, variable: string }): Promise<void>  {
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

	// update filtered data based on the current filter state
	fetchIncludedTableData(context: DatasetContext, args: { dataset: string, filters: FilterParams, highlightRoot: HighlightRoot }) {
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

		// request filtered data from server - no data is valid given filter settings
		return axios.post(`distil/data/${ES_INDEX}/${args.dataset}/false`, args.filters)
			.then(response => {
				mutations.setIncludedTableData(context, response.data);
			})
			.catch(error => {
				console.error(error);
				mutations.setIncludedTableData(context, createEmptyTableData(args.dataset));
			});
	},

	// update filtered data based on the  current filter state
	fetchExcludedTableData(context: DatasetContext, args: { dataset: string, filters: FilterParams, highlightRoot: HighlightRoot }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.filters) {
			console.warn('`filters` argument is missing');
			return null;
		}

		const highlightFilter = createFilterFromHighlightRoot(args.highlightRoot, EXCLUDE_FILTER);
		if (highlightFilter) {
			args.filters.filters.push(highlightFilter);
		}

		return axios.post(`distil/data/${ES_INDEX}/${args.dataset}/true`, args.filters)
			.then(response => {
				mutations.setExcludedTableData(context, response.data);
			})
			.catch(error => {
				console.error(error);
				mutations.setExcludedTableData(context, createEmptyTableData(args.dataset));
			});
	},

}
