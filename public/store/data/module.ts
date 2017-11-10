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
	getResultsSummaries: read(moduleGetters.getResultsSummaries),
	getSelectedFilters: read(moduleGetters.getSelectedFilters),
	getAvailableVariableSummaries: read(moduleGetters.getAvailableVariableSummaries),
	getTrainingVariableSummaries: read(moduleGetters.getTrainingVariableSummaries),
	getTargetVariableSummaries: read(moduleGetters.getTargetVariableSummaries),
	getFilteredData: read(moduleGetters.getFilteredData),
	getFilteredDataItems: read(moduleGetters.getFilteredDataItems),
	getFilteredDataFields: read(moduleGetters.getFilteredDataFields),
	getResultData: read(moduleGetters.getResultData),
	getResultDataItems: read(moduleGetters.getResultDataItems),
	getResultDataFields: read(moduleGetters.getResultDataFields),
	getSelectedData: read(moduleGetters.getSelectedData),
	getSelectedDataItems: read(moduleGetters.getSelectedDataItems),
	getSelectedDataFields: read(moduleGetters.getSelectedDataFields),
	getHighlightedFeatureValues: read(moduleGetters.getHighlightedFeatureValues),
	getHighlightedFeatureRanges: read(moduleGetters.getHighlightedFeatureRanges)
}

// Typed actions
export const actions = {
	searchDatasets: dispatch(moduleActions.searchDatasets),
	getVariables: dispatch(moduleActions.getVariables),
	setVariableType: dispatch(moduleActions.setVariableType),
	getVariableSummaries: dispatch(moduleActions.getVariableSummaries),
	getVariableSummary: dispatch(moduleActions.getVariableSummary),
	updateFilteredData: dispatch(moduleActions.updateFilteredData),
	updateSelectedData: dispatch(moduleActions.updateSelectedData),
	getResultsSummaries: dispatch(moduleActions.getResultsSummaries),
	updateResults: dispatch(moduleActions.updateResults),
	highlightFeatureRange: dispatch(moduleActions.highlightFeatureRange),
	clearFeatureHighlightRange: dispatch(moduleActions.clearFeatureHighlightRange),
	highlightFeatureValues: dispatch(moduleActions.highlightFeatureValues),
	clearFeatureHighlightValues: dispatch(moduleActions.clearFeatureHighlightValues)
}


// Typed mutations
export const mutations = {
	updateVariableType: commit(moduleMutations.updateVariableType),
	setVariables: commit(moduleMutations.setVariables),
	setDatasets: commit(moduleMutations.setDatasets),
	setVariableSummaries: commit(moduleMutations.setVariableSummaries),
	updateVariableSummaries: commit(moduleMutations.updateVariableSummaries),
	setResultsSummaries: commit(moduleMutations.setResultsSummaries),
	updateResultsSummaries: commit(moduleMutations.updateResultsSummaries),
	setFilteredData: commit(moduleMutations.setFilteredData),
	setSelectedData: commit(moduleMutations.setSelectedData),
	setResultData: commit(moduleMutations.setResultData),
	highlightFeatureRange: commit(moduleMutations.highlightFeatureRange),
	clearFeatureHighlightRange: commit(moduleMutations.clearFeatureHighlightRange),
	highlightFeatureValues: commit(moduleMutations.highlightFeatureValues),
	clearFeatureHighlightValues: commit(moduleMutations.clearFeatureHighlightValues)
}
