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
			return v.key === update.field;
		});
		state.variables[index].type = update.type;
	},

	updateVariableSummaries(state: DatasetState, summary: VariableSummary) {
		updateSummaries(summary, state.variableSummaries);
	},

	updateVariableRankings(state: DatasetState, rankings: Dictionary<number>) {
		state.variables.forEach(v => {
			if (rankings[v.key]) {
				// add ranking
				v.ranking = rankings[v.key];
			}
		});
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
