import {
  VariableSummary,
  Extrema,
  TableData,
  TableRow,
  TableColumn
} from "../dataset/index";
import { PredictionState } from "./index";
import { getTableDataItems, getTableDataFields } from "../../util/data";
import { Dictionary } from "../../util/dict";

export const getters = {
  // results

  getPredictionDataNumRows(state: PredictionState): number {
    return state.includedPredictionTableData
      ? state.includedPredictionTableData.numRows
      : 0;
  },

  getFittedSolutionIdFromPrediction (state: PredictionState): string {
    return state.includedPredictionTableData.fittedSolutionId;
  },

  getProduceRequestIdFromPrediction (state: PredictionState): string {
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

  hasExcludedPredictionTableData(state: PredictionState): boolean {
    return !!state.excludedPredictionTableData;
  },

  getExcludedPredictionTableData(state: PredictionState): TableData {
    return state.excludedPredictionTableData;
  },

  getExcludedPredictionTableDataItems(
    state: PredictionState,
    getters: any
  ): TableRow[] {
    return getTableDataItems(state.excludedPredictionTableData);
  },

  getExcludedPredictionTableDataFields(
    state: PredictionState
  ): Dictionary<TableColumn> {
    return getTableDataFields(state.excludedPredictionTableData);
  },
    
  // predicted

  getPredictionSummaries(state: PredictionState): VariableSummary[] {
    return state.predictedSummaries;
  },


  // forecasts

  getPredictionTimeseries(
    state: PredictionState
  ): Dictionary<Dictionary<number[][]>> {
    return state.timeseries;
  },

  getPredictionForecasts(
    state: PredictionState
  ): Dictionary<Dictionary<number[][]>> {
    return state.forecasts;
  }
};
