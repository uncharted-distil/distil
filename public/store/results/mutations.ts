import Vue from "vue";
import { Dictionary } from "vue-router/types/router";
import { updateSummaries, updateSummariesPerVariable } from "../../util/data";
import {
  Extrema,
  TableData,
  TimeSeriesValue,
  VariableSummary,
} from "../dataset/index";
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
      max: null,
    };
  },

  // correctness

  updateCorrectnessSummaries(state: ResultsState, summary: VariableSummary) {
    updateSummaries(summary, state.correctnessSummaries);
  },

  // forecast

  updatePredictedTimeseries(
    state: ResultsState,
    args: {
      solutionId: string;
      id: string;
      timeseries: number[][];
      isDateTime: boolean;
      min: number;
      max: number;
      mean: number;
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
      Object.freeze(args.timeseries)
    );

    if (!state.timeseries[args.solutionId].isDateTime) {
      Vue.set(state.timeseries[args.solutionId], "isDateTime", {});
    }
    Vue.set(
      state.timeseries[args.solutionId].isDateTime,
      args.id,
      args.isDateTime
    );

    // Set the min/max/mean for each timeseries data
    if (!state.timeseries[args.solutionId].info) {
      Vue.set(state.timeseries[args.solutionId], "info", {});
    }
    Vue.set(state.timeseries[args.solutionId].info, args.id, {
      min: args.min as number,
      max: args.max as number,
      mean: args.mean as number,
    });
  },

  updatePredictedForecast(
    state: ResultsState,
    args: {
      solutionId: string;
      id: string;
      forecast: TimeSeriesValue[];
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
      Object.freeze(args.forecast)
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
  },

  setFeatureImportanceRanking(
    state: ResultsState,
    args: { solutionID: string; rankings: Dictionary<number> }
  ) {
    Vue.set(state.featureImportanceRanking, args.solutionID, args.rankings);
  },
};
