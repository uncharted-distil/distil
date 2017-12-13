import _ from 'lodash';
import Vue from 'vue';
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

	setVariableSummaries(state: DataState, summaries: VariableSummary[]) {
		state.variableSummaries = summaries;
	},

	updateVariableSummaries(state: DataState, summary: VariableSummary) {
		updateSummaries(summary, state.variableSummaries, 'name');
	},

	setResultsSummaries(state: DataState, summaries: VariableSummary[]) {
		state.resultsSummaries = summaries;
	},

	updateResultsSummaries(state: DataState, summary: VariableSummary) {
		updateSummaries(summary, state.resultsSummaries, 'pipelineId');
	},

	setResidualsSummaries(state: DataState, summaries: VariableSummary[]) {
		state.residualSummaries = summaries;
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
	},

	highlightFeatureRange(state: DataState, highlight: { name: string, to: number, from: number }) {
		Vue.set(state.highlightedFeatureRanges, highlight.name, {
			from: highlight.from,
			to: highlight.to
		});
	},

	clearFeatureHighlightRange(state: DataState, name: string) {
		Vue.delete(state.highlightedFeatureRanges, name);
	},

	highlightFeatureValues(state: DataState, highlights: { [name: string]: any }) {
		state.highlightedFeatureValues = highlights;
	},

	clearFeatureHighlightValues(state: DataState) {
		state.highlightedFeatureValues = {};
	}
}
