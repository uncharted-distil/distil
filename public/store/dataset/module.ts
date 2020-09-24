import { Module } from "vuex";
import { state, DatasetState } from "./index";
import { getters as moduleGetters } from "./getters";
import { actions as moduleActions } from "./actions";
import { mutations as moduleMutations } from "./mutations";
import { DistilState } from "../store";
import { getStoreAccessors } from "vuex-typescript";

export const datasetModule: Module<DatasetState, DistilState> = {
  getters: moduleGetters,
  actions: moduleActions,
  mutations: moduleMutations,
  state: state,
};

const { commit, read, dispatch } = getStoreAccessors<DatasetState, DistilState>(
  null
);

// Typed getters
export const getters = {
  // dataset
  getDatasets: read(moduleGetters.getDatasets),
  getFilteredDatasets: read(moduleGetters.getFilteredDatasets),
  getCountOfFilteredDatasets: read(moduleGetters.getCountOfFilteredDatasets),
  getPendingRequests: read(moduleGetters.getPendingRequests),

  // variables
  getVariables: read(moduleGetters.getVariables),
  getGroupings: read(moduleGetters.getGroupings),
  getTimeseriesGroupingVariables: read(
    moduleGetters.getTimeseriesGroupingVariables
  ),
  getVariablesMap: read(moduleGetters.getVariablesMap),
  getVariableTypesMap: read(moduleGetters.getVariableTypesMap),
  getVariableSummariesDictionary: read(
    moduleGetters.getVariableSummariesDictionary
  ),
  getIncludedVariableSummariesDictionary: read(
    moduleGetters.getIncludedVariableSummariesDictionary
  ),
  getExcludedVariableSummariesDictionary: read(
    moduleGetters.getExcludedVariableSummariesDictionary
  ),
  getVariableRankings: read(moduleGetters.getVariableRankings),

  // files
  getFiles: read(moduleGetters.getFiles),
  getTimeseries: read(moduleGetters.getTimeseries),
  getTimeseriesExtrema: read(moduleGetters.getTimeseriesExtrema),

  // join data
  getJoinDatasetsTableData: read(moduleGetters.getJoinDatasetsTableData),

  // included data
  hasIncludedTableData: read(moduleGetters.hasIncludedTableData),
  getIncludedTableData: read(moduleGetters.getIncludedTableData),
  getIncludedTableDataLength: read(moduleGetters.getIncludedTableDataLength),
  getIncludedTableDataNumRows: read(moduleGetters.getIncludedTableDataNumRows),
  getIncludedTableDataItems: read(moduleGetters.getIncludedTableDataItems),
  getIncludedTableDataFields: read(moduleGetters.getIncludedTableDataFields),
  getIncludedSelectedRowData: read(moduleGetters.getIncludedSelectedRowData),

  // excluded data
  hasExcludedTableData: read(moduleGetters.hasExcludedTableData),
  getExcludedTableData: read(moduleGetters.getExcludedTableData),
  getExcludedTableDataLength: read(moduleGetters.getExcludedTableDataLength),
  getExcludedTableDataNumRows: read(moduleGetters.getExcludedTableDataNumRows),
  getExcludedTableDataItems: read(moduleGetters.getExcludedTableDataItems),
  getExcludedTableDataFields: read(moduleGetters.getExcludedTableDataFields),
  getExcludedSelectedRowData: read(moduleGetters.getExcludedSelectedRowData),

  // Remote sensing image band combinatinos
  getMultiBandCombinations: read(moduleGetters.getMultiBandCombinations),
};

// Typed actions
export const actions = {
  // dataset
  fetchDataset: dispatch(moduleActions.fetchDataset),
  searchDatasets: dispatch(moduleActions.searchDatasets),
  geocodeVariable: dispatch(moduleActions.geocodeVariable),
  importDataset: dispatch(moduleActions.importDataset),
  importFullDataset: dispatch(moduleActions.importFullDataset),
  deleteVariable: dispatch(moduleActions.deleteVariable),
  setGrouping: dispatch(moduleActions.setGrouping),
  removeGrouping: dispatch(moduleActions.removeGrouping),
  uploadDataFile: dispatch(moduleActions.uploadDataFile),
  // variables
  fetchVariables: dispatch(moduleActions.fetchVariables),
  fetchVariableSummary: dispatch(moduleActions.fetchVariableSummary),
  fetchJoinDatasetsVariables: dispatch(
    moduleActions.fetchJoinDatasetsVariables
  ),
  setVariableType: dispatch(moduleActions.setVariableType),
  reviewVariableType: dispatch(moduleActions.reviewVariableType),
  fetchIncludedVariableSummaries: dispatch(
    moduleActions.fetchIncludedVariableSummaries
  ),
  fetchExcludedVariableSummaries: dispatch(
    moduleActions.fetchExcludedVariableSummaries
  ),
  fetchGeocodingResults: dispatch(moduleActions.fetchGeocodingResults),

  // ranking
  fetchVariableRankings: dispatch(moduleActions.fetchVariableRankings),
  updateVariableRankings: dispatch(moduleActions.updateVariableRankings),
  // pending request
  updatePendingRequestStatus: dispatch(
    moduleActions.updatePendingRequestStatus
  ),
  removePendingRequest: dispatch(moduleActions.removePendingRequest),
  // join suggestions
  fetchJoinSuggestions: dispatch(moduleActions.fetchJoinSuggestions),
  importJoinDataset: dispatch(moduleActions.importJoinDataset),
  // clusters variables in a dataset for which the operation is meaningful (ie. timeseries)
  fetchClusters: dispatch(moduleActions.fetchClusters),

  // files
  fetchFiles: dispatch(moduleActions.fetchFiles),
  fetchImage: dispatch(moduleActions.fetchImage),
  fetchMultiBandImage: dispatch(moduleActions.fetchMultiBandImage),
  fetchTimeseries: dispatch(moduleActions.fetchTimeseries),
  fetchGraph: dispatch(moduleActions.fetchGraph),
  fetchFile: dispatch(moduleActions.fetchFile),
  // join data
  joinDatasetsPreview: dispatch(moduleActions.joinDatasetsPreview),
  fetchJoinDatasetsTableData: dispatch(
    moduleActions.fetchJoinDatasetsTableData
  ),
  // included / excluded table data
  fetchIncludedTableData: dispatch(moduleActions.fetchIncludedTableData),
  fetchExcludedTableData: dispatch(moduleActions.fetchExcludedTableData),
  // task info
  fetchTask: dispatch(moduleActions.fetchTask),
  // multiband image band combinations
  fetchMultiBandCombinations: dispatch(
    moduleActions.fetchMultiBandCombinations
  ),
  updateRowSelectionData: dispatch(moduleActions.updateRowSelectionData),
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
  updateIncludedVariableSummaries: commit(
    moduleMutations.updateIncludedVariableSummaries
  ),
  updateExcludedVariableSummaries: commit(
    moduleMutations.updateExcludedVariableSummaries
  ),
  clearVariableSummaries: commit(moduleMutations.clearVariableSummaries),
  // ranking
  setVariableRankings: commit(moduleMutations.setVariableRankings),
  updateVariableRankings: commit(moduleMutations.updateVariableRankings),
  // pending update
  updatePendingRequests: commit(moduleMutations.updatePendingRequests),
  removePendingRequest: commit(moduleMutations.removePendingRequest),
  // files
  updateFile: commit(moduleMutations.updateFile),
  updateTimeseries: commit(moduleMutations.updateTimeseries),
  // included / excluded table data
  setJoinDatasetsTableData: commit(moduleMutations.setJoinDatasetsTableData),
  clearJoinDatasetsTableData: commit(
    moduleMutations.clearJoinDatasetsTableData
  ),
  setIncludedTableData: commit(moduleMutations.setIncludedTableData),
  setExcludedTableData: commit(moduleMutations.setExcludedTableData),
  updateBands: commit(moduleMutations.updateBands),
  updateRowSelectionData: commit(moduleMutations.updateRowSelectionData),
};
