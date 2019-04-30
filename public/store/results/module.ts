import { Module } from 'vuex';
import { state, ResultsState } from './index';
import { getters as moduleGetters } from './getters';
import { actions as moduleActions } from './actions';
import { mutations as moduleMutations } from './mutations';
import { DistilState } from '../store';
import { getStoreAccessors } from 'vuex-typescript';

export const resultsModule: Module<ResultsState, DistilState> = {
	getters: moduleGetters,
	actions: moduleActions,
	mutations: moduleMutations,
	state: state
};

const { commit, read, dispatch } = getStoreAccessors<ResultsState, DistilState>(null);

// Typed getters
export const getters = {
	// training / target
	getTrainingSummaries: read(moduleGetters.getTrainingSummaries),
	getTargetSummary: read(moduleGetters.getTargetSummary),
	// result
	hasIncludedResultTableData: read(moduleGetters.hasIncludedResultTableData),
	getIncludedResultTableData: read(moduleGetters.getIncludedResultTableData),
	getIncludedResultTableDataItems: read(moduleGetters.getIncludedResultTableDataItems),
	getIncludedResultTableDataFields: read(moduleGetters.getIncludedResultTableDataFields),
	hasExcludedResultTableData: read(moduleGetters.hasExcludedResultTableData),
	getExcludedResultTableData: read(moduleGetters.getExcludedResultTableData),
	getExcludedResultTableDataItems: read(moduleGetters.getExcludedResultTableDataItems),
	getExcludedResultTableDataFields: read(moduleGetters.getExcludedResultTableDataFields),
	// predicted
	getPredictedSummaries: read(moduleGetters.getPredictedSummaries),
	// residual
	getResidualsSummaries: read(moduleGetters.getResidualsSummaries),
	getResidualsExtrema: read(moduleGetters.getResidualsExtrema),
	// correctness
	getCorrectnessSummaries: read(moduleGetters.getCorrectnessSummaries),
	// forecasting
	getForecastingSummaries: read(moduleGetters.getForecastingSummaries),
	// result table data
	getResultDataNumRows: read(moduleGetters.getResultDataNumRows)
};

// Typed actions
export const actions = {
	// training / target
	fetchTrainingSummaries: dispatch(moduleActions.fetchTrainingSummaries),
	fetchTargetSummary: dispatch(moduleActions.fetchTargetSummary),
	// result
	fetchIncludedResultTableData: dispatch(moduleActions.fetchIncludedResultTableData),
	fetchExcludedResultTableData: dispatch(moduleActions.fetchExcludedResultTableData),
	fetchResultTableData: dispatch(moduleActions.fetchExcludedResultTableData),
	// predicted
	fetchPredictedSummaries: dispatch(moduleActions.fetchPredictedSummaries),
	// residuals
	fetchResidualsSummaries: dispatch(moduleActions.fetchResidualsSummaries),
	fetchResidualsExtrema: dispatch(moduleActions.fetchResidualsExtrema),
	// correctness
	fetchCorrectnessSummaries: dispatch(moduleActions.fetchCorrectnessSummaries),
	// forecasting
	fetchForecastingSummaries: dispatch(moduleActions.fetchForecastingSummaries)
};

// Typed mutations
export const mutations = {
	// training / target
	clearTrainingSummaries: commit(moduleMutations.clearTrainingSummaries),
	clearTargetSummary: commit(moduleMutations.clearTargetSummary),
	updateTrainingSummary: commit(moduleMutations.updateTrainingSummary),
	updateTargetSummary: commit(moduleMutations.updateTargetSummary),
	// result
	setIncludedResultTableData: commit(moduleMutations.setIncludedResultTableData),
	setExcludedResultTableData: commit(moduleMutations.setExcludedResultTableData),
	// predicted
	updatePredictedSummaries: commit(moduleMutations.updatePredictedSummaries),
	// residuals
	updateResidualsSummaries: commit(moduleMutations.updateResidualsSummaries),
	updateResidualsExtrema: commit(moduleMutations.updateResidualsExtrema),
	clearResidualsExtrema: commit(moduleMutations.clearResidualsExtrema),
	// correctness
	updateCorrectnessSummaries: commit(moduleMutations.updateCorrectnessSummaries)
};
