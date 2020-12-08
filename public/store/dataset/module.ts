import { Module } from "vuex";
import { getStoreAccessors } from "vuex-typescript";
import { DistilState } from "../store";
import { actions as moduleActions } from "./actions";
import { getters as moduleGetters } from "./getters";
import { DatasetState, state } from "./index";
import { mutations as moduleMutations } from "./mutations";

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
  // highlighted data
  getHighlightedIncludeTableDataItems: read(
    moduleGetters.getHighlightedIncludeTableDataItems
  ),
  getHighlightedExcludeTableDataItems: read(
    moduleGetters.getHighlightedExcludeTableDataItems
  ),
  getNumberOfRecords: read(moduleGetters.getNumberOfRecords),
  // included data
  hasIncludedTableData: read(moduleGetters.hasIncludedTableData),
  getIncludedTableData: read(moduleGetters.getIncludedTableData),
  getIncludedTableDataLength: read(moduleGetters.getIncludedTableDataLength),
  getIncludedTableDataNumRows: read(moduleGetters.getIncludedTableDataNumRows),
  getIncludedTableDataItems: read(moduleGetters.getIncludedTableDataItems),
  getIncludedTableDataFields: read(moduleGetters.getIncludedTableDataFields),
  getIncludedSelectedRowData: read(moduleGetters.getIncludedSelectedRowData),
  getAreaOfInterestIncludeInnerItems: read(
    moduleGetters.getAreaOfInterestIncludeInnerItems
  ),
  getAreaOfInterestIncludeOuterItems: read(
    moduleGetters.getAreaOfInterestIncludeOuterItems
  ),

  // excluded data
  hasExcludedTableData: read(moduleGetters.hasExcludedTableData),
  getExcludedTableData: read(moduleGetters.getExcludedTableData),
  getExcludedTableDataLength: read(moduleGetters.getExcludedTableDataLength),
  getExcludedTableDataNumRows: read(moduleGetters.getExcludedTableDataNumRows),
  getExcludedTableDataItems: read(moduleGetters.getExcludedTableDataItems),
  getExcludedTableDataFields: read(moduleGetters.getExcludedTableDataFields),
  getExcludedSelectedRowData: read(moduleGetters.getExcludedSelectedRowData),
  getAreaOfInterestExcludeInnerItems: read(
    moduleGetters.getAreaOfInterestExcludeInnerItems
  ),
  getAreaOfInterestExcludeOuterItems: read(
    moduleGetters.getAreaOfInterestExcludeOuterItems
  ),
  // Remote sensing image band combinatinos
  getMultiBandCombinations: read(moduleGetters.getMultiBandCombinations),

  // Modeling metric methologies
  getModelingMetrics: read(moduleGetters.getModelingMetrics),
};

// Typed actions
export const actions = {
  // dataset
  fetchDataset: dispatch(moduleActions.fetchDataset),
  searchDatasets: dispatch(moduleActions.searchDatasets),
  geocodeVariable: dispatch(moduleActions.geocodeVariable),
  importDataset: dispatch(moduleActions.importDataset),
  importFullDataset: dispatch(moduleActions.importFullDataset),
  importAvailableDataset: dispatch(moduleActions.importAvailableDataset),
  deleteVariable: dispatch(moduleActions.deleteVariable),
  setGrouping: dispatch(moduleActions.setGrouping),
  removeGrouping: dispatch(moduleActions.removeGrouping),
  updateGrouping: dispatch(moduleActions.updateGrouping),
  importDataFile: dispatch(moduleActions.importDataFile),
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
  fetchImageAttention: dispatch(moduleActions.fetchImageAttention),
  fetchTimeseries: dispatch(moduleActions.fetchTimeseries),
  fetchGraph: dispatch(moduleActions.fetchGraph),
  fetchFile: dispatch(moduleActions.fetchFile),
  // join data
  joinDatasetsPreview: dispatch(moduleActions.joinDatasetsPreview),
  fetchJoinDatasetsTableData: dispatch(
    moduleActions.fetchJoinDatasetsTableData
  ),
  fetchTableData: dispatch(moduleActions.fetchTableData),
  // included / excluded table data
  fetchIncludedTableData: dispatch(moduleActions.fetchIncludedTableData),
  fetchExcludedTableData: dispatch(moduleActions.fetchExcludedTableData),
  fetchHighlightedTableData: dispatch(moduleActions.fetchHighlightedTableData),
  fetchAreaOfInterestData: dispatch(moduleActions.fetchAreaOfInterestData),
  // task info
  fetchTask: dispatch(moduleActions.fetchTask),
  // multiband image band combinations
  fetchMultiBandCombinations: dispatch(
    moduleActions.fetchMultiBandCombinations
  ),
  // modeling metric methodologies
  fetchModelingMetrics: dispatch(moduleActions.fetchModelingMetrics),
  updateRowSelectionData: dispatch(moduleActions.updateRowSelectionData),
  cloneDataset: dispatch(moduleActions.cloneDataset),
  addField: dispatch(moduleActions.addField),
  updateDataset: dispatch(moduleActions.updateDataset),
  extractDataset: dispatch(moduleActions.extractDataset),
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
  removeFile: commit(moduleMutations.removeFile),
  bulkUpdateTimeseries: commit(moduleMutations.bulkUpdateTimeseries),
  // included / excluded table data
  setJoinDatasetsTableData: commit(moduleMutations.setJoinDatasetsTableData),
  clearJoinDatasetsTableData: commit(
    moduleMutations.clearJoinDatasetsTableData
  ),
  setHighlightedIncludeTableData: commit(
    moduleMutations.setHighlightedIncludeTableData
  ),
  setHighlightedExcludeTableData: commit(
    moduleMutations.setHighlightedExcludeTableData
  ),
  setAreaOfInterestIncludeInner: commit(
    moduleMutations.setAreaOfInterestIncludeInner
  ),
  setAreaOfInterestIncludeOuter: commit(
    moduleMutations.setAreaOfInterestIncludeOuter
  ),
  setAreaOfInterestExcludeInner: commit(
    moduleMutations.setAreaOfInterestExcludeInner
  ),
  setAreaOfInterestExcludeOuter: commit(
    moduleMutations.setAreaOfInterestExcludeOuter
  ),
  removeTimeseries: commit(moduleMutations.removeTimeseries),
  setIncludedTableData: commit(moduleMutations.setIncludedTableData),
  setExcludedTableData: commit(moduleMutations.setExcludedTableData),
  updateBands: commit(moduleMutations.updateBands),
  updateRowSelectionData: commit(moduleMutations.updateRowSelectionData),
  updateMetrics: commit(moduleMutations.updateMetrics),
};
