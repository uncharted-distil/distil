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
	// dataset
	getDatasets: read(moduleGetters.getDatasets),
	// variables
	getVariables: read(moduleGetters.getVariables),
	getVariablesMap: read(moduleGetters.getVariablesMap),
	getAvailableVariables: read(moduleGetters.getAvailableVariables),
	getTrainingVariables: read(moduleGetters.getTrainingVariables),
	getAvailableVariablesMap: read(moduleGetters.getAvailableVariablesMap),
	getTrainingVariablesMap: read(moduleGetters.getTrainingVariablesMap),
	// variable summaries
	getVariableSummaries: read(moduleGetters.getVariableSummaries),
	getResultSummaries: read(moduleGetters.getResultSummaries),
	getPredictedSummaries: read(moduleGetters.getPredictedSummaries),
	getResidualsSummaries: read(moduleGetters.getResidualsSummaries),
	getAvailableVariableSummaries: read(moduleGetters.getAvailableVariableSummaries),
	getTrainingVariableSummaries: read(moduleGetters.getTrainingVariableSummaries),
	getTargetVariableSummaries: read(moduleGetters.getTargetVariableSummaries),
	// filters
	getFilters: read(moduleGetters.getFilters),
	getSelectedFilterParams: read(moduleGetters.getSelectedFilterParams),
	// selected data
	hasSelectedData: read(moduleGetters.hasSelectedData),
	getSelectedData: read(moduleGetters.getSelectedData),
	getSelectedDataNumRows: read(moduleGetters.getSelectedDataNumRows),
	getSelectedDataItems: read(moduleGetters.getSelectedDataItems),
	getSelectedDataFields: read(moduleGetters.getSelectedDataFields),
	// excluded data
	hasExcludedData: read(moduleGetters.hasExcludedData),
	getExcludedData: read(moduleGetters.getExcludedData),
	getExcludedDataNumRows: read(moduleGetters.getExcludedDataNumRows),
	getExcludedDataItems: read(moduleGetters.getExcludedDataItems),
	getExcludedDataFields: read(moduleGetters.getExcludedDataFields),
	// result data
	hasResultData: read(moduleGetters.hasResultData),
	getResultData: read(moduleGetters.getResultData),
	getResultDataNumRows: read(moduleGetters.getResultDataNumRows),
	getResultDataItems: read(moduleGetters.getResultDataItems),
	getResultDataFields: read(moduleGetters.getResultDataFields),
	// extrema
	getPredictedExtrema: read(moduleGetters.getPredictedExtrema),
	getResidualExtrema: read(moduleGetters.getResidualExtrema),
	// highlights
	getHighlightedSamples: read(moduleGetters.getHighlightedSamples),
	getHighlightedSummaries: read(moduleGetters.getHighlightedSummaries),
}

// Typed actions
export const actions = {

	searchDatasets: dispatch(moduleActions.searchDatasets),
	setVariableType: dispatch(moduleActions.setVariableType),

	exportProblem: dispatch(moduleActions.exportProblem),

	fetchVariables: dispatch(moduleActions.fetchVariables),

	fetchVariableSummary: dispatch(moduleActions.fetchVariableSummary),
	fetchVariableSummaries: dispatch(moduleActions.fetchVariableSummaries),

	fetchTrainingResultSummaries: dispatch(moduleActions.fetchTrainingResultSummaries),
	fetchResultSummary: dispatch(moduleActions.fetchResultSummary),

	fetchVariablesAndVariableSummaries: dispatch(moduleActions.fetchVariablesAndVariableSummaries),

	fetchSelectedTableData: dispatch(moduleActions.fetchSelectedTableData),
	fetchExcludedTableData: dispatch(moduleActions.fetchExcludedTableData),

	fetchData: dispatch(moduleActions.fetchData),

	fetchPredictedSummaries: dispatch(moduleActions.fetchPredictedSummaries),
	fetchResidualsSummaries: dispatch(moduleActions.fetchResidualsSummaries),
	fetchTargetResultExtrema: dispatch(moduleActions.fetchTargetResultExtrema),
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
	updateTargetResultExtrema: commit(moduleMutations.updateTargetResultExtrema),
	updatePredictedExtremas: commit(moduleMutations.updatePredictedExtremas),
	updateResidualsExtremas: commit(moduleMutations.updateResidualsExtremas),
	clearTargetResultExtrema: commit(moduleMutations.clearTargetResultExtrema),
	clearPredictedExtremas: commit(moduleMutations.clearPredictedExtremas),
	clearPredictedExtrema: commit(moduleMutations.clearPredictedExtrema),
	clearResidualsExtremas: commit(moduleMutations.clearResidualsExtremas),
	clearResidualsExtrema: commit(moduleMutations.clearResidualsExtrema),
	setSelectedData: commit(moduleMutations.setSelectedData),
	setExcludedData: commit(moduleMutations.setExcludedData),
	setResultData: commit(moduleMutations.setResultData),
	updateHighlightSamples: commit(moduleMutations.updateHighlightSamples),
	updateHighlightSummaries: commit(moduleMutations.updateHighlightSummaries),
	updatePredictedHighlightSummaries: commit(moduleMutations.updatePredictedHighlightSummaries),
	clearHighlightSummaries: commit(moduleMutations.clearHighlightSummaries),

}
