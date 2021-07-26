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
import { state, PredictionState } from "./index";
import { getters as moduleGetters } from "./getters";
import { actions as moduleActions } from "./actions";
import { mutations as moduleMutations } from "./mutations";
import { DistilState } from "../store";
import { getStoreAccessors } from "vuex-typescript";

export const predictionsModule: Module<PredictionState, DistilState> = {
  getters: moduleGetters,
  actions: moduleActions,
  mutations: moduleMutations,
  state: state,
  namespaced: true,
};

const { commit, read, dispatch } = getStoreAccessors<
  PredictionState,
  DistilState
>("predictionsModule");

// Typed getters
export const getters = {
  getAreaOfInterestInnerDataItems: read(
    moduleGetters.getAreaOfInterestInnerDataItems
  ),
  getAreaOfInterestOuterDataItems: read(
    moduleGetters.getAreaOfInterestOuterDataItems
  ),
  // result
  getFittedSolutionIdFromPrediction: read(
    moduleGetters.getFittedSolutionIdFromPrediction
  ),
  getProduceRequestIdFromPrediction: read(
    moduleGetters.getProduceRequestIdFromPrediction
  ),
  hasIncludedPredictionTableData: read(
    moduleGetters.hasIncludedPredictionTableData
  ),
  getIncludedPredictionTableData: read(
    moduleGetters.getIncludedPredictionTableData
  ),
  getIncludedPredictionTableDataItems: read(
    moduleGetters.getIncludedPredictionTableDataItems
  ),
  getBaselinePredictionTableDataItems: read(
    moduleGetters.getBaselinePredictionTableDataItems
  ),
  getIncludedPredictionTableDataFields: read(
    moduleGetters.getIncludedPredictionTableDataFields
  ),

  // predicted
  getPredictionSummaries: read(moduleGetters.getPredictionSummaries),
  getTrainingSummariesDictionary: read(
    moduleGetters.getTrainingSummariesDictionary
  ),
  getConfidenceSummaries: read(moduleGetters.getConfidenceSummaries),
  getRankSummaries: read(moduleGetters.getRankSummaries),
  // result table data
  getPredictionDataNumRows: read(moduleGetters.getPredictionDataNumRows),

  // forecasts
  getPredictedTimeseries: read(moduleGetters.getPredictionTimeseries),
  getPredictedForecasts: read(moduleGetters.getPredictionForecasts),
};

// Typed actions
export const actions = {
  //areaOfInterest for geoplot
  fetchAreaOfInterestInner: dispatch(moduleActions.fetchAreaOfInterestInner),
  fetchAreaOfInterestOuter: dispatch(moduleActions.fetchAreaOfInterestOuter),
  // input inference data
  fetchTrainingSummaries: dispatch(moduleActions.fetchTrainingSummaries),

  // result table data
  fetchIncludedPredictionTableData: dispatch(
    moduleActions.fetchIncludedPredictionTableData
  ),
  fetchPredictionTableData: dispatch(moduleActions.fetchPredictionTableData),

  // predicted value summary
  fetchPredictedSummary: dispatch(moduleActions.fetchPredictionSummary),
  fetchPredictedSummaries: dispatch(moduleActions.fetchPredictionSummaries),
  fetchConfidenceSummary: dispatch(moduleActions.fetchConfidenceSummary),
  fetchRankSummary: dispatch(moduleActions.fetchRankSummary),
  // time series forecast data
  fetchForecastedTimeseries: dispatch(moduleActions.fetchForecastedTimeseries),
  // csv export data
  fetchExportData: dispatch(moduleActions.fetchExportData),
  // cloning results of a prediction to a new dataset
  createDataset: dispatch(moduleActions.createDataset),
  resetState: dispatch(moduleActions.resetState),
};

// Typed mutations
export const mutations = {
  //AreaofInterest for geoplot
  setAreaOfInterestInner: commit(moduleMutations.setAreaOfInterestInner),
  setAreaOfInterestOuter: commit(moduleMutations.setAreaOfInterestOuter),
  clearAreaOfInterestInner: commit(moduleMutations.clearAreaOfInterestInner),
  clearAreaOfInterestOuter: commit(moduleMutations.clearAreaOfInterestOuter),
  // training
  clearTrainingSummaries: commit(moduleMutations.clearTrainingSummaries),
  updateTrainingSummary: commit(moduleMutations.updateTrainingSummary),
  // result
  setIncludedPredictionTableData: commit(
    moduleMutations.setIncludedPredictionTableData
  ),
  // predicted
  clearPredictedSummary: commit(moduleMutations.clearPredictedSummary),
  updatePredictedSummary: commit(moduleMutations.updatePredictedSummary),
  setBaselinePredictionTableData: commit(
    moduleMutations.setBaselinePredictionTableData
  ),
  updateConfidenceSummary: commit(moduleMutations.updateConfidenceSummary),
  updateRankSummary: commit(moduleMutations.updateRankSummary),
  // forecasts
  updatePredictedTimeseries: commit(moduleMutations.updatePredictedTimeseries),
  bulkUpdatePredictedTimeseries: commit(
    moduleMutations.bulkUpdatePredictedTimeseries
  ),
  updatePredictedForecast: commit(moduleMutations.updatePredictedForecast),
  bulkUpdatePredictedForecast: commit(
    moduleMutations.bulkUpdatePredictedForecast
  ),
  removeTimeseries: commit(moduleMutations.removeTimeseries),
  resetState: commit(moduleMutations.resetState),
};
