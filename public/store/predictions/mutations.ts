import Vue from "vue";
import { PredictionState } from "./index";
import { VariableSummary, TableData } from "../dataset/index";
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
    // freezing the return to prevent slow, unnecessary deep reactivity.
    state.includedPredictionTableData = Object.freeze(predictionData);
  },
  setAreaOfInterestInner(state: PredictionState, resultData: TableData) {
    state.areaOfInterestInner = resultData;
  },
  setAreaOfInterestOuter(state: PredictionState, resultData: TableData) {
    state.areaOfInterestOuter = resultData;
  },
  clearAreaOfInterestInner(state: PredictionState) {
    state.areaOfInterestInner = null;
  },
  clearAreaOfInterestOuter(state: PredictionState) {
    state.areaOfInterestOuter = null;
  },
  // predicted
  updatePredictedSummary(state: PredictionState, summary: VariableSummary) {
    updateSummaries(summary, state.predictedSummaries);
  },

  clearPredictedSummary(state: PredictionState) {
    state.predictedSummaries = [];
  },
  removeTimeseries(
    state: PredictionState,
    args: { predictionsId: string; ids: string[] }
  ) {
    args.ids.forEach((id) => {
      // predicted data
      delete state.timeseries[args.predictionsId].timeseriesData[id];
      delete state.timeseries[args.predictionsId].isDateTime[id];
      delete state.timeseries[args.predictionsId].info[id];
      // predicted forecast
      delete state.forecasts[args.predictionsId].forecastData[id];
      delete state.forecasts[args.predictionsId].forecastRange[id];
      delete state.forecasts[args.predictionsId].isDateTime[id];
    });
  },
  // forecast
  bulkUpdatePredictedTimeseries(
    state: PredictionState,
    args: {
      predictionsId: string;
      map: Map<
        string,
        {
          timeseries: number[][];
          isDateTime: boolean;
          min: number;
          max: number;
          mean: number;
        }
      >;
    }
  ) {
    args.map.forEach((val, key) => {
      mutations.updatePredictedTimeseries(state, {
        predictionsId: args.predictionsId,
        id: key,
        timeseries: val.timeseries,
        isDateTime: val.isDateTime,
        min: val.min,
        max: val.max,
        mean: val.mean,
      });
    });
  },
  updatePredictedTimeseries(
    state: PredictionState,
    args: {
      predictionsId: string;
      id: string;
      timeseries: number[][];
      isDateTime: boolean;
      min: number;
      max: number;
      mean: number;
    }
  ) {
    if (!state.timeseries[args.predictionsId]) {
      Vue.set(state.timeseries, args.predictionsId, {});
    }

    if (!state.timeseries[args.predictionsId].timeseriesData) {
      Vue.set(state.timeseries[args.predictionsId], "timeseriesData", {});
    }
    Vue.set(
      state.timeseries[args.predictionsId].timeseriesData,
      args.id,
      args.timeseries
    );

    if (!state.timeseries[args.predictionsId].isDateTime) {
      Vue.set(state.timeseries[args.predictionsId], "isDateTime", {});
    }
    Vue.set(
      state.timeseries[args.predictionsId].isDateTime,
      args.id,
      args.isDateTime
    );

    // Set the min/max/mean for each timeseries data
    if (!state.timeseries[args.predictionsId].info) {
      Vue.set(state.timeseries[args.predictionsId], "info", {});
    }
    Vue.set(state.timeseries[args.predictionsId].info, args.id, {
      min: args.min as number,
      max: args.max as number,
      mean: args.mean as number,
    });
  },
  bulkUpdatePredictedForecast(
    state: PredictionState,
    args: {
      predictionsId: string;
      map: Map<
        string,
        {
          forecast: number[][];
          forecastTestRange: number[];
          isDateTime: boolean;
        }
      >;
    }
  ) {
    args.map.forEach((val, key) => {
      mutations.updatePredictedForecast(state, {
        predictionsId: args.predictionsId,
        id: key,
        forecast: val.forecast,
        forecastTestRange: val.forecastTestRange,
        isDateTime: val.isDateTime,
      });
    });
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
