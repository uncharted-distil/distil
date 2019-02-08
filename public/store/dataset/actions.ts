import _ from 'lodash';
import axios from 'axios';
import { Dictionary } from '../../util/dict';
import { ActionContext } from 'vuex';
import { Dataset, DatasetState, Variable, VariableSummary } from './index';
import { mutations } from './module';
import { DistilState } from '../store';
import { HighlightRoot } from '../highlights/index';
import { FilterParams, INCLUDE_FILTER } from '../../util/filters';
import { createPendingSummary, createErrorSummary, createEmptyTableData } from '../../util/data';
import { addHighlightToFilterParams } from '../../util/highlights';
import { loadImage } from '../../util/image';
import { getVarType, IMAGE_TYPE, TIMESERIES_TYPE, GEOCODED_LON_PREFIX, GEOCODED_LAT_PREFIX } from '../../util/types';

export type DatasetContext = ActionContext<DatasetState, DistilState>;

export const actions = {

	// fetches a dataset description.
	fetchDataset(context: DatasetContext, dataset: string): Promise<void> {
		return axios.get(`/distil/datasets/${dataset}`)
			.then(response => {
				mutations.setDataset(context, response.data.dataset);
			})
			.catch(error => {
				console.error(error);
				mutations.setDatasets(context, []);
			});
	},

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

	// fetches all variables for a two datasets.
	fetchJoinDatasetsVariables(context: DatasetContext, args: { datasets: string[] }): Promise<void>  {
		if (!args.datasets) {
			console.warn('`datasets` argument is missing');
			return null;
		}
		return Promise.all([
			axios.get(`/distil/variables/${args.datasets[0]}`),
			axios.get(`/distil/variables/${args.datasets[1]}`)
		]).then(res => {
			const varsA = res[0].data.variables;
			const varsB = res[1].data.variables;
			mutations.setVariables(context, varsA.concat(varsB));
		})
		.catch(error => {
			console.error(error);
			mutations.setVariables(context, []);
		});
	},

	geocodeVariable(context: DatasetContext, args: { dataset: string, field: string }): Promise<any>  {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.field) {
			console.warn('`field` argument is missing');
			return null;
		}
		return axios.post(`/distil/geocode/${args.dataset}/${args.field}`, {})
			.then(() => {
				// upon success pull the updated dataset, vars, and summaries
				return Promise.all([
					context.dispatch('fetchDataset', args.dataset),
					context.dispatch('fetchVariables', {
						dataset: args.dataset
					}),
					context.dispatch('fetchVariableSummary', {
						dataset: args.dataset,
						variable: GEOCODED_LON_PREFIX + args.field
					}),
					context.dispatch('fetchVariableSummary', {
						dataset: args.dataset,
						variable: GEOCODED_LAT_PREFIX + args.field
					})
				]);
			})
			.catch(error => {
				console.error(error);
			});
	},

	importDataset(context: DatasetContext, args: { datasetID: string, source: string, provenance: string, terms: string }): Promise<void>  {
		if (!args.datasetID) {
			console.warn('`datasetID` argument is missing');
			return null;
		}
		if (!args.source) {
			console.warn('`terms` argument is missing');
			return null;

		}
		return axios.post(`/distil/import/${args.datasetID}/${args.source}/${args.provenance}`, {})
			.then(response => {
				return context.dispatch('searchDatasets', args.terms);
			});
	},

	joinDatasetsPreview(context: DatasetContext, args: { datasetA: Dataset, datasetB: Dataset, datasetAColumn: string, datasetBColumn: string, joinAccuracy: number }): Promise<void>  {
		if (!args.datasetA) {
			console.warn('`datasetA` argument is missing');
			return null;
		}
		if (!args.datasetB) {
			console.warn('`datasetB` argument is missing');
			return null;
		}
		if (!args.datasetAColumn) {
			console.warn('`datasetAColumn` argument is missing');
			return null;
		}
		if (!args.datasetBColumn) {
			console.warn('`datasetBColumn` argument is missing');
			return null;
		}

		if (_.isNil(args.joinAccuracy)) {
			console.warn('`joinAccuracy` argument is missing');
			return null;
		}

		return axios.post(`/distil/join/${args.datasetA.id}/${args.datasetAColumn}/${args.datasetA.source}/${args.datasetB.id}/${args.datasetBColumn}/${args.datasetB.source}`, {
			accuracy: args.joinAccuracy
		})
			.then(response => {
				return response.data;
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
				return v.dataset === args.dataset && v.key === variable.colName;
			});
			if (!exists) {
				// add placeholder
				const key = variable.colName;
				const label = variable.colDisplayName;
				const dataset = args.dataset;
				mutations.updateVariableSummaries(context, createPendingSummary(key, label, dataset));
				// fetch summary
				promises.push(context.dispatch('fetchVariableSummary', {
					dataset: args.dataset,
					variable: variable.colName
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
				mutations.updateVariableRankings(context, response.data.rankings);
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
			if (type === IMAGE_TYPE) {
				return context.dispatch('fetchImage', {
					dataset: args.dataset,
					url: url
				});
			}
			if (type === TIMESERIES_TYPE) {
				return context.dispatch('fetchTimeseries', {
					dataset: args.dataset,
					source: 'seed',
					url: url
				});
			}
			if (type === 'graph') {
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

	fetchTimeseries(context: DatasetContext, args: { dataset: string, source: string, url: string }) {
		if (!args.url) {
			console.warn('`url` argument is missing');
			return null;
		}
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		return axios.get(`distil/timeseries/${args.dataset}/${args.source}/${args.url}`)
			.then(response => {
				mutations.updateTimeseriesFile(context, { dataset: args.dataset, url: args.url, file: response.data.timeseries });
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
	fetchJoinDatasetsTableData(context: DatasetContext, args: { datasets: string[], filterParams: Dictionary<FilterParams>, highlightRoot: HighlightRoot }) {
		if (!args.datasets) {
			console.warn('`datasets` argument is missing');
			return null;
		}
		if (!args.filterParams) {
			console.warn('`filterParams` argument is missing');
			return null;
		}
		return Promise.all(args.datasets.map(dataset => {

			const highlightRoot = (args.highlightRoot && args.highlightRoot.dataset) === dataset ? args.highlightRoot : null;
			const filterParams = addHighlightToFilterParams(args.filterParams[dataset], highlightRoot, INCLUDE_FILTER);

			return axios.post(`distil/data/${dataset}/false`, filterParams)
				.then(response => {
					mutations.setJoinDatasetsTableData(context, {
						dataset: dataset,
						data: response.data
					});
				})
				.catch(error => {
					console.error(error);
					mutations.setJoinDatasetsTableData(context, {
						dataset: dataset,
						data: createEmptyTableData()
					});
				});
		}));
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

};
