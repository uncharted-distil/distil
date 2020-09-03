import Vue from "vue";
import _ from "lodash";
import { PredictionState } from "./index";
import { VariableSummary, Extrema, TableData } from "../dataset/index";
import { updateSummaries, updateSummariesPerVariable } from "../../util/data";

export const mutations = {
  // training / target

  clearTrainingSummaries(state: PredictionState) {
    state.trainingSummaries = {};
  },

  updateTrainingSummary(state: PredictionState, summary: VariableSummary) {
    updateSummariesPerVariable(summary, state.trainingSummaries);
  },

  // sets the current Prediction data into the store
  setIncludedPredictionTableData(
    state: PredictionState,
    predictionData: TableData
  ) {
    state.includedPredictionTableData = predictionData;
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
      predictionsId: string;
      id: string;
      timeseries: number[][];
      isDateTime: boolean;
    }
  ) {
    if (!state.timeseries[args.predictionsId]) {
      Vue.set(state.timeseries, args.predictionsId, {});
    }
    if (!state.timeseries[args.predictionsId].timeseriesData) {
      Vue.set(state.timeseries[args.predictionsId], "timeseriesData", {});
    }
    if (!state.timeseries[args.predictionsId].isDateTime) {
      Vue.set(state.timeseries[args.predictionsId], "isDateTime", {});
    }
    Vue.set(
      state.timeseries[args.predictionsId].timeseriesData,
      args.id,
      args.timeseries
    );
    Vue.set(
      state.timeseries[args.predictionsId].isDateTime,
      args.id,
      args.isDateTime
    );
  },

  updatePredictedForecast(
    state: PredictionState,
    args: {
      predictionsId: string;
      id: string;
      forecast: number[][];
      forecastTestRange: number[];
      isDateTime: boolean;
    }
  ) {
    if (!state.forecasts[args.predictionsId]) {
      Vue.set(state.forecasts, args.predictionsId, {});
    }
    if (!state.forecasts[args.predictionsId].forecastData) {
      Vue.set(state.forecasts[args.predictionsId], "forecastData", {});
    }
    if (!state.forecasts[args.predictionsId].forecastRange) {
      Vue.set(state.forecasts[args.predictionsId], "forecastRange", {});
    }
    if (!state.forecasts[args.predictionsId].isDateTime) {
      Vue.set(state.forecasts[args.predictionsId], "isDateTime", {});
    }
    Vue.set(
      state.forecasts[args.predictionsId].forecastData,
      args.id,
      args.forecast
    );
    Vue.set(
      state.forecasts[args.predictionsId].forecastRange,
      args.id,
      args.forecastTestRange
    );
    Vue.set(
      state.forecasts[args.predictionsId].isDateTime,
      args.id,
      args.isDateTime
    );
  },
};
