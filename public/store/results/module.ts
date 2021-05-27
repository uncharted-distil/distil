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
import { ResultsState, state } from "./index";
import { mutations as moduleMutations } from "./mutations";

export const resultsModule: Module<ResultsState, DistilState> = {
  getters: moduleGetters,
  actions: moduleActions,
  mutations: moduleMutations,
  state: state,
  namespaced: true,
};

const { commit, read, dispatch } = getStoreAccessors<ResultsState, DistilState>(
  "resultsModule"
);

// Typed getters
export const getters = {
  // training / target
  getTrainingSummaries: read(moduleGetters.getTrainingSummaries),
  getTrainingSummariesDictionary: read(
    moduleGetters.getTrainingSummariesDictionary
  ),
  getTargetSummary: read(moduleGetters.getTargetSummary),
  // result
  getFittedSolutionId: read(moduleGetters.getFittedSolutionId),
  getProduceRequestId: read(moduleGetters.getProduceRequestId),
  hasIncludedResultTableData: read(moduleGetters.hasIncludedResultTableData),
  getIncludedResultTableData: read(moduleGetters.getIncludedResultTableData),
  getIncludedResultTableDataItems: read(
    moduleGetters.getIncludedResultTableDataItems
  ),
  getIncludedResultTableDataFields: read(
    moduleGetters.getIncludedResultTableDataFields
  ),
  getIncludedResultTableDataCount: read(
    moduleGetters.getIncludedResultTableDataCount
  ),
  hasExcludedResultTableData: read(moduleGetters.hasExcludedResultTableData),
  getExcludedResultTableData: read(moduleGetters.getExcludedResultTableData),
  getExcludedResultTableDataItems: read(
    moduleGetters.getExcludedResultTableDataItems
  ),
  getExcludedResultTableDataFields: read(
    moduleGetters.getExcludedResultTableDataFields
  ),
  getExcludedResultTableDataCount: read(
    moduleGetters.getExcludedResultTableDataCount
  ),
  getFullIncludedResultTableDataItems: read(
    moduleGetters.getFullIncludedResultTableDataItems
  ),
  hasResultTableDataItemsWeight: read(
    moduleGetters.hasResultTableDataItemsWeight
  ),
  getAreaOfInterestInnerDataItems: read(
    moduleGetters.getAreaOfInterestInnerDataItems
  ),
  getAreaOfInterestOuterDataItems: read(
    moduleGetters.getAreaOfInterestOuterDataItems
  ),
  // predicted
  getPredictedSummaries: read(moduleGetters.getPredictedSummaries),
  // residual
  getResidualsSummaries: read(moduleGetters.getResidualsSummaries),
  getResidualsExtrema: read(moduleGetters.getResidualsExtrema),
  // correctness
  getCorrectnessSummaries: read(moduleGetters.getCorrectnessSummaries),
  // confidence
  getConfidenceSummaries: read(moduleGetters.getConfidenceSummaries),
  // ranking
  getRankingSummaries: read(moduleGetters.getRankingSummaries),
  // result table data
  getResultDataNumRows: read(moduleGetters.getResultDataNumRows),
  // forecasts
  getPredictedTimeseries: read(moduleGetters.getPredictedTimeseries),
  getPredictedForecasts: read(moduleGetters.getPredictedForecasts),
  // rankings
  getFeatureImportanceRanking: read(moduleGetters.getFeatureImportanceRanking),
  // gets the number of records in db
  getNumOfRecords: read(moduleGetters.getNumOfRecords),
};

// Typed actions
export const actions = {
  // training / target
  fetchTrainingSummaries: dispatch(moduleActions.fetchTrainingSummaries),
  fetchTargetSummary: dispatch(moduleActions.fetchTargetSummary),
  // result
  fetchIncludedResultTableData: dispatch(
    moduleActions.fetchIncludedResultTableData
  ),
  fetchExcludedResultTableData: dispatch(
    moduleActions.fetchExcludedResultTableData
  ),
  fetchResultTableData: dispatch(moduleActions.fetchResultTableData),
  // predicted
  fetchPredictedSummary: dispatch(moduleActions.fetchPredictedSummary),
  fetchPredictedSummaries: dispatch(moduleActions.fetchPredictedSummaries),
  // residuals
  fetchResidualsSummary: dispatch(moduleActions.fetchResidualsSummary),
  fetchResidualsSummaries: dispatch(moduleActions.fetchResidualsSummaries),
  fetchResidualsExtrema: dispatch(moduleActions.fetchResidualsExtrema),
  // correctness
  fetchCorrectnessSummary: dispatch(moduleActions.fetchCorrectnessSummary),
  fetchCorrectnessSummaries: dispatch(moduleActions.fetchCorrectnessSummaries),
  // correctness
  fetchConfidenceSummary: dispatch(moduleActions.fetchConfidenceSummary),
  fetchConfidenceSummaries: dispatch(moduleActions.fetchConfidenceSummaries),
  // ranking
  fetchRankingSummary: dispatch(moduleActions.fetchRankingSummary),
  fetchRankingSummaries: dispatch(moduleActions.fetchRankingSummaries),
  // forecast
  fetchForecastedTimeseries: dispatch(moduleActions.fetchForecastedTimeseries),
  // variable rankings
  fetchFeatureImportanceRanking: dispatch(
    moduleActions.fetchFeatureImportanceRanking
  ),
  // area of interest for tile clicks
  fetchAreaOfInterestInner: dispatch(moduleActions.fetchAreaOfInterestInner),
  fetchAreaOfInterestOuter: dispatch(moduleActions.fetchAreaOfInterestOuter),
};

// Typed mutations
export const mutations = {
  // training / target
  clearTrainingSummaries: commit(moduleMutations.clearTrainingSummaries),
  clearTargetSummary: commit(moduleMutations.clearTargetSummary),
  updateTrainingSummary: commit(moduleMutations.updateTrainingSummary),
  updateTargetSummary: commit(moduleMutations.updateTargetSummary),
  // result
  setIncludedResultTableData: commit(
    moduleMutations.setIncludedResultTableData
  ),
  setExcludedResultTableData: commit(
    moduleMutations.setExcludedResultTableData
  ),
  setFullIncludedResultTableData: commit(
    moduleMutations.setFullIncludeResultTableData
  ),
  clearAreaOfInterestInner: commit(moduleMutations.clearAreaOfInterestInner),
  clearAreaOfInterestOuter: commit(moduleMutations.clearAreaOfInterestOuter),
  setAreaOfInterestInner: commit(moduleMutations.setAreaOfInterestInner),
  setAreaOfInterestOuter: commit(moduleMutations.setAreaOfInterestOuter),
  // predicted
  updatePredictedSummaries: commit(moduleMutations.updatePredictedSummaries),
  // residuals
  updateResidualsSummaries: commit(moduleMutations.updateResidualsSummaries),
  updateResidualsExtrema: commit(moduleMutations.updateResidualsExtrema),
  clearResidualsExtrema: commit(moduleMutations.clearResidualsExtrema),
  // correctness
  clearCorrectnessSummaries: commit(moduleMutations.clearCorrectnessSummaries),
  updateCorrectnessSummaries: commit(
    moduleMutations.updateCorrectnessSummaries
  ),
  updateRankingSummaries: commit(moduleMutations.updateRankingSummaries),
  // predicted
  updateConfidenceSummaries: commit(moduleMutations.updateConfidenceSummaries),
  // forecasts
  bulkUpdatePredictedTimeseries: commit(
    moduleMutations.bulkUpdatePredictedTimeseries
  ),
  bulkUpdatePredictedForecast: commit(
    moduleMutations.bulkUpdatePredictedForecast
  ),
  // variable rankings
  setFeatureImportanceRanking: commit(
    moduleMutations.setFeatureImportanceRanking
  ),
  removeTimeseries: commit(moduleMutations.removeTimeseries),
};
