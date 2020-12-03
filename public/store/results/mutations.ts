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
  setAreaOfInterestInner(state: ResultsState, resultData: TableData) {
    state.areaOfInterestInner = Object.freeze(resultData);
  },
  setAreaOfInterestOuter(state: ResultsState, resultData: TableData) {
    state.areaOfInterestOuter = Object.freeze(resultData);
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
      Vue.delete(state.timeseries[args.solutionId].timeseriesData, id);
      // delete is date time
      Vue.delete(state.timeseries[args.solutionId].isDateTime, id);
      // delete info
      Vue.delete(state.timeseries[args.solutionId].info, id);
      // remove predictedForecast data
      Vue.delete(state.forecasts[args.solutionId].forecastData, id);
      // delete forecast range
      Vue.delete(state.forecasts[args.solutionId].forecastRange, id);
      // delete isDateTime
      Vue.delete(state.forecasts[args.solutionId].isDateTime, id);
    });
  },
  bulkUpdatePredictedTimeseries(
    state: ResultsState,
    args: {
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
      solutionId: string;
      uniqueTrail?: string;
    }
  ) {
    args.map.forEach((val, key) => {
      mutations.updatePredictedTimeseries(state, {
        solutionId: args.solutionId,
        id: key + (args.uniqueTrail ?? ""),
        timeseries: val.timeseries,
        isDateTime: val.isDateTime,
        min: val.min,
        max: val.max,
        mean: val.mean,
      });
    });
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
    // freezing the return to prevent slow, unnecessary deep reactivity.
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
  bulkUpdatePredictedForecast(
    state: ResultsState,
    args: {
      solutionId: string;
      uniqueTrail?: string;
      map: Map<
        string,
        {
          forecast: TimeSeriesValue[];
          forecastTestRange: number[];
          isDateTime: boolean;
        }
      >;
    }
  ) {
    args.map.forEach((val, key) => {
      mutations.updatePredictedForecast(state, {
        solutionId: args.solutionId,
        id: key + (args.uniqueTrail ?? ""),
        forecast: val.forecast,
        forecastTestRange: val.forecastTestRange,
        isDateTime: val.isDateTime,
      });
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
    // freezing the return to prevent slow, unnecessary deep reactivity.
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
