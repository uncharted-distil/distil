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
  state: state
};

const { commit, read, dispatch } = getStoreAccessors<PredictionState, DistilState>(
  null
);

// Typed getters
export const getters = {
  // result
  getFittedSolutionIdFromPrediction: read(moduleGetters.getFittedSolutionIdFromPrediction),
  getProduceRequestIdFromPrediction: read(moduleGetters.getProduceRequestIdFromPrediction),
  hasIncludedPredictionTableData: read(moduleGetters.hasIncludedPredictionTableData),
  getIncludedPredictionTableData: read(moduleGetters.getIncludedPredictionTableData),
  getIncludedPredictionTableDataItems: read(
    moduleGetters.getIncludedPredictionTableDataItems
  ),
  getIncludedPredictionTableDataFields: read(
    moduleGetters.getIncludedPredictionTableDataFields
  ),
  hasExcludedPredictionTableData: read(moduleGetters.hasExcludedPredictionTableData),
  getExcludedPredictionTableData: read(moduleGetters.getExcludedPredictionTableData),
  getExcludedPredictionTableDataItems: read(
    moduleGetters.getExcludedPredictionTableDataItems
  ),
  getExcludedPredictionTableDataFields: read(
    moduleGetters.getExcludedPredictionTableDataFields
  ),
  // predicted
  getPredictionSummaries: read(moduleGetters.getPredictionSummaries),

  // result table data
  getPredictionDataNumRows: read(moduleGetters.getPredictionDataNumRows),
  // forecasts
  getPredictedTimeseries: read(moduleGetters.getPredictionTimeseries),
  getPredictedForecasts: read(moduleGetters.getPredictionForecasts)
};

// Typed actions
export const actions = {
  // training / target
  fetchTrainingSummaries: dispatch(moduleActions.fetchTrainingSummaries),
  fetchTargetSummary: dispatch(moduleActions.fetchTargetSummary),
  // result
  fetchIncludedPredictionTableData: dispatch(
    moduleActions.fetchIncludedPredictionTableData
  ),
  fetchExcludedPredictionTableData: dispatch(
    moduleActions.fetchExcludedPredictionTableData
  ),
  fetchPredictionTableData: dispatch(moduleActions.fetchPredictionTableData),
  // predicted
  fetchPredictedSummary: dispatch(moduleActions.fetchPredictionSummary),
  fetchPredictedSummaries: dispatch(moduleActions.fetchPredictionSummaries),
   
  // forecast
  fetchForecastedTimeseries: dispatch(moduleActions.fetchForecastedTimeseries)
};

// Typed mutations
export const mutations = {
  // training / target
  clearTrainingSummaries: commit(moduleMutations.clearTrainingSummaries),
  clearTargetSummary: commit(moduleMutations.clearTargetSummary),
  updateTrainingSummary: commit(moduleMutations.updateTrainingSummary),
  updateTargetSummary: commit(moduleMutations.updateTargetSummary),
  removeTrainingSummary: commit(moduleMutations.removeTrainingSummary),
  // result
  setIncludedPredictionTableData: commit(
    moduleMutations.setIncludedPredictionTableData
  ),
  setExcludedPredictionTableData: commit(
    moduleMutations.setExcludedPredictionTableData
  ),
  // predicted
  updatePredictedSummaries: commit(moduleMutations.updatePredictedSummaries),
  // forecasts
  updatePredictedTimeseries: commit(moduleMutations.updatePredictedTimeseries),
  updatePredictedForecast: commit(moduleMutations.updatePredictedForecast)
};
