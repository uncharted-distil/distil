import {
  VariableSummary,
  TableData,
  TableRow,
  TableColumn,
} from "../dataset/index";
import { PredictionState } from "./index";
import { getTableDataItems, getTableDataFields } from "../../util/data";
import { Dictionary } from "../../util/dict";
import { Forecast, TimeSeries } from "../results";

export const getters = {
  // results

  getPredictionDataNumRows(state: PredictionState): number {
    return state.includedPredictionTableData
      ? state.includedPredictionTableData.numRows
      : 0;
  },

  getFittedSolutionIdFromPrediction(state: PredictionState): string {
    return state.includedPredictionTableData.fittedSolutionId;
  },

  getProduceRequestIdFromPrediction(state: PredictionState): string {
    return state.includedPredictionTableData.produceRequestId;
  },

  hasIncludedPredictionTableData(state: PredictionState): boolean {
    return !!state.includedPredictionTableData;
  },

  getIncludedPredictionTableData(state: PredictionState): TableData {
    return state.includedPredictionTableData;
  },

  getIncludedPredictionTableDataItems(
    state: PredictionState,
    getters: any
  ): TableRow[] {
    return getTableDataItems(state.includedPredictionTableData);
  },

  getIncludedPredictionTableDataFields(
    state: PredictionState
  ): Dictionary<TableColumn> {
    return getTableDataFields(state.includedPredictionTableData);
  },

  // predicted

  getPredictionSummaries(state: PredictionState): VariableSummary[] {
    return state.predictedSummaries;
  },

  getTrainingSummaries(state: PredictionState): VariableSummary[] {
    return state.trainingSummaries.variableSummaries;
  },

  getTrainingSummariesDictionary(
    state: PredictionState
  ): Dictionary<Dictionary<VariableSummary>> {
    return state.trainingSummaries.variableSummariesByKey;
  },

  // forecasts

  getPredictionTimeseries(state: PredictionState): Dictionary<TimeSeries> {
    return state.timeseries;
  },

  getPredictionForecasts(state: PredictionState): Dictionary<Forecast> {
    return state.forecasts;
  },
};
