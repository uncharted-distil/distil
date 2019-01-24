import _ from 'lodash';
import Vue from 'vue';
import { Dictionary } from '../../util/dict';
import { DatasetState, Variable, Dataset, VariableSummary, TableData, DATAMART_PROVENANCE } from './index';
import { updateSummaries } from '../../util/data';

function sortDatasets(a: Dataset, b: Dataset) {

	if (a.provenance === DATAMART_PROVENANCE && b.provenance !== DATAMART_PROVENANCE) {
		return 1;
	}
	if (b.provenance === DATAMART_PROVENANCE && a.provenance !== DATAMART_PROVENANCE) {
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
		state.variables = variables;
	},

	updateVariableType(state: DatasetState, update) {
		const index = _.findIndex(state.variables, v => {
			return v.colName === update.field;
		});
		state.variables[index].colType = update.type;
	},

	updateVariableSummaries(state: DatasetState, summary: VariableSummary) {
		updateSummaries(summary, state.variableSummaries);
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

	updateFile(state: DatasetState, args: { url: string, file: any }) {
		Vue.set(state.files, args.url, args.file);
	},

	updateTimeseriesFile(state: DatasetState, args: { dataset: string, url: string, file: number[][] }) {

		Vue.set(state.files, args.url, args.file);

		const minX = _.minBy(args.file, d => d[0])[0];
		const maxX = _.maxBy(args.file, d => d[0])[0];
		const minY = _.minBy(args.file, d => d[1])[1];
		const maxY = _.maxBy(args.file, d => d[1])[1];

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
	setIncludedTableData(state: DatasetState, includedTableData: TableData) {
		state.includedTableData = includedTableData;
	},

	// sets the current excluded data into the store
	setExcludedTableData(state: DatasetState, excludedTableData: TableData) {
		state.excludedTableData = excludedTableData;
	}

};
