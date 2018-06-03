import Vue from 'vue';
import { ResultsState } from './index';
import { VariableSummary, Extrema, TableData } from '../dataset/index';
import { updateSummaries } from '../../util/data';

const SOLUTION_ID = 'solutionId'

export const mutations = {

	// results

	updateResultSummaries(state: ResultsState, summary: VariableSummary) {
		updateSummaries(summary, state.resultSummaries, 'name');
	},

	updateResultExtrema(state: ResultsState, args: { extrema: Extrema }) {
		state.targetResultExtrema = args.extrema;
	},

	clearResultExtrema(state: ResultsState) {
		state.targetResultExtrema = null;
	},

	// sets the current result data into the store
	setIncludedResultTableData(state: ResultsState, resultData: TableData) {
		state.includedResultTableData = resultData;
	},

	// sets the current result data into the store
	setExcludedResultTableData(state: ResultsState, resultData: TableData) {
		state.excludedResultTableData = resultData;
	},

	// predicted

	updatePredictedSummaries(state: ResultsState, summary: VariableSummary) {
		updateSummaries(summary, state.predictedSummaries, SOLUTION_ID);
	},

	updatePredictedExtremas(state: ResultsState, args: { solutionId: string, extrema: Extrema }) {
		Vue.set(state.predictedExtremas, args.solutionId, args.extrema);
	},

	clearPredictedExtrema(state: ResultsState, solutionId: string) {
		Vue.delete(state.predictedExtremas, solutionId);
	},

	clearPredictedExtremas(state: ResultsState) {
		state.predictedExtremas = {};
	},

	// residuals

	updateResidualsSummaries(state: ResultsState, summary: VariableSummary) {
		updateSummaries(summary, state.residualSummaries, SOLUTION_ID);
	},

	updateResidualsExtremas(state: ResultsState, args: { solutionId: string, extrema: Extrema }) {
		Vue.set(state.residualExtremas, args.solutionId, args.extrema);
	},

	clearResidualsExtrema(state: ResultsState, solutionId: string) {
		Vue.delete(state.residualExtremas, solutionId);
	},

	clearResidualsExtremas(state: ResultsState) {
		state.residualExtremas = {};
	},

	// correctness

	updateCorrectnessSummaries(state: ResultsState, summary: VariableSummary) {
		updateSummaries(summary, state.correctnessSummaries, SOLUTION_ID);
	}
}
