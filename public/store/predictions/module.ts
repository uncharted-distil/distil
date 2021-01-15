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
  getIncludedPredictionTableDataFields: read(
    moduleGetters.getIncludedPredictionTableDataFields
  ),

  // predicted
  getPredictionSummaries: read(moduleGetters.getPredictionSummaries),
  getTrainingSummariesDictionary: read(
    moduleGetters.getTrainingSummariesDictionary
  ),

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

  // time series forecast data
  fetchForecastedTimeseries: dispatch(moduleActions.fetchForecastedTimeseries),
  // csv export data
  fetchExportData: dispatch(moduleActions.fetchExportData),
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
};
