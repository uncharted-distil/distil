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
	getResidualsSummaries: read(moduleGetters.getResidualsSummaries),
	getSelectedFilters: read(moduleGetters.getSelectedFilters),
	getAvailableVariableSummaries: read(moduleGetters.getAvailableVariableSummaries),
	getTrainingVariableSummaries: read(moduleGetters.getTrainingVariableSummaries),
	getTargetVariableSummaries: read(moduleGetters.getTargetVariableSummaries),
	getFilteredData: read(moduleGetters.getFilteredData),
	getFilteredDataNumRows: read(moduleGetters.getFilteredDataNumRows),
	getFilteredDataItems: read(moduleGetters.getFilteredDataItems),
	getFilteredDataFields: read(moduleGetters.getFilteredDataFields),
	getResultData: read(moduleGetters.getResultData),
	getResultDataNumRows: read(moduleGetters.getResultDataNumRows),
	getResultDataItems: read(moduleGetters.getResultDataItems),
	getResultDataFields: read(moduleGetters.getResultDataFields),
	getSelectedData: read(moduleGetters.getSelectedData),
	getSelectedDataNumRows: read(moduleGetters.getSelectedDataNumRows),
	getSelectedDataItems: read(moduleGetters.getSelectedDataItems),
	getSelectedDataFields: read(moduleGetters.getSelectedDataFields)
}

// Typed actions
export const actions = {
	searchDatasets: dispatch(moduleActions.searchDatasets),
	setVariableType: dispatch(moduleActions.setVariableType),
	fetchVariables: dispatch(moduleActions.fetchVariables),
	fetchVariableSummary: dispatch(moduleActions.fetchVariableSummary),
	fetchVariableSummaries: dispatch(moduleActions.fetchVariableSummaries),
	fetchVariablesAndVariableSummaries: dispatch(moduleActions.fetchVariablesAndVariableSummaries),
	updateFilteredData: dispatch(moduleActions.updateFilteredData),
	updateSelectedData: dispatch(moduleActions.updateSelectedData),
	fetchData: dispatch(moduleActions.fetchData),
	fetchResultsSummaries: dispatch(moduleActions.fetchResultsSummaries),
	fetchResidualsSummaries: dispatch(moduleActions.fetchResidualsSummaries),
	updateResults: dispatch(moduleActions.updateResults),
	fetchResults: dispatch(moduleActions.fetchResults)
}


// Typed mutations
export const mutations = {
	updateVariableType: commit(moduleMutations.updateVariableType),
	setVariables: commit(moduleMutations.setVariables),
	setDatasets: commit(moduleMutations.setDatasets),
	updateVariableSummaries: commit(moduleMutations.updateVariableSummaries),
	updateResultsSummaries: commit(moduleMutations.updateResultsSummaries),
	updateResidualsSummaries: commit(moduleMutations.updateResidualsSummaries),
	setFilteredData: commit(moduleMutations.setFilteredData),
	setSelectedData: commit(moduleMutations.setSelectedData),
	setResultData: commit(moduleMutations.setResultData)
}
