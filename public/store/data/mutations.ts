import _ from 'lodash';
import { DataState, Variable, Datasets, VariableSummary, Data } from './index';
import { updateSummaries } from '../../util/data';

export const mutations = {

	updateVariableType(state: DataState, update) {
		const index = _.findIndex(state.variables, elem => {
			return elem.name === update.field;
		});
		state.variables[index].type = update.type;
	},

	setVariables(state: DataState, variables: Variable[]) {
		state.variables = variables;
	},

	setDatasets(state: DataState, datasets: Datasets[]) {
		state.datasets = datasets;
	},

	updateVariableSummaries(state: DataState, summary: VariableSummary) {
		updateSummaries(summary, state.variableSummaries, 'name');
	},

	updateResultsSummaries(state: DataState, summary: VariableSummary) {
		updateSummaries(summary, state.resultsSummaries, 'pipelineId');
	},

	updateResidualsSummaries(state: DataState, summary: VariableSummary) {
		updateSummaries(summary, state.residualSummaries, 'pipelineId');
	},

	// sets the current filtered data into the store
	setFilteredData(state: DataState, filteredData: Data) {
		state.filteredData = filteredData;
	},

	// sets the current selected data into the store
	setSelectedData(state: DataState, selectedData: Data) {
		state.selectedData = selectedData;
	},

	// sets the current result data into the store
	setResultData(state: DataState, resultData: Data) {
		state.resultData = resultData;
	}
}
