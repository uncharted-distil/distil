import Vue from "vue";
import _ from "lodash";
import { PredictionState } from "./index";
import { VariableSummary, Extrema, TableData } from "../dataset/index";
import { updateSummaries, removeSummary } from "../../util/data";

export const mutations = {
  // training / target

  clearTrainingSummaries(state: PredictionState) {
    state.trainingSummaries = [];
  },

  clearTargetSummary(state: PredictionState) {
    state.targetSummary = null;
  },

  updateTrainingSummary(state: PredictionState, summary: VariableSummary) {
    updateSummaries(summary, state.trainingSummaries);
  },

  removeTrainingSummary(state: PredictionState, summary: VariableSummary) {
    removeSummary(summary, state.trainingSummaries);
  },

  updateTargetSummary(state: PredictionState, summary: VariableSummary) {
    state.targetSummary = summary;
  },

  // sets the current Prediction data into the store
  setIncludedPredictionTableData(
    state: PredictionState,
    predictionData: TableData
  ) {
    state.includedPredictionTableData = predictionData;
  },

  // sets the current Prediction data into the store
  setExcludedPredictionTableData(
    state: PredictionState,
    predictionData: TableData
  ) {
    state.excludedPredictionTableData = predictionData;
  },

  // predicted

  updatePredictedSummaries(state: PredictionState, summary: VariableSummary) {
    updateSummaries(summary, state.predictedSummaries);
  },

  // forecast

  updatePredictedTimeseries(
    state: PredictionState,
    args: { solutionId: string; id: string; timeseries: number[][] }
  ) {
    if (!state.timeseries[args.solutionId]) {
      Vue.set(state.timeseries, args.solutionId, {});
    }
    Vue.set(state.timeseries[args.solutionId], args.id, args.timeseries);
  },

  updatePredictedForecast(
    state: PredictionState,
    args: {
      solutionId: string;
      id: string;
      forecast: number[][];
      forecastTestRange: number[];
    }
  ) {
    if (!state.forecasts[args.solutionId]) {
      Vue.set(state.forecasts, args.solutionId, {});
    }
    if (!state.forecasts[args.solutionId].forecastData) {
      Vue.set(state.forecasts[args.solutionId], "forecastData", {});
    }
    if (!state.forecasts[args.solutionId].forecastRange) {
      Vue.set(state.forecasts[args.solutionId], "forecastRange", {});
    }
    Vue.set(
      state.forecasts[args.solutionId].forecastData,
      args.id,
      args.forecast
    );
    Vue.set(
      state.forecasts[args.solutionId].forecastRange,
      args.id,
      args.forecastTestRange
    );
  }
};
