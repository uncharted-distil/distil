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
}

const { commit, read, dispatch } = getStoreAccessors<ResultsState, DistilState>(null);

// Typed getters
export const getters = {
	// result
	getResultSummaries: read(moduleGetters.getResultSummaries),
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
	getPredictedExtrema: read(moduleGetters.getPredictedExtrema),
	// residual
	getResidualsSummaries: read(moduleGetters.getResidualsSummaries),
	getResidualExtrema: read(moduleGetters.getResidualExtrema),
	// correctness
	getCorrectnessSummaries: read(moduleGetters.getCorrectnessSummaries),
	// result table data
	getResultDataNumRows: read(moduleGetters.getResultDataNumRows)
}

// Typed actions
export const actions = {
	// result
	fetchResultSummary: dispatch(moduleActions.fetchResultSummary),
	fetchResultExtrema: dispatch(moduleActions.fetchResultExtrema),
	fetchTrainingResultSummaries: dispatch(moduleActions.fetchTrainingResultSummaries),

	fetchIncludedResultTableData: dispatch(moduleActions.fetchIncludedResultTableData),
	fetchExcludedResultTableData: dispatch(moduleActions.fetchExcludedResultTableData),
	// predicted
	fetchPredictedSummaries: dispatch(moduleActions.fetchPredictedSummaries),
	fetchPredictedExtrema: dispatch(moduleActions.fetchPredictedExtrema),
	fetchPredictedExtremas: dispatch(moduleActions.fetchPredictedExtremas),
	// residuals
	fetchResidualsSummaries: dispatch(moduleActions.fetchResidualsSummaries),
	fetchResidualsExtrema: dispatch(moduleActions.fetchResidualsExtrema),
	fetchResidualsExtremas: dispatch(moduleActions.fetchResidualsExtremas),
	// correctness
	fetchCorrectnessSummaries: dispatch(moduleActions.fetchCorrectnessSummaries)
}

// Typed mutations
export const mutations = {
	// result
	updateResultSummaries: commit(moduleMutations.updateResultSummaries),
	updateResultExtrema: commit(moduleMutations.updateResultExtrema),
	clearResultExtrema: commit(moduleMutations.clearResultExtrema),
	setIncludedResultTableData: commit(moduleMutations.setIncludedResultTableData),
	setExcludedResultTableData: commit(moduleMutations.setExcludedResultTableData),
	// predicted
	updatePredictedSummaries: commit(moduleMutations.updatePredictedSummaries),
	updatePredictedExtremas: commit(moduleMutations.updatePredictedExtremas),
	clearPredictedExtrema: commit(moduleMutations.clearPredictedExtrema),
	clearPredictedExtremas: commit(moduleMutations.clearPredictedExtremas),
	// residuals
	updateResidualsSummaries: commit(moduleMutations.updateResidualsSummaries),
	updateResidualsExtremas: commit(moduleMutations.updateResidualsExtremas),
	clearResidualsExtrema: commit(moduleMutations.clearResidualsExtrema),
	clearResidualsExtremas: commit(moduleMutations.clearResidualsExtremas),
	// correctness
	updateCorrectnessSummaries: commit(moduleMutations.updateCorrectnessSummaries)
}
