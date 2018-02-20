import { Module } from 'vuex';
import { state, DataState } from './index';
import { getters as moduleGetters } from './getters';
import { actions as moduleActions } from './actions';
import { mutations as moduleMutations } from './mutations';
import { DistilState } from '../store';
import { getStoreAccessors } from 'vuex-typescript';

export const dataModule: Module<DataState, DistilState> = {
	getters: moduleGetters,
	actions: moduleActions,
	mutations: moduleMutations,
	state: state
}

const { commit, read, dispatch } = getStoreAccessors<DataState, DistilState>(null);

// Typed getters
export const getters = {
	getVariables: read(moduleGetters.getVariables),
	getVariablesMap: read(moduleGetters.getVariablesMap),
	getDatasets: read(moduleGetters.getDatasets),
	getAvailableVariables: read(moduleGetters.getAvailableVariables),
	getAvailableVariablesMap: read(moduleGetters.getAvailableVariablesMap),
	getTrainingVariablesMap: read(moduleGetters.getTrainingVariablesMap),
	getVariableSummaries: read(moduleGetters.getVariableSummaries),
	getResultSummaries: read(moduleGetters.getResultSummaries),
	getPredictedSummaries: read(moduleGetters.getPredictedSummaries),
	getResidualsSummaries: read(moduleGetters.getResidualsSummaries),
	getSelectedFilters: read(moduleGetters.getSelectedFilters),
	getAvailableVariableSummaries: read(moduleGetters.getAvailableVariableSummaries),
	getTrainingVariableSummaries: read(moduleGetters.getTrainingVariableSummaries),
	getTargetVariableSummaries: read(moduleGetters.getTargetVariableSummaries),
	hasFilteredData: read(moduleGetters.hasFilteredData),
	getFilteredData: read(moduleGetters.getFilteredData),
	getFilteredDataNumRows: read(moduleGetters.getFilteredDataNumRows),
	getFilteredDataItems: read(moduleGetters.getFilteredDataItems),
	getFilteredDataFields: read(moduleGetters.getFilteredDataFields),
	hasResultData: read(moduleGetters.hasResultData),
	getResultData: read(moduleGetters.getResultData),
	getResultDataNumRows: read(moduleGetters.getResultDataNumRows),
	getResultDataItems: read(moduleGetters.getResultDataItems),
	getResultDataFields: read(moduleGetters.getResultDataFields),
	hasSelectedData: read(moduleGetters.hasSelectedData),
	getSelectedData: read(moduleGetters.getSelectedData),
	getSelectedDataNumRows: read(moduleGetters.getSelectedDataNumRows),
	getSelectedDataItems: read(moduleGetters.getSelectedDataItems),
	getSelectedDataFields: read(moduleGetters.getSelectedDataFields),
	hasExcludedData: read(moduleGetters.hasExcludedData),
	getExcludedData: read(moduleGetters.getExcludedData),
	getExcludedDataNumRows: read(moduleGetters.getExcludedDataNumRows),
	getExcludedDataItems: read(moduleGetters.getExcludedDataItems),
	getExcludedDataFields: read(moduleGetters.getExcludedDataFields),
	getHighlightedValues: read(moduleGetters.getHighlightedValues),
	getPredictedExtrema: read(moduleGetters.getPredictedExtrema),
	getResidualExtrema: read(moduleGetters.getResidualExtrema),
}

// Typed actions
export const actions = {
	searchDatasets: dispatch(moduleActions.searchDatasets),
	setVariableType: dispatch(moduleActions.setVariableType),
	exportProblem: dispatch(moduleActions.exportProblem),
	fetchVariables: dispatch(moduleActions.fetchVariables),
	fetchVariableSummary: dispatch(moduleActions.fetchVariableSummary),
	fetchVariableSummaries: dispatch(moduleActions.fetchVariableSummaries),
	fetchResultSummaries: dispatch(moduleActions.fetchResultSummaries),
	fetchResultSummary: dispatch(moduleActions.fetchResultSummary),
	fetchVariablesAndVariableSummaries: dispatch(moduleActions.fetchVariablesAndVariableSummaries),
	fetchFilteredTableData: dispatch(moduleActions.fetchFilteredTableData),
	fetchSelectedTableData: dispatch(moduleActions.fetchSelectedTableData),
	fetchExcludedTableData: dispatch(moduleActions.fetchExcludedTableData),
	fetchData: dispatch(moduleActions.fetchData),
	fetchPredictedSummaries: dispatch(moduleActions.fetchPredictedSummaries),
	fetchResidualsSummaries: dispatch(moduleActions.fetchResidualsSummaries),
	fetchPredictedExtrema: dispatch(moduleActions.fetchPredictedExtrema),
	fetchPredictedExtremas: dispatch(moduleActions.fetchPredictedExtremas),
	fetchResidualsExtrema: dispatch(moduleActions.fetchResidualsExtrema),
	fetchResidualsExtremas: dispatch(moduleActions.fetchResidualsExtremas),
	fetchResultTableData: dispatch(moduleActions.fetchResultTableData),
	fetchResults: dispatch(moduleActions.fetchResults),
	fetchDataHighlightValues: dispatch(moduleActions.fetchDataHighlightValues),
	fetchResultHighlightValues: dispatch(moduleActions.fetchResultHighlightValues)
}

// Typed mutations
export const mutations = {
	updateVariableType: commit(moduleMutations.updateVariableType),
	setVariables: commit(moduleMutations.setVariables),
	setDatasets: commit(moduleMutations.setDatasets),
	updateVariableSummaries: commit(moduleMutations.updateVariableSummaries),
	updateResultSummaries: commit(moduleMutations.updateResultSummaries),
	updatePredictedSummaries: commit(moduleMutations.updatePredictedSummaries),
	updateResidualsSummaries: commit(moduleMutations.updateResidualsSummaries),
	updatePredictedExtremas: commit(moduleMutations.updatePredictedExtremas),
	updateResidualsExtremas: commit(moduleMutations.updateResidualsExtremas),
	clearPredictedExtremas: commit(moduleMutations.clearPredictedExtremas),
	clearResidualsExtremas: commit(moduleMutations.clearResidualsExtremas),
	setFilteredData: commit(moduleMutations.setFilteredData),
	setSelectedData: commit(moduleMutations.setSelectedData),
	setExcludedData: commit(moduleMutations.setExcludedData),
	setResultData: commit(moduleMutations.setResultData),
	setHighlightedValues: commit(moduleMutations.setHighlightedValues)
}
