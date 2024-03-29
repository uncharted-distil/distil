/**
 *
 *    Copyright © 2021 Uncharted Software Inc.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

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
  getVariables: read(moduleGetters.getVariables), //filters hidden variables.
  getAllVariables: read(moduleGetters.getAllVariables), //includes hidden variables.
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
  getBaselineIncludeTableDataItems: read(
    moduleGetters.getBaselineIncludeTableDataItems
  ),
  getBaselineExcludeTableDataItems: read(
    moduleGetters.getBaselineExcludeTableDataItems
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
  deleteDataset: dispatch(moduleActions.deleteDataset),
  searchDatasets: dispatch(moduleActions.searchDatasets),
  geocodeVariable: dispatch(moduleActions.geocodeVariable),
  importDataset: dispatch(moduleActions.importDataset),
  importFullDataset: dispatch(moduleActions.importFullDataset),
  importAvailableDataset: dispatch(moduleActions.importAvailableDataset),
  clearVariable: dispatch(moduleActions.clearVariable),
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

  // Outliers detection
  fetchOutliers: dispatch(moduleActions.fetchOutliers),
  applyOutliers: dispatch(moduleActions.applyOutliers),

  // files
  fetchFiles: dispatch(moduleActions.fetchFiles),
  fetchImage: dispatch(moduleActions.fetchImage),
  fetchImagePack: dispatch(moduleActions.fetchImagePack),
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
  fetchBaselineTableData: dispatch(moduleActions.fetchBaselineTableData),
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
  saveDataset: dispatch(moduleActions.saveDataset),
  resetState: dispatch(moduleActions.resetState),
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
  clearVariableSummary: commit(moduleMutations.clearVariableSummary),
  clearVariableSummaries: commit(moduleMutations.clearVariableSummaries),
  setVariableSummary: commit(moduleMutations.setIncludedVariableSummary),
  // ranking
  setVariableRankings: commit(moduleMutations.setVariableRankings),
  updateVariableRankings: commit(moduleMutations.updateVariableRankings),
  // pending update
  updatePendingRequests: commit(moduleMutations.updatePendingRequests),
  removePendingRequest: commit(moduleMutations.removePendingRequest),
  // files
  updateFile: commit(moduleMutations.updateFile),
  removeFile: commit(moduleMutations.removeFile),
  bulkRemoveFiles: commit(moduleMutations.bulkRemoveFiles),
  bulkUpdateFiles: commit(moduleMutations.bulkUpdateFiles),
  bulkUpdateTimeseries: commit(moduleMutations.bulkUpdateTimeseries),
  // included / excluded table data
  setJoinDatasetsTableData: commit(moduleMutations.setJoinDatasetsTableData),
  clearJoinDatasetsTableData: commit(
    moduleMutations.clearJoinDatasetsTableData
  ),
  setBaselineIncludeTableData: commit(
    moduleMutations.setBaselineIncludeTableData
  ),
  setBaselineExcludeTableData: commit(
    moduleMutations.setBaselineExcludeTableData
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
  clearAreaOfInterestIncludeInner: commit(
    moduleMutations.clearAreaOfInterestIncludeInner
  ),
  clearAreaOfInterestIncludeOuter: commit(
    moduleMutations.clearAreaOfInterestIncludeOuter
  ),
  clearAreaOfInterestExcludeInner: commit(
    moduleMutations.clearAreaOfInterestExcludeInner
  ),
  clearAreaOfInterestExcludeOuter: commit(
    moduleMutations.clearAreaOfInterestExcludeOuter
  ),
  updateAreaOfInterestIncludeInner: commit(
    moduleMutations.updateAreaOfInterestIncludeInner
  ),
  removeTimeseries: commit(moduleMutations.removeTimeseries),
  setIncludedTableData: commit(moduleMutations.setIncludedTableData),
  setExcludedTableData: commit(moduleMutations.setExcludedTableData),
  updateBands: commit(moduleMutations.updateBands),
  updateRowSelectionData: commit(moduleMutations.updateRowSelectionData),
  updateMetrics: commit(moduleMutations.updateMetrics),
  resetState: commit(moduleMutations.resetState),
};
