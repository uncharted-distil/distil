import _ from 'lodash';
import Vue from 'vue';
import { DataState, Variable, Datasets, VariableSummary, Data } from './index';

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

	updateVariableSummaries(state: DataState, histogram) {
		const index = _.findIndex(state.variableSummaries, elem => {
			return elem.name === histogram.name;
		});
		Vue.set(state.variableSummaries, index, histogram);
	},

	setResultsSummaries(state: DataState, summaries: VariableSummary[]) {
		state.resultsSummaries = summaries;
	},

	updateResultsSummaries(state: DataState, summary: VariableSummary) {
		const index = _.findIndex(state.resultsSummaries, r => r.name === summary.name);
		if (index >= 0) {
			Vue.set(state.resultsSummaries, index, summary);
		} else {
			state.resultsSummaries.push(summary);
		}
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

	highlightFeatureRange(state: DataState, highlight: { name: string, to: string, from: string }) {
		Vue.set(state.highlightedFeatureRanges, highlight.name, {
			from: highlight.from,
			to: highlight.to
		});
		if (state.resultsSummaries) {
			state.resultsSummaries.forEach(summary => {
				Vue.set(state.highlightedFeatureRanges, summary.feature, highlight);
			});
		}
	},

	clearFeatureHighlightRange(state: DataState, name: string) {
		Vue.delete(state.highlightedFeatureRanges, name);
	},

	highlightFeatureValues(state: DataState, highlights: { [name: string]: any }) {
		Vue.set(state, 'highlightedFeatureValues', highlights);
		if (state.resultsSummaries) {
			state.resultsSummaries.forEach(summary => {
				Vue.set(state.highlightedFeatureValues, summary.name, highlights[summary.feature]);
			});
		}
	},

	clearFeatureHighlightValues(state: DataState) {
		Vue.delete(state, 'highlightedFeatureValues');
	}
}
