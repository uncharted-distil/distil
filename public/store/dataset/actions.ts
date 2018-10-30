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
import { loadImage } from '../../util/image';
import { getVarType } from '../../util/types';

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

				const histogram = response.data.histogram;
				if (histogram.files) {
					// if there a linked files, fetch those before resolving
					return context.dispatch('fetchFiles', {
						dataset: args.dataset,
						variable: args.variable,
						urls: histogram.files
					}).then(() => {
						mutations.updateVariableSummaries(context, histogram);
					});
				} else {
					mutations.updateVariableSummaries(context, histogram);
				}

			})
			.catch(error => {
				console.error(error);
				const key = args.variable;
				const label = args.variable;
				const dataset = args.dataset;
				mutations.updateVariableSummaries(context,  createErrorSummary(key, label, dataset, error));
			});
	},

	fetchVariableRankings(context: DatasetContext, args: { dataset: string, target: string }) {
		return axios.get(`/distil/variable-rankings/${args.dataset}/${args.target}`)
			.then(response => {
				mutations.updateVariableRankings(context, response.data);
			})
			.catch(error => {
				console.error(error);
			});
	},

	// update filtered data based on the current filter state
	fetchFiles(context: DatasetContext, args: { dataset: string, variable: string, urls: string[] }) {
		if (!args.urls) {
			console.warn('`url` argument is missing');
			return null;
		}
		const type = getVarType(args.variable);
		return Promise.all(args.urls.map(url => {
			if (type == 'image') {
				return context.dispatch('fetchImage', {
					dataset: args.dataset,
					url: url
				});
			}
			if (type == 'timeseries') {
				return context.dispatch('fetchTimeseries', {
					dataset: args.dataset,
					url: url
				});
			}
			if (type == 'graph') {
				return context.dispatch('fetchGraph', {
					dataset: args.dataset,
					url: url
				});
			}
			return context.dispatch('fetchFile', {
				dataset: args.dataset,
				url: url
			});
		}));
	},

	fetchImage(context: DatasetContext, args: { dataset: string, url: string }) {
		if (!args.url) {
			console.warn('`url` argument is missing');
			return null;
		}
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		return loadImage(`distil/image/${args.dataset}/${args.url}`)
			.then(response => {
				mutations.updateFile(context, { url: args.url, file: response });
			})
			.catch(error => {
				console.error(error);
			});
	},

	fetchTimeseries(context: DatasetContext, args: { dataset: string, url: string }) {
		if (!args.url) {
			console.warn('`url` argument is missing');
			return null;
		}
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		return axios.get(`distil/timeseries/${args.dataset}/${args.url}`)
			.then(response => {
				mutations.updateFile(context, { url: args.url, file: response.data.timeseries });
			})
			.catch(error => {
				console.error(error);
			});
	},

	fetchGraph(context: DatasetContext, args: { dataset: string, url: string }) {
		if (!args.url) {
			console.warn('`url` argument is missing');
			return null;
		}
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		return axios.get(`distil/graphs/${args.dataset}/${args.url}`)
			.then(response => {
				if (response.data.graphs.length > 0) {
					const graph = response.data.graphs[0];
					const parsed = {
						nodes: graph.nodes.map(n => {
							return {
								id: n.id,
								label: n.label,
								x: n.attributes.attr1,
								y: n.attributes.attr2,
								size: 1,
								color: '#ec5148'
							};
						}),
						edges: graph.edges.map((e, i) => {
							return {
								id: `e${i}`,
								source: e.source,
								target: e.target,
								color: '#aaa'
							};
						})
					};
					mutations.updateFile(context, { url: args.url, file: parsed });
				}
			})
			.catch(error => {
				console.error(error);
			});
	},

	fetchFile(context: DatasetContext, args: { dataset: string, url: string }) {
		if (!args.url) {
			console.warn('`url` argument is missing');
			return null;
		}
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		return axios.get(`distil/resource/${args.dataset}/${args.url}`)
			.then(response => {
				mutations.updateFile(context, { url: args.url, file: response.data });
			})
			.catch(error => {
				console.error(error);
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

		const filterParams = addHighlightToFilterParams(args.filterParams, args.highlightRoot, INCLUDE_FILTER);

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
