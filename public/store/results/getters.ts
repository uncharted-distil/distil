import {
  VariableSummary,
  Extrema,
  TableData,
  TableRow,
  TableColumn,
} from "../dataset/index";
import { ResultsState, Forecast, TimeSeries } from "./index";
import { getTableDataItems, getTableDataFields } from "../../util/data";
import { Dictionary } from "../../util/dict";

export const getters = {
  // results

  getTrainingSummaries(state: ResultsState): VariableSummary[] {
    return state.trainingSummaries.variableSummaries;
  },

  getTrainingSummariesDictionary(
    state: ResultsState
  ): Dictionary<Dictionary<VariableSummary>> {
    return state.trainingSummaries.variableSummariesByKey;
  },

  getTargetSummary(state: ResultsState): VariableSummary {
    return state.targetSummary;
  },

  getResultDataNumRows(state: ResultsState): number {
    return state.includedResultTableData
      ? state.includedResultTableData.numRows
      : 0;
  },

  getFittedSolutionId(state: ResultsState): string {
    return state.includedResultTableData.fittedSolutionId;
  },

  getProduceRequestId(state: ResultsState): string {
    return state.includedResultTableData.produceRequestId;
  },

  hasIncludedResultTableData(state: ResultsState): boolean {
    return !!state.includedResultTableData;
  },

  getIncludedResultTableData(state: ResultsState): TableData {
    return state.includedResultTableData;
  },

  getIncludedResultTableDataItems(
    state: ResultsState,
    getters: any
  ): TableRow[] {
    return getTableDataItems(state.includedResultTableData);
  },

  getIncludedResultTableDataFields(
    state: ResultsState
  ): Dictionary<TableColumn> {
    return getTableDataFields(state.includedResultTableData);
  },

  getIncludedResultTableDataCount(state: ResultsState): number {
    return state.includedResultTableData.numRowsFiltered
      ? state.includedResultTableData.numRowsFiltered
      : 0;
  },

  hasExcludedResultTableData(state: ResultsState): boolean {
    return !!state.excludedResultTableData;
  },

  getExcludedResultTableData(state: ResultsState): TableData {
    return state.excludedResultTableData;
  },

  getExcludedResultTableDataItems(
    state: ResultsState,
    getters: any
  ): TableRow[] {
    return getTableDataItems(state.excludedResultTableData);
  },

  getExcludedResultTableDataFields(
    state: ResultsState
  ): Dictionary<TableColumn> {
    return getTableDataFields(state.excludedResultTableData);
  },

  getExcludedResultTableDataCount(state: ResultsState): number {
    return state.excludedResultTableData.numRowsFiltered
      ? state.excludedResultTableData.numRowsFiltered
      : 0;
  },

  /* Check if any items have a weight property */
  hasResultTableDataItemsWeight(state: ResultsState): boolean {
    const data = getTableDataItems(state.includedResultTableData) ?? [];
    return data.some((item) =>
      Object.keys(item).some((variable) =>
        item[variable].hasOwnProperty("weight")
      )
    );
  },

  // predicted

  getPredictedSummaries(state: ResultsState): VariableSummary[] {
    return state.predictedSummaries;
  },

  // residual

  getResidualsSummaries(state: ResultsState): VariableSummary[] {
    return state.residualSummaries;
  },

  getResidualsExtrema(state: ResultsState): Extrema {
    return state.residualsExtrema;
  },

  // correctness

  getCorrectnessSummaries(state: ResultsState): VariableSummary[] {
    return state.correctnessSummaries;
  },

  // forecasts

  getPredictedTimeseries(state: ResultsState): Dictionary<TimeSeries> {
    return state.timeseries;
  },

  getPredictedForecasts(state: ResultsState): Dictionary<Forecast> {
    return state.forecasts;
  },

  // variable rankings

  getVariableRankings(state: ResultsState): Dictionary<Dictionary<number>> {
    return state.variableRankings;
  },
};
