import _ from 'lodash';
import Vue from 'vue';
import { Dictionary } from '../../util/dict';
import { DatasetState, Variable, Dataset, VariableSummary, TimeseriesSummary, TableData, DatasetPendingRequest, VariableRankingPendingRequest, GeocodingPendingRequest } from './index';
import { updateSummaries, isDatamartProvenance } from '../../util/data';

function sortDatasets(a: Dataset, b: Dataset) {

	if (isDatamartProvenance(a.provenance) && !isDatamartProvenance(b.provenance)) {
		return 1;
	}
	if (isDatamartProvenance(b.provenance) && !isDatamartProvenance(a.provenance)) {
		return -1;
	}
	const aID = a.id.toUpperCase();
	const bID = b.id.toUpperCase();
	if (aID < bID) {
		return -1;
	}
	if (aID > bID) {
		return 1;
	}

	return 0;
}

export const mutations = {

	setDataset(state: DatasetState, dataset: Dataset) {
		const index = _.findIndex(state.datasets, d => {
			return d.id === dataset.id;
		});
		if (index === -1) {
			state.datasets.push(dataset);
		} else {
			Vue.set(state.datasets, index, dataset);
		}
		state.datasets.sort(sortDatasets);
	},

	setDatasets(state: DatasetState, datasets: Dataset[]) {
		// individually add datasets if they do not exist
		const lookup = {};
		state.datasets.forEach((d, index) => {
			lookup[d.id] = index;
		});
		datasets.forEach(d => {
			const index = lookup[d.id];
			if (index !== undefined) {
				// update if it already exists
				Vue.set(state.datasets, index, d);
			} else {
				// push if not
				state.datasets.push(d);
			}
		});
		state.datasets.sort(sortDatasets);

		// replace all filtered datasets
		state.filteredDatasets = datasets;
		state.filteredDatasets.sort(sortDatasets);
	},

	setVariables(state: DatasetState, variables: Variable[]) {

		const oldVariables = new Map();

		state.variables.forEach(variable => {
			const {datasetName, colName} = variable;
			oldVariables.set(`${datasetName}:${colName}`, variable);
		});

		const newVariables = variables.map(variable => {
			const {datasetName, colName} = variable;
			const variableKey = `${datasetName}:${colName}`;
			const oldVariable = oldVariables.get(variableKey);

			if (oldVariable) {
				// keep previous column type reviewed state
				variable.isColTypeReviewed = oldVariable.isColTypeReviewed;
				// keep previous variable rankings
				variable.ranking = oldVariable.ranking;
			}
			return variable;
		});
		state.variables = newVariables;
	},

	updateVariableType(state: DatasetState, args: { dataset: string, field: string, type: string }) {

		// TODO: fix this, this is hacky and error prone manually changing the
		// type across the app state.
		// Ideally we have it only in one state, or instead refresh all the
		// relevant store data.

		// update dataset variables
		const dataset = state.datasets.find(d => d.name === args.dataset);
		if (dataset) {
			const variable = dataset.variables.find(v => v.colName === args.field);
			if (variable) {
				variable.colType = args.type;
			}
		}

		// update variables
		const variable = state.variables.find(v => {
			return v.colName === args.field && v.datasetName === args.dataset;
		});
		if (variable) {
			variable.colType = args.type;
		}

		// update table data
		if (state.includedSet.tableData) {
			const col = state.includedSet.tableData.columns.find(c => c.key === args.field);
			if (col) {
				col.type = args.type;
			}
		}
		if (state.excludedSet.tableData) {
			const col = state.excludedSet.tableData.columns.find(c => c.key === args.field);
			if (col) {
				col.type = args.type;
			}
		}

		const joined = state.joinTableData[args.dataset];
		if (joined) {
			const col = joined.columns.find(c => c.key === args.field);
			if (col) {
				col.type = args.type;
			}
		}
	},

	reviewVariableType(state: DatasetState, update) {
		const index = _.findIndex(state.variables, v => {
			return v.colName === update.field;
		});
		state.variables[index].isColTypeReviewed = update.isColTypeReviewed;
	},

	updateIncludedVariableSummaries(state: DatasetState, summary: VariableSummary) {
		updateSummaries(summary, state.includedSet.variableSummaries);
	},

	updateExcludedVariableSummaries(state: DatasetState, summary: VariableSummary) {
		updateSummaries(summary, state.excludedSet.variableSummaries);
	},

	clearVariableSummaries(state: DatasetState) {
		state.includedSet.variableSummaries = [];
		state.excludedSet.variableSummaries = [];
	},

	setVariableRankings(state: DatasetState, args: { dataset: string, rankings: Dictionary<number>  }) {
		Vue.set(state.variableRankings, args.dataset, args.rankings);
	},

	updateVariableRankings(state: DatasetState, rankings: Dictionary<number>) {
		// add rank property if ranking data returned, otherwise don't include it
		if (!_.isEmpty(rankings)) {
			state.variables.forEach(v => {
				let rank = 0;
				if (rankings[v.colName]) {
					rank = rankings[v.colName];
				}
				Vue.set(v, 'ranking', rank);
			});
		} else {
			state.variables.forEach(v => Vue.delete(v, 'ranking'));
		}
	},

	updatePendingRequests(state: DatasetState, pendingRequest: DatasetPendingRequest) {
		const sameIdIndex = state.pendingRequests.findIndex(item => pendingRequest.id === item.id);
		const sameTypeIndex = state.pendingRequests.findIndex(item => pendingRequest.type === item.type);
		if (sameIdIndex >= 0) {
			Vue.set(state.pendingRequests, sameIdIndex, pendingRequest);
		// only keep latest single request object for each type in the pendingRequests list
		} else if (sameTypeIndex >= 0) {
			Vue.set(state.pendingRequests, sameTypeIndex, pendingRequest);
		} else {
			state.pendingRequests.push(pendingRequest);
		}
	},

	removePendingRequest(state: DatasetState, id: string) {
		state.pendingRequests = state.pendingRequests.filter(item => item.id !== id);
	},

	updateFile(state: DatasetState, args: { url: string, file: any }) {
		Vue.set(state.files, args.url, args.file);
	},

	updateTimeseries(state: DatasetState, args: { dataset: string, id: string, timeseries: number[][] }) {

		if (!state.timeseries[args.dataset]) {
			Vue.set(state.timeseries, args.dataset, {});
		}
		Vue.set(state.timeseries[args.dataset], args.id, args.timeseries);

		const minX = _.minBy(args.timeseries, d => d[0])[0];
		const maxX = _.maxBy(args.timeseries, d => d[0])[0];
		const minY = _.minBy(args.timeseries, d => d[1])[1];
		const maxY = _.maxBy(args.timeseries, d => d[1])[1];

		if (!state.timeseriesExtrema[args.dataset]) {
			Vue.set(state.timeseriesExtrema, args.dataset, {
				x: {
					min: minX,
					max: maxX
				},
				y: {
					min: minY,
					max: maxY
				}
			});
			return;
		}
		const x = state.timeseriesExtrema[args.dataset].x;
		const y = state.timeseriesExtrema[args.dataset].y;
		Vue.set(x, 'min', Math.min(x.min, minX));
		Vue.set(x, 'max', Math.max(x.max, maxX));
		Vue.set(y, 'min', Math.min(y.min, minY));
		Vue.set(y, 'max', Math.max(y.max, maxY));
	},

	setJoinDatasetsTableData(state: DatasetState, args: { dataset: string, data: TableData }) {
		Vue.set(state.joinTableData, args.dataset, args.data);
	},

	clearJoinDatasetsTableData(state: DatasetState) {
		state.joinTableData = {};
	},

	// sets the current selected data into the store
	setIncludedTableData(state: DatasetState, tableData: TableData) {
		state.includedSet.tableData = tableData;
	},

	// sets the current excluded data into the store
	setExcludedTableData(state: DatasetState, tableData: TableData) {
		state.excludedSet.tableData = tableData;
	}

};
