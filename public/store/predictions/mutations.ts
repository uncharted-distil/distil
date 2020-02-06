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

  updateTrainingSummary(state: PredictionState, summary: VariableSummary) {
    updateSummaries(summary, state.trainingSummaries);
  },

  removeTrainingSummary(state: PredictionState, summary: VariableSummary) {
    removeSummary(summary, state.trainingSummaries);
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
  updatePredictedSummary(state: PredictionState, summary: VariableSummary) {
    updateSummaries(summary, state.predictedSummaries);
  },

  clearPredictedSummary(state: PredictionState) {
    state.predictedSummaries = [];
  },

  // forecast

  updatePredictedTimeseries(
    state: PredictionState,
    args: {
      solutionId: string;
      id: string;
      timeseries: number[][];
      isDateTime: boolean;
    }
  ) {
    if (!state.timeseries[args.solutionId]) {
      Vue.set(state.timeseries, args.solutionId, {});
    }
    if (!state.timeseries[args.solutionId].timeseriesData) {
      Vue.set(state.timeseries[args.solutionId], "timeseriesData", {});
    }
    Vue.set(
      state.timeseries[args.solutionId].timeseriesData,
      args.id,
      args.timeseries
    );
    Vue.set(
      state.timeseries[args.solutionId].isDateTime,
      args.id,
      args.isDateTime
    );
  },

  updatePredictedForecast(
    state: PredictionState,
    args: {
      solutionId: string;
      id: string;
      forecast: number[][];
      forecastTestRange: number[];
      isDateTime: boolean;
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
    if (!state.forecasts[args.solutionId].isDateTime) {
      Vue.set(state.forecasts[args.solutionId], "isDateTime", {});
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
    Vue.set(
      state.forecasts[args.solutionId].isDateTime,
      args.id,
      args.isDateTime
    );
  }
};
