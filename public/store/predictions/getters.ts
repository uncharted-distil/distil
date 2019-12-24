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

  getTrainingSummaries(state: PredictionState): VariableSummary[] {
    return state.trainingSummaries;
  },

  getTargetSummary(state: PredictionState): VariableSummary {
    return state.targetSummary;
  },

  getResultDataNumRows(state: PredictionState): number {
    return state.includedResultTableData
      ? state.includedResultTableData.numRows
      : 0;
  },

  getFittedSolutionId (state: PredictionState): string {
    return state.includedResultTableData.fittedSolutionId;
  },

  getProduceRequestId (state: PredictionState): string {
    return state.includedResultTableData.produceRequestId;
  },

  hasIncludedResultTableData(state: PredictionState): boolean {
    return !!state.includedResultTableData;
  },

  getIncludedResultTableData(state: PredictionState): TableData {
    return state.includedResultTableData;
  },

  getIncludedResultTableDataItems(
    state: PredictionState,
    getters: any
  ): TableRow[] {
    return getTableDataItems(state.includedResultTableData);
  },

  getIncludedResultTableDataFields(
    state: PredictionState
  ): Dictionary<TableColumn> {
    return getTableDataFields(state.includedResultTableData);
  },

  hasExcludedResultTableData(state: PredictionState): boolean {
    return !!state.excludedResultTableData;
  },

  getExcludedResultTableData(state: PredictionState): TableData {
    return state.excludedResultTableData;
  },

  getExcludedResultTableDataItems(
    state: PredictionState,
    getters: any
  ): TableRow[] {
    return getTableDataItems(state.excludedResultTableData);
  },

  getExcludedResultTableDataFields(
    state: PredictionState
  ): Dictionary<TableColumn> {
    return getTableDataFields(state.excludedResultTableData);
  },

  // predicted

  getPredictedSummaries(state: PredictionState): VariableSummary[] {
    return state.predictedSummaries;
  },

  // residual

  getResidualsSummaries(state: PredictionState): VariableSummary[] {
    return state.residualSummaries;
  },

  getResidualsExtrema(state: PredictionState): Extrema {
    return state.residualsExtrema;
  },

  // correctness

  getCorrectnessSummaries(state: PredictionState): VariableSummary[] {
    return state.correctnessSummaries;
  },

  // forecasts

  getPredictedTimeseries(
    state: PredictionState
  ): Dictionary<Dictionary<number[][]>> {
    return state.timeseries;
  },

  getPredictedForecasts(
    state: PredictionState
  ): Dictionary<Dictionary<number[][]>> {
    return state.forecasts;
  }
};
