import _ from 'lodash';
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
		updateSummaries(summary, state.variableSummaries, 'name');
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
