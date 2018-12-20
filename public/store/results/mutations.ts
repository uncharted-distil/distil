import { ResultsState } from './index';
import { VariableSummary, Extrema, TableData } from '../dataset/index';
import { updateSummaries } from '../../util/data';

export const mutations = {

	// training / target

	clearTrainingSummaries(state: ResultsState) {
		state.trainingSummaries = [];
	},

	clearTargetSummary(state: ResultsState) {
		state.targetSummary = null;
	},

	updateTrainingSummary(state: ResultsState, summary: VariableSummary) {
		updateSummaries(summary, state.trainingSummaries);
	},

	updateTargetSummary(state: ResultsState, summary: VariableSummary) {
		state.targetSummary = summary;
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
		updateSummaries(summary, state.predictedSummaries);
	},

	// residuals

	updateResidualsSummaries(state: ResultsState, summary: VariableSummary) {
		updateSummaries(summary, state.residualSummaries);
	},

	updateResidualsExtrema(state: ResultsState, extrema: Extrema) {
		state.residualsExtrema = extrema;
	},

	clearResidualsExtrema(state: ResultsState, solutionId: string) {
		state.residualsExtrema = {
			min: null,
			max: null
		};
	},

	// correctness

	updateCorrectnessSummaries(state: ResultsState, summary: VariableSummary) {
		updateSummaries(summary, state.correctnessSummaries);
	}
};
