import { Module } from 'vuex';
import { state, DatasetState } from './index';
import { getters as moduleGetters } from './getters';
import { actions as moduleActions } from './actions';
import { mutations as moduleMutations } from './mutations';
import { DistilState } from '../store';
import { getStoreAccessors } from 'vuex-typescript';

export const datasetModule: Module<DatasetState, DistilState> = {
	getters: moduleGetters,
	actions: moduleActions,
	mutations: moduleMutations,
	state: state
}

const { commit, read, dispatch } = getStoreAccessors<DatasetState, DistilState>(null);

// Typed getters
export const getters = {
	// dataset
	getDatasets: read(moduleGetters.getDatasets),
	// variables
	getVariables: read(moduleGetters.getVariables),
	getVariablesMap: read(moduleGetters.getVariablesMap),
	getVariableTypesMap: read(moduleGetters.getVariableTypesMap),
	getVariableSummaries: read(moduleGetters.getVariableSummaries),
	// files
	getFiles: read(moduleGetters.getFiles),
	// included data
	hasIncludedTableData: read(moduleGetters.hasIncludedTableData),
	getIncludedTableData: read(moduleGetters.getIncludedTableData),
	getIncludedTableDataNumRows: read(moduleGetters.getIncludedTableDataNumRows),
	getIncludedTableDataItems: read(moduleGetters.getIncludedTableDataItems),
	getIncludedTableDataFields: read(moduleGetters.getIncludedTableDataFields),
	// excluded data
	hasExcludedTableData: read(moduleGetters.hasExcludedTableData),
	getExcludedTableData: read(moduleGetters.getExcludedTableData),
	getExcludedTableDataNumRows: read(moduleGetters.getExcludedTableDataNumRows),
	getExcludedTableDataItems: read(moduleGetters.getExcludedTableDataItems),
	getExcludedTableDataFields: read(moduleGetters.getExcludedTableDataFields),
}

// Typed actions
export const actions = {
	// dataset
	searchDatasets: dispatch(moduleActions.searchDatasets),
	// variables
	fetchVariables: dispatch(moduleActions.fetchVariables),
	setVariableType: dispatch(moduleActions.setVariableType),
	fetchVariableSummary: dispatch(moduleActions.fetchVariableSummary),
	fetchVariableSummaries: dispatch(moduleActions.fetchVariableSummaries),
	// files
	fetchFiles: dispatch(moduleActions.fetchFiles),
	fetchImage: dispatch(moduleActions.fetchImage),
	fetchTimeseries: dispatch(moduleActions.fetchTimeseries),
	fetchGraph: dispatch(moduleActions.fetchGraph),
	fetchFile: dispatch(moduleActions.fetchFile),
	// included / excluded table data
	fetchIncludedTableData: dispatch(moduleActions.fetchIncludedTableData),
	fetchExcludedTableData: dispatch(moduleActions.fetchExcludedTableData),
}

// Typed mutations
export const mutations = {
	// dataset
	setDatasets: commit(moduleMutations.setDatasets),
	// variables
	setVariables: commit(moduleMutations.setVariables),
	updateVariableType: commit(moduleMutations.updateVariableType),
	updateVariableSummaries: commit(moduleMutations.updateVariableSummaries),
	// files
	updateFile: commit(moduleMutations.updateFile),
	// included / excluded table data
	setIncludedTableData: commit(moduleMutations.setIncludedTableData),
	setExcludedTableData: commit(moduleMutations.setExcludedTableData),

}
