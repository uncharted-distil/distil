import {
  Variable,
  TimeseriesExtrema,
  DatasetState,
  Dataset,
  VariableSummary,
  TableData,
  TableRow,
  TableColumn,
  TimeSeries,
  BandCombination,
} from "./index";
import { Dictionary } from "../../util/dict";
import { getTableDataItems, getTableDataFields } from "../../util/data";
import { isInteger, values } from "lodash";

export const getters = {
  getDatasets(state: DatasetState): Dataset[] {
    return state.datasets;
  },

  getFilteredDatasets(state: DatasetState): Dataset[] {
    return state.filteredDatasets;
  },

  getCountOfFilteredDatasets(state: DatasetState): number {
    const count = values(state.filteredDatasets).length;
    return isInteger(count) ? count : 0;
  },

  getVariables(state: DatasetState, getters: any): Variable[] {
    return state.variables;
  },

  getGroupings(state: DatasetState, getters: any): Variable[] {
    return state.variables.filter((v) => v.grouping);
  },

  /**
   * Return the varibles used on the timeseries grouping.
   * @return {Array<String>}
   */
  getTimeseriesGroupingVariables(state: DatasetState): string[] {
    // Get only the timeseries grouping.
    const timeseriesGrouping = state.variables.find(
      (v) => v.grouping && v.grouping.type === "timeseries"
    );

    // Return an empty array if none have been found.
    if (!timeseriesGrouping) {
      return [];
    }

    return timeseriesGrouping.grouping.subIds;
  },

  getPendingRequests(state: DatasetState) {
    return state.pendingRequests;
  },

  getVariablesMap(state: DatasetState): Dictionary<Variable> {
    const map = {};
    state.variables.forEach((variable) => {
      map[variable.colName] = variable;
      map[variable.colName.toLowerCase()] = variable;
    });
    return map;
  },

  getVariableTypesMap(state: DatasetState): Dictionary<string> {
    const map = {};
    state.variables.forEach((variable) => {
      map[variable.colName] = variable.colType;
      map[variable.colName.toLowerCase()] = variable.colType;
    });
    return map;
  },

  getVariableRankings(state: DatasetState): Dictionary<Dictionary<number>> {
    return state.variableRankings;
  },

  getVariableSummaries(state: DatasetState): VariableSummary[] {
    return state.includedSet.variableSummaries;
  },

  getIncludedVariableSummaries(state: DatasetState): VariableSummary[] {
    return state.includedSet.variableSummaries;
  },

  getExcludedVariableSummaries(state: DatasetState): VariableSummary[] {
    return state.excludedSet.variableSummaries;
  },

  getFiles(state: DatasetState): Dictionary<any> {
    return state.files;
  },

  getTimeseriesExtrema(state: DatasetState): Dictionary<TimeseriesExtrema> {
    return state.timeseriesExtrema;
  },

  getTimeseries(state: DatasetState): Dictionary<TimeSeries> {
    return state.timeseries;
  },

  getJoinDatasetsTableData(state: DatasetState): Dictionary<TableData> {
    return state.joinTableData;
  },

  hasIncludedTableData(state: DatasetState): boolean {
    return !!state.includedSet.tableData;
  },

  getIncludedTableData(state: DatasetState): TableData {
    return state.includedSet.tableData;
  },

  getIncludedTableDataNumRows(state: DatasetState): number {
    return state.includedSet.tableData
      ? state.includedSet.tableData.numRowsFiltered
      : 0;
  },

  getIncludedTableDataItems(state: DatasetState, getters: any): TableRow[] {
    return getTableDataItems(state.includedSet.tableData);
  },

  getIncludedTableDataFields(state: DatasetState): Dictionary<TableColumn> {
    return getTableDataFields(state.includedSet.tableData);
  },

  hasExcludedTableData(state: DatasetState): boolean {
    return !!state.excludedSet.tableData;
  },

  getExcludedTableData(state: DatasetState): TableData {
    return state.excludedSet.tableData;
  },

  getExcludedTableDataNumRows(state: DatasetState): number {
    return state.excludedSet.tableData
      ? state.excludedSet.tableData.numRowsFiltered
      : 0;
  },

  getExcludedTableDataItems(state: DatasetState, getters: any): TableRow[] {
    return getTableDataItems(state.excludedSet.tableData);
  },

  getExcludedTableDataFields(state: DatasetState): Dictionary<TableColumn> {
    return getTableDataFields(state.excludedSet.tableData);
  },

  getMultiBandCombinations(state: DatasetState): BandCombination[] {
    return state.bands;
  },
};
