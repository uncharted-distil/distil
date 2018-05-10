import _ from 'lodash';
import Vue from 'vue';
import { DataState, Variable, Datasets, VariableSummary, Data, Extrema } from './index';
import { updateSummaries, isPredicted, isAccuracy } from '../../util/data';
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
		updateSummaries(summary, state.predictedSummaries, 'solutionId');
	},

	updateResidualsSummaries(state: DataState, summary: VariableSummary) {
		updateSummaries(summary, state.residualSummaries, 'solutionId');
	},

	updateAccuracySummaries(state: DataState, summary: VariableSummary) {
		updateSummaries(summary, state.accuracySummaries, 'pipelineId');
	},

	clearPredictedExtremas(state: DataState) {
		state.predictedExtremas = {};
	},

	clearPredictedExtrema(state: DataState, solutionId: string) {
		Vue.delete(state.predictedExtremas, solutionId);
	},

	updatePredictedExtremas(state: DataState, args: { solutionId: string, extrema: Extrema }) {
		Vue.set(state.predictedExtremas, args.solutionId, args.extrema);
	},

	clearResidualsExtremas(state: DataState) {
		state.residualExtremas = {};
	},

	clearResidualsExtrema(state: DataState, solutionId: string) {
		Vue.delete(state.residualExtremas, solutionId);
	},

	updateResidualsExtremas(state: DataState, args: { solutionId: string, extrema: Extrema }) {
		Vue.set(state.residualExtremas, args.solutionId, args.extrema);
	},

	updateTargetResultExtrema(state: DataState, args: { extrema: Extrema }) {
		state.resultExtrema = args.extrema;
	},

	clearTargetResultExtrema(state: DataState) {
		state.resultExtrema = null;
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
	setHighlightedResultData(state: DataState, resultData: Data) {
		state.highlightedResultData = resultData;
	},

	// sets the current result data into the store
	setUnhighlightedResultData(state: DataState, resultData: Data) {
		state.unhighlightedResultData = resultData;
	},


	updateHighlightSamples(state: DataState, samples: Dictionary<string[]>) {
		state.highlightValues.samples = samples;
	},

	updateHighlightSummaries(state: DataState, summary: VariableSummary) {
		if (!summary) {
			return;
		}
		const index = _.findIndex(state.highlightValues.summaries, s => s.name === summary.name);
		if (index !== -1) {
			Vue.set(state.highlightValues.summaries, index, summary);
			return;
		}
		state.highlightValues.summaries.push(summary);
	},

	updatePredictedHighlightSummaries(state: DataState, summary: VariableSummary) {
		if (!summary) {
			return;
		}
		const index = _.findIndex(state.highlightValues.summaries, s => s.solutionId === summary.solutionId && isPredicted(s.name));
		if (index !== -1) {
			Vue.set(state.highlightValues.summaries, index, summary);
			return;
		}
		state.highlightValues.summaries.push(summary);
	},

	updateAccuracyHighlightSummaries(state: DataState, summary: VariableSummary) {
		if (!summary) {
			return;
		}
		const index = _.findIndex(state.highlightValues.summaries, s => s.pipelineId === summary.pipelineId && isAccuracy(s.name));
		if (index !== -1) {
			Vue.set(state.highlightValues.summaries, index, summary);
			return;
		}
		state.highlightValues.summaries.push(summary);
	},

	clearHighlightSummaries(state: DataState) {
		state.highlightValues.summaries = [];
	}
}
