import _ from 'lodash';
import axios from 'axios';
import { Dictionary } from '../../util/dict';
import { ActionContext } from 'vuex';
import { Dataset, DatasetState, Variable, VariableSummary, Grouping, DatasetPendingRequestType, DatasetPendingRequestStatus, DatasetPendingRequest, VariableRankingPendingRequest, GeocodingPendingRequest } from './index';
import { mutations } from './module';
import { DistilState } from '../store';
import { HighlightRoot } from '../highlights/index';
import { FilterParams, INCLUDE_FILTER } from '../../util/filters';
import { createPendingSummary, createErrorSummary, createEmptyTableData, fetchHistogramExemplars } from '../../util/data';
import { addHighlightToFilterParams } from '../../util/highlights';
import { loadImage } from '../../util/image';
import { getVarType, IMAGE_TYPE, TIMESERIES_TYPE, GEOCODED_LON_PREFIX, GEOCODED_LAT_PREFIX } from '../../util/types';

// fetches variables and add dataset name to each variable
function getVariables(dataset: string): Promise<Variable[]> {
	return axios.get(`/distil/variables/${dataset}`).then(response => {
		// extend variable with datasetName and isColTypeReviewed property to track type reviewed state in the client state
		return response.data.variables.map(variable => ({
			...variable,
			datasetName: dataset,
			isColTypeReviewed: false,
		}));
	});
}

export type DatasetContext = ActionContext<DatasetState, DistilState>;

export const actions = {

	// fetches a dataset description.
	fetchDataset(context: DatasetContext, args: { dataset: string }): Promise<void> {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		return axios.get(`/distil/datasets/${args.dataset}`)
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
		return getVariables(args.dataset)
			.then(variables => {
				mutations.setVariables(context, variables);
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
			getVariables(args.datasets[0]),
			getVariables(args.datasets[1])
		]).then(res => {
			const varsA = res[0];
			const varsB = res[1];
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
		const update: GeocodingPendingRequest = {
			id: _.uniqueId(),
			dataset: args.dataset,
			type: DatasetPendingRequestType.GEOCODING,
			field: args.field,
			status: DatasetPendingRequestStatus.PENDING,
		};
		mutations.updatePendingRequests(context, update);
		return axios.post(`/distil/geocode/${args.dataset}/${args.field}`, {})
			.then(() => {
			mutations.updatePendingRequests(context, { ...update, status: DatasetPendingRequestStatus.RESOLVED });
			})
			.catch(error => {
				mutations.updatePendingRequests(context, { ...update, status: DatasetPendingRequestStatus.ERROR });
				console.error(error);
			});
	},

	fetchGeocodingResults(context: DatasetContext, args: { dataset: string, field: string }) {
		// pull the updated dataset, vars, and summaries
		return Promise.all([
			context.dispatch('fetchDataset', {
				dataset: args.dataset
			}),
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
	},

	uploadDataFile(context: DatasetContext, args: { datasetID: string, file: File }) {
		if (!args.datasetID) {
			console.warn('`datasetID` argument is missing');
			return null;
		}
		if (!args.file) {
			console.warn('`file` argument is missing');
			return null;
		}
		const data = new FormData();
		data.append('file', args.file);
		return axios.post(`/distil/upload/${args.datasetID}`, data, {
			headers: { 'Content-Type': 'multipart/form-data' },
		}).then(response => {
			return context.dispatch('importDataset', {
				datasetID: args.datasetID,
				source: 'augmented',
				provenance: 'local',
				terms: args.datasetID
			});
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

	setGrouping(context: DatasetContext, args: { dataset: string, grouping: Grouping }): Promise<any>  {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.grouping) {
			console.warn('`grouping` argument is missing');
			return null;
		}
		return axios.post(`/distil/grouping/${args.dataset}`, {
				grouping: args.grouping
			})
			.then(() => {
				// update dataset
				return Promise.all([
					context.dispatch('fetchDataset', {
						dataset: args.dataset
					}),
					context.dispatch('fetchVariables', {
						dataset: args.dataset
					}),
				]).then(() => {
					mutations.clearVariableSummaries(context);
					const variables = context.getters.getVariables;
					return context.dispatch('fetchVariableSummaries', {
						dataset: args.dataset,
						variables: variables
					});
				});
			})
			.catch(error => {
				console.error(error);
			});
	},

	removeGrouping(context: DatasetContext, args: { dataset: string, grouping: Grouping }): Promise<any>  {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.grouping) {
			console.warn('`grouping` argument is missing');
			return null;
		}
		return axios.post(`/distil/remove-grouping/${args.dataset}`, {
				grouping: args.grouping
			})
			.then(() => {
				// update dataset
				return Promise.all([
					context.dispatch('fetchDataset', {
						dataset: args.dataset
					}),
					context.dispatch('fetchVariables', {
						dataset: args.dataset
					}),
				]).then(() => {
					mutations.clearVariableSummaries(context);
					const variables = context.getters.getVariables;
					return context.dispatch('fetchVariableSummaries', {
						dataset: args.dataset,
						variables: variables
					});
				});
			})
			.catch(error => {
				console.error(error);
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

	reviewVariableType(context: DatasetContext, args: { dataset: string, field: string, isColTypeReviewed: boolean }) {
		mutations.reviewVariableType(context, args);
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

		const timeseries = context.getters.getRouteTimeseriesAnalysis;

		if (timeseries) {
			return axios.post(`distil/timeseries-summary/${args.dataset}/${timeseries}/${args.variable}`, {})
				.then(response => {
					const histogram = response.data.histogram;
					mutations.updateVariableSummaries(context, histogram);
				})
				.catch(error => {
					console.error(error);
				});
		}

		return axios.post(`/distil/variable-summary/${args.dataset}/${args.variable}`, {})
			.then(response => {

				const histogram = response.data.histogram;
				return fetchHistogramExemplars(args.dataset, args.variable, histogram)
					.then(() => {
						mutations.updateVariableSummaries(context, histogram);
					});

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
		const id = _.uniqueId();
		const update: VariableRankingPendingRequest = {
			id,
			dataset: args.dataset,
			type: DatasetPendingRequestType.VARIABLE_RANKING,
			status: DatasetPendingRequestStatus.PENDING,
			rankings: null,
			target: args.target,
		};
		mutations.updatePendingRequests(context, update);
		return axios.get(`/distil/variable-rankings/${args.dataset}/${args.target}`)
			.then(response => {
				console.log(response.data.rankings);
				mutations.updatePendingRequests(context, { ...update, status: DatasetPendingRequestStatus.RESOLVED, rankings: response.data.rankings});
			})
			.catch(error => {
				mutations.updatePendingRequests(context, { ...update, status: DatasetPendingRequestStatus.ERROR });
				console.error(error);
			});
	},

	updateVariableRankings(context: DatasetContext, rankings: Dictionary<number>) {
		mutations.updateVariableRankings(context, rankings);
	},

	updatePendingRequestStatus(context: DatasetContext, args: { id: string, status: DatasetPendingRequestStatus}) {
		const update = context.getters.getPendingRequests.find(item => item.id === args.id);
		if (update) {
			mutations.updatePendingRequests(context, { ...update, status: args.status });
		}
	},

	removePendingRequest(context: DatasetContext, id: string) {
		mutations.removePendingRequest(context, id);
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

	fetchImage(context: DatasetContext, args: { dataset: string, source: string, url: string }) {
		if (!args.url) {
			console.warn('`url` argument is missing');
			return null;
		}
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		return loadImage(`distil/image/${args.dataset}/${args.source}/${args.url}`)
			.then(response => {
				mutations.updateFile(context, { url: args.url, file: response });
			})
			.catch(error => {
				console.error(error);
			});
	},

	fetchTimeseries(context: DatasetContext, args: { dataset: string, xColName: string, yColName: string, timeseriesColName: string, timeseriesID: any }) {

		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.xColName) {
			console.warn('`xColName` argument is missing');
			return null;
		}
		if (!args.yColName) {
			console.warn('`yColName` argument is missing');
			return null;
		}
		if (!args.timeseriesColName) {
			console.warn('`timeseriesColName` argument is missing');
			return null;
		}
		if (!args.timeseriesID) {
			console.warn('`timeseriesID` argument is missing');
			return null;
		}

		return axios.post(`distil/timeseries/${args.dataset}/${args.timeseriesColName}/${args.xColName}/${args.yColName}/${args.timeseriesID}`, {})
			.then(response => {
				mutations.updateTimeseries(context, {
					dataset: args.dataset,
					id: args.timeseriesID,
					timeseries: response.data.timeseries
				});
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
			console.warn('`filterParams` argument is missing');
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
	fetchExcludedTableData(context: DatasetContext, args: { dataset: string, filterParams: FilterParams, highlightRoot: HighlightRoot }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.filterParams) {
			console.warn('`filterParams` argument is missing');
			return null;
		}

		// NOTE: we use an `INCLUDE_FILTER` here because we are inverting all the filters in the REST param
		const filterParams = addHighlightToFilterParams(args.filterParams, args.highlightRoot, INCLUDE_FILTER);

		return axios.post(`distil/data/${args.dataset}/true`, filterParams)
			.then(response => {
				mutations.setExcludedTableData(context, response.data);
			})
			.catch(error => {
				console.error(error);
				mutations.setExcludedTableData(context, createEmptyTableData());
			});
	},

};
