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

import { Dictionary } from "lodash";
import Vue from "vue";
import { updateSummaries, updateSummariesPerVariable } from "../../util/data";
import { Extrema, TableData, VariableSummary } from "../dataset/index";
import { TimeSeriesForecastUpdate } from "../dataset/mutations";
import { ResultsState } from "./index";

export const mutations = {
  // training / target

  clearTrainingSummaries(state: ResultsState) {
    state.trainingSummaries = {};
  },

  clearCorrectnessSummaries(state: ResultsState) {
    state.correctnessSummaries = [];
  },

  clearTargetSummary(state: ResultsState) {
    state.targetSummary = null;
  },

  updateTrainingSummary(state: ResultsState, summary: VariableSummary) {
    updateSummariesPerVariable(summary, state.trainingSummaries);
  },

  updateTargetSummary(state: ResultsState, summary: VariableSummary) {
    state.targetSummary = summary;
  },

  // sets the current result data into the store
  setIncludedResultTableData(state: ResultsState, resultData: TableData) {
    // freezing the return to prevent slow, unnecessary deep reactivity.
    state.includedResultTableData = Object.freeze(resultData);
  },

  // sets the current result data into the store
  setExcludedResultTableData(state: ResultsState, resultData: TableData) {
    // freezing the return to prevent slow, unnecessary deep reactivity.
    state.excludedResultTableData = Object.freeze(resultData);
  },
  setFullExcludeResultTableData(state: ResultsState, resultData: TableData) {
    state.fullExcludedResultTableData = Object.freeze(resultData);
  },
  setFullIncludeResultTableData(state: ResultsState, resultData: TableData) {
    state.fullIncludedResultTableData = Object.freeze(resultData);
  },
  clearAreaOfInterestInner(state: ResultsState) {
    state.areaOfInterestInner = null;
  },
  clearAreaOfInterestOuter(state: ResultsState) {
    state.areaOfInterestOuter = null;
  },
  setAreaOfInterestInner(state: ResultsState, resultData: TableData) {
    state.areaOfInterestInner = resultData;
  },
  setAreaOfInterestOuter(state: ResultsState, resultData: TableData) {
    state.areaOfInterestOuter = resultData;
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
      max: null,
    };
  },

  // correctness

  updateCorrectnessSummaries(state: ResultsState, summary: VariableSummary) {
    updateSummaries(summary, state.correctnessSummaries);
  },
  // ranking
  updateRankingSummaries(state: ResultsState, summary: VariableSummary) {
    updateSummaries(summary, state.rankingSummaries);
  },
  // confidence

  updateConfidenceSummaries(state: ResultsState, summary: VariableSummary) {
    updateSummaries(summary, state.confidenceSummaries);
  },
  removeTimeseries(
    state: ResultsState,
    args: {
      solutionId: string;
      ids: string[];
    }
  ) {
    args.ids.forEach((id) => {
      // delete timeseries data
      delete state.timeseries[args.solutionId].timeseriesData[id];
      // delete is date time
      delete state.timeseries[args.solutionId].isDateTime[id];
      // delete info
      delete state.timeseries[args.solutionId].info[id];
      // remove predictedForecast data
      delete state.forecasts[args.solutionId].forecastData[id];
      // delete forecast range
      delete state.forecasts[args.solutionId].forecastRange[id];
      // delete isDateTime
      delete state.forecasts[args.solutionId].isDateTime[id];
    });
  },

  bulkUpdatePredictedTimeseries(
    state: ResultsState,
    args: {
      solutionId: string;
      uniqueTrail?: string;
      updates: TimeSeriesForecastUpdate[];
    }
  ) {
    args.updates.forEach((update) => {
      mutations.updatePredictedTimeseries(state, {
        solutionId: args.solutionId,
        uniqueTrail: args.uniqueTrail,
        update: update,
      });
    });
  },

  updatePredictedTimeseries(
    state: ResultsState,
    args: {
      solutionId: string;
      uniqueTrail: string;
      update: TimeSeriesForecastUpdate;
    }
  ) {
    if (!state.timeseries[args.solutionId]) {
      Vue.set(state.timeseries, args.solutionId, {});
    }

    if (!state.timeseries[args.solutionId].timeseriesData) {
      Vue.set(state.timeseries[args.solutionId], "timeseriesData", {});
    }

    const timeseriesKey =
      args.update.variableKey +
      (args.update.seriesID ?? "") +
      (args.uniqueTrail ?? "");

    // freezing the return to prevent slow, unnecessary deep reactivity.
    Vue.set(
      state.timeseries[args.solutionId].timeseriesData,
      timeseriesKey,
      Object.freeze(args.update.timeseries)
    );

    if (!state.timeseries[args.solutionId].isDateTime) {
      Vue.set(state.timeseries[args.solutionId], "isDateTime", {});
    }
    Vue.set(
      state.timeseries[args.solutionId].isDateTime,
      timeseriesKey,
      args.update.isDateTime
    );

    // Set the min/max/mean for each timeseries data
    if (!state.timeseries[args.solutionId].info) {
      Vue.set(state.timeseries[args.solutionId], "info", {});
    }
    Vue.set(state.timeseries[args.solutionId].info, timeseriesKey, {
      min: args.update.min as number,
      max: args.update.max as number,
      mean: args.update.mean as number,
    });
  },

  bulkUpdatePredictedForecast(
    state: ResultsState,
    args: {
      solutionId: string;
      uniqueTrail?: string;
      updates: TimeSeriesForecastUpdate[];
    }
  ) {
    args.updates.forEach((update) => {
      mutations.updatePredictedForecast(state, {
        solutionId: args.solutionId,
        uniqueTrail: args.uniqueTrail,
        update: update,
      });
    });
  },

  updatePredictedForecast(
    state: ResultsState,
    args: {
      solutionId: string;
      uniqueTrail: string;
      update: TimeSeriesForecastUpdate;
    }
  ) {
    const timeseriesKey =
      args.update.variableKey +
      (args.update.seriesID ?? "") +
      (args.uniqueTrail ?? "");

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
    // freezing the return to prevent slow, unnecessary deep reactivity.
    Vue.set(
      state.forecasts[args.solutionId].forecastData,
      timeseriesKey,
      Object.freeze(args.update.forecast)
    );
    Vue.set(
      state.forecasts[args.solutionId].forecastRange,
      timeseriesKey,
      args.update.forecastTestRange
    );
    Vue.set(
      state.forecasts[args.solutionId].isDateTime,
      timeseriesKey,
      args.update.isDateTime
    );
  },

  setFeatureImportanceRanking(
    state: ResultsState,
    args: { solutionID: string; rankings: Dictionary<number> }
  ) {
    Vue.set(state.featureImportanceRanking, args.solutionID, args.rankings);
  },
};
