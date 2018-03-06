import _ from 'lodash';
import Vue from 'vue';
import { DataState, Variable, Datasets, VariableSummary, Data, Extrema } from './index';
import { updateSummaries } from '../../util/data';
import { Dictionary } from '../../util/dict';

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

	updateResultSummaries(state: DataState, summary: VariableSummary) {
		updateSummaries(summary, state.resultSummaries, 'name');
	},

	updatePredictedSummaries(state: DataState, summary: VariableSummary) {
		updateSummaries(summary, state.predictedSummaries, 'pipelineId');
	},

	updateResidualsSummaries(state: DataState, summary: VariableSummary) {
		updateSummaries(summary, state.residualSummaries, 'pipelineId');
	},

	clearPredictedExtremas(state: DataState) {
		state.predictedExtremas = {};
	},

	updatePredictedExtremas(state: DataState, args: { pipelineId: string, extrema: Extrema }) {
		Vue.set(state.predictedExtremas, args.pipelineId, args.extrema);
	},

	clearResidualsExtremas(state: DataState) {
		state.residualExtremas = {};
	},

	updateResidualsExtremas(state: DataState, args: { pipelineId: string, extrema: Extrema }) {
		Vue.set(state.residualExtremas, args.pipelineId, args.extrema);
	},

	updateResultExtrema(state: DataState, args: { extrema: Extrema }) {
		state.resultExtrema = args.extrema;
	},

	clearResultExtrema(state: DataState) {
		state.resultExtrema = null;
	},

	// sets the current filtered data into the store
	setFilteredData(state: DataState, filteredData: Data) {
		state.filteredData = filteredData;
	},

	// sets the current selected data into the store
	setSelectedData(state: DataState, selectedData: Data) {
		state.selectedData = selectedData;
	},

	// sets the current excluded data into the store
	setExcludedData(state: DataState, excludedData: Data) {
		state.excludedData = excludedData;
	},

	// sets the current result data into the store
	setResultData(state: DataState, resultData: Data) {
		state.resultData = resultData;
	},

	setHighlightedValues(state: DataState, highlightedValues: Dictionary<string[]>) {
		state.highlightedValues = highlightedValues;
	}
}
