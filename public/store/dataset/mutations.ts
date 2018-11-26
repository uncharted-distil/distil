import _ from 'lodash';
import Vue from 'vue';
import { Dictionary } from '../../util/dict';
import { DatasetState, Variable, Dataset, VariableSummary, TableData } from './index';
import { updateSummaries } from '../../util/data';

export const mutations = {

	setDatasets(state: DatasetState, datasets: Dataset[]) {
		state.datasets = datasets;
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

	// sets the current selected data into the store
	setIncludedTableData(state: DatasetState, includedTableData: TableData) {
		state.includedTableData = includedTableData;
	},

	// sets the current excluded data into the store
	setExcludedTableData(state: DatasetState, excludedTableData: TableData) {
		state.excludedTableData = excludedTableData;
	}

}
