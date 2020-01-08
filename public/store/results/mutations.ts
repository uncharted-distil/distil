import Vue from "vue";
import _ from "lodash";
import { ResultsState } from "./index";
import { VariableSummary, Extrema, TableData } from "../dataset/index";
import { updateSummaries, removeSummary } from "../../util/data";

export const mutations = {
  // training / target

  clearTrainingSummaries(state: ResultsState) {
    state.trainingSummaries = [];
  },

  clearCorrectnessSummaries(state: ResultsState) {
    state.correctnessSummaries = [];
  },

  clearTargetSummary(state: ResultsState) {
    state.targetSummary = null;
  },

  updateTrainingSummary(state: ResultsState, summary: VariableSummary) {
    updateSummaries(summary, state.trainingSummaries);
  },

  removeTrainingSummary(state: ResultsState, summary: VariableSummary) {
    removeSummary(summary, state.trainingSummaries);
  },

  updateTargetSummary(state: ResultsState, summary: VariableSummary) {
    state.targetSummary = summary;
  },

  // sets the current result data into the store
  setIncludedResultTableData(state: ResultsState, resultData: TableData) {
    state.includedResultTableData = resultData;
  },

  // sets the current result data into the store
  setExcludedResultTableData(state: ResultsState, resultData: TableData) {
    state.excludedResultTableData = resultData;
  },

  // predicted

  updatePredictedSummaries(state: ResultsState, summary: VariableSummary) {
    updateSummaries(summary, state.predictedSummaries);
  },

  // residuals

  updateResidualsSummaries(state: ResultsState, summary: VariableSummary) {
    updateSummaries(summary, state.residualSummaries);
  },

  updateResidualsExtrema(state: ResultsState, extrema: Extrema) {
    state.residualsExtrema = extrema;
  },

  clearResidualsExtrema(state: ResultsState) {
    state.residualsExtrema = {
      min: null,
      max: null
    };
  },

  // correctness

  updateCorrectnessSummaries(state: ResultsState, summary: VariableSummary) {
    updateSummaries(summary, state.correctnessSummaries);
  },

  // forecast

  updatePredictedTimeseries(
    state: ResultsState,
    args: { solutionId: string; id: string; timeseries: number[][] }
  ) {
    if (!state.timeseries[args.solutionId]) {
      Vue.set(state.timeseries, args.solutionId, {});
    }
    Vue.set(state.timeseries[args.solutionId], args.id, args.timeseries);
  },

  updatePredictedForecast(
    state: ResultsState,
    args: { solutionId: string; id: string; forecast: number[][] }
  ) {
    if (!state.forecasts[args.solutionId]) {
      Vue.set(state.forecasts, args.solutionId, {});
    }
    Vue.set(state.forecasts[args.solutionId], args.id, args.forecast);
  }
};
