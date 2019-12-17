import { Module } from "vuex";
import { state, ResultsState } from "./index";
import { getters as moduleGetters } from "./getters";
import { actions as moduleActions } from "./actions";
import { mutations as moduleMutations } from "./mutations";
import { DistilState } from "../store";
import { getStoreAccessors } from "vuex-typescript";

export const resultsModule: Module<ResultsState, DistilState> = {
  getters: moduleGetters,
  actions: moduleActions,
  mutations: moduleMutations,
  state: state
};

const { commit, read, dispatch } = getStoreAccessors<ResultsState, DistilState>(
  null
);

// Typed getters
export const getters = {
  // training / target
  getTrainingSummaries: read(moduleGetters.getTrainingSummaries),
  getTargetSummary: read(moduleGetters.getTargetSummary),
  // result
  hasIncludedResultTableData: read(moduleGetters.hasIncludedResultTableData),
  getIncludedResultTableData: read(moduleGetters.getIncludedResultTableData),
  getIncludedResultTableDataItems: read(
    moduleGetters.getIncludedResultTableDataItems
  ),
  getIncludedResultTableDataFields: read(
    moduleGetters.getIncludedResultTableDataFields
  ),
  hasExcludedResultTableData: read(moduleGetters.hasExcludedResultTableData),
  getExcludedResultTableData: read(moduleGetters.getExcludedResultTableData),
  getExcludedResultTableDataItems: read(
    moduleGetters.getExcludedResultTableDataItems
  ),
  getExcludedResultTableDataFields: read(
    moduleGetters.getExcludedResultTableDataFields
  ),
  // predicted
  getPredictedSummaries: read(moduleGetters.getPredictedSummaries),
  // residual
  getResidualsSummaries: read(moduleGetters.getResidualsSummaries),
  getResidualsExtrema: read(moduleGetters.getResidualsExtrema),
  // correctness
  getCorrectnessSummaries: read(moduleGetters.getCorrectnessSummaries),
  // result table data
  getResultDataNumRows: read(moduleGetters.getResultDataNumRows),
  // forecasts
  getPredictedTimeseries: read(moduleGetters.getPredictedTimeseries),
  getPredictedForecasts: read(moduleGetters.getPredictedForecasts)
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
  setIncludedResultTableData: commit(
    moduleMutations.setIncludedResultTableData
  ),
  setExcludedResultTableData: commit(
    moduleMutations.setExcludedResultTableData
  ),
  // predicted
  updatePredictedSummaries: commit(moduleMutations.updatePredictedSummaries),
  // residuals
  updateResidualsSummaries: commit(moduleMutations.updateResidualsSummaries),
  updateResidualsExtrema: commit(moduleMutations.updateResidualsExtrema),
  clearResidualsExtrema: commit(moduleMutations.clearResidualsExtrema),
  // correctness
  updateCorrectnessSummaries: commit(
    moduleMutations.updateCorrectnessSummaries
  ),
  // forecasts
  updatePredictedTimeseries: commit(moduleMutations.updatePredictedTimeseries),
  updatePredictedForecast: commit(moduleMutations.updatePredictedForecast)
};
