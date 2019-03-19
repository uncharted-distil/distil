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
};

const { commit, read, dispatch } = getStoreAccessors<DatasetState, DistilState>(null);

// Typed getters
export const getters = {
	// dataset
	getDatasets: read(moduleGetters.getDatasets),
	getFilteredDatasets: read(moduleGetters.getFilteredDatasets),
	// variables
	getVariables: read(moduleGetters.getVariables),
	getVariablesMap: read(moduleGetters.getVariablesMap),
	getVariableTypesMap: read(moduleGetters.getVariableTypesMap),
	getVariableSummaries: read(moduleGetters.getVariableSummaries),
	// files
	getFiles: read(moduleGetters.getFiles),
	getTimeseriesExtrema: read(moduleGetters.getTimeseriesExtrema),
	// join data
	getJoinDatasetsTableData: read(moduleGetters.getJoinDatasetsTableData),
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
};

// Typed actions
export const actions = {
	// dataset
	fetchDataset: dispatch(moduleActions.fetchDataset),
	searchDatasets: dispatch(moduleActions.searchDatasets),
	geocodeVariable: dispatch(moduleActions.geocodeVariable),
	importDataset: dispatch(moduleActions.importDataset),
	uploadDataFile: dispatch(moduleActions.uploadDataFile),
	// variables
	fetchVariables: dispatch(moduleActions.fetchVariables),
	fetchJoinDatasetsVariables: dispatch(moduleActions.fetchJoinDatasetsVariables),
	setVariableType: dispatch(moduleActions.setVariableType),
	reviewVariableType: dispatch(moduleActions.reviewVariableType),
	fetchVariableSummary: dispatch(moduleActions.fetchVariableSummary),
	fetchVariableSummaries: dispatch(moduleActions.fetchVariableSummaries),
	// ranking
	fetchVariableRankings: dispatch(moduleActions.fetchVariableRankings),
	// files
	fetchFiles: dispatch(moduleActions.fetchFiles),
	fetchImage: dispatch(moduleActions.fetchImage),
	fetchTimeseries: dispatch(moduleActions.fetchTimeseries),
	fetchGraph: dispatch(moduleActions.fetchGraph),
	fetchFile: dispatch(moduleActions.fetchFile),
	// join data
	joinDatasetsPreview : dispatch(moduleActions.joinDatasetsPreview),
	fetchJoinDatasetsTableData: dispatch(moduleActions.fetchJoinDatasetsTableData),
	// included / excluded table data
	fetchIncludedTableData: dispatch(moduleActions.fetchIncludedTableData),
	fetchExcludedTableData: dispatch(moduleActions.fetchExcludedTableData),
};

// Typed mutations
export const mutations = {
	// dataset
	setDataset: commit(moduleMutations.setDataset),
	setDatasets: commit(moduleMutations.setDatasets),
	// variables
	setVariables: commit(moduleMutations.setVariables),
	updateVariableType: commit(moduleMutations.updateVariableType),
	reviewVariableType: commit(moduleMutations.reviewVariableType),
	updateVariableSummaries: commit(moduleMutations.updateVariableSummaries),
	// ranking
	updateVariableRankings: commit(moduleMutations.updateVariableRankings),
	// files
	updateFile: commit(moduleMutations.updateFile),
	updateTimeseriesFile: commit(moduleMutations.updateTimeseriesFile),
	// included / excluded table data
	setJoinDatasetsTableData: commit(moduleMutations.setJoinDatasetsTableData),
	clearJoinDatasetsTableData: commit(moduleMutations.clearJoinDatasetsTableData),
	setIncludedTableData: commit(moduleMutations.setIncludedTableData),
	setExcludedTableData: commit(moduleMutations.setExcludedTableData),

};
