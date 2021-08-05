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

import Vue from "vue";
import { defaultState, PredictionState } from "./index";
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
  setBaselinePredictionTableData(
    state: PredictionState,
    predictionData: TableData
  ) {
    state.baselinePredictionTableData = Object.freeze(predictionData);
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
    state.areaOfInterestInner = Object.freeze(resultData);
  },
  setAreaOfInterestOuter(state: PredictionState, resultData: TableData) {
    state.areaOfInterestOuter = Object.freeze(resultData);
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
  updateConfidenceSummary(state: PredictionState, summary: VariableSummary) {
    updateSummaries(summary, state.confidenceSummaries);
  },
  updateRankSummary(state: PredictionState, summary: VariableSummary) {
    updateSummaries(summary, state.rankSummaries);
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
  resetState(state: PredictionState): void {
    Object.assign(state, defaultState());
  },
};
