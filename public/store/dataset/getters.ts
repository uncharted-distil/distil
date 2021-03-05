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
  Row,
  Metric,
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

  //filters the hidden variables used in groupings
  getVariables(state: DatasetState, getters: any): Variable[] {
    const groupings = state.variables.reduce((a, v) => {
      const hiddenVars = v.grouping?.hidden as string[];
      if (hiddenVars) {
        a = a.concat(hiddenVars);
      }
      return a;
    }, []);
    return state.variables.filter((v) => groupings.indexOf(v.key) < 0);
  },

  //includes the hidden variables, only used for lex
  getAllVariables(state: DatasetState, getters: any): Variable[] {
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
      map[variable.key] = variable;
      map[variable.key.toLowerCase()] = variable;
    });
    return map;
  },

  getVariableTypesMap(state: DatasetState): Dictionary<string> {
    const map = {};
    state.variables.forEach((variable) => {
      map[variable.key] = variable.colType;
      map[variable.key.toLowerCase()] = variable.colType;
    });
    return map;
  },

  getVariableRankings(state: DatasetState): Dictionary<Dictionary<number>> {
    return state.variableRankings;
  },

  getVariableSummariesDictionary(
    state: DatasetState
  ): Dictionary<Dictionary<VariableSummary>> {
    return state.includedSet.variableSummariesByKey;
  },

  getIncludedVariableSummariesDictionary(
    state: DatasetState
  ): Dictionary<Dictionary<VariableSummary>> {
    return state.includedSet.variableSummariesByKey;
  },

  getExcludedVariableSummariesDictionary(
    state: DatasetState
  ): Dictionary<Dictionary<VariableSummary>> {
    return state.excludedSet.variableSummariesByKey;
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
  getHighlightedIncludeSet(state: DatasetState): TableData {
    return state.highlightedIncludeSet;
  },
  getHighlightedExcludeSet(state: DatasetState): TableData {
    return state.highlightedExcludeSet;
  },
  getIncludedTableData(state: DatasetState): TableData {
    return state.includedSet.tableData;
  },

  getIncludedTableDataLength(state: DatasetState): number {
    return state.includedSet.tableData?.values?.length;
  },

  getIncludedTableDataNumRows(state: DatasetState): number {
    return state.includedSet.tableData
      ? state.includedSet.tableData.numRowsFiltered
      : 0;
  },
  getNumberOfRecords(state: DatasetState): number {
    return state.includedSet.tableData
      ? state.includedSet.tableData.numRows
      : 0;
  },
  getAreaOfInterestIncludeInnerItems(state: DatasetState) {
    return getTableDataItems(state.areaOfInterestIncludeInner);
  },
  getAreaOfInterestIncludeOuterItems(state: DatasetState) {
    return getTableDataItems(state.areaOfInterestIncludeOuter);
  },
  getAreaOfInterestExcludeInnerItems(state: DatasetState) {
    return getTableDataItems(state.areaOfInterestExcludeInner);
  },
  getAreaOfInterestExcludeOuterItems(state: DatasetState) {
    return getTableDataItems(state.areaOfInterestExcludeOuter);
  },
  getHighlightedIncludeTableDataItems(state: DatasetState) {
    return getTableDataItems(state.highlightedIncludeSet);
  },
  getHighlightedExcludeTableDataItems(state: DatasetState) {
    return getTableDataItems(state.highlightedExcludeSet);
  },
  getIncludedTableDataItems(state: DatasetState, getters: any): TableRow[] {
    return getTableDataItems(state.includedSet.tableData);
  },

  getIncludedTableDataFields(state: DatasetState): Dictionary<TableColumn> {
    return getTableDataFields(state.includedSet.tableData);
  },

  getIncludedSelectedRowData(state: DatasetState): Row[] {
    return state.includedSet.rowSelectionData;
  },

  hasExcludedTableData(state: DatasetState): boolean {
    return !!state.excludedSet.tableData;
  },

  getExcludedTableData(state: DatasetState): TableData {
    return state.excludedSet.tableData;
  },

  getExcludedTableDataLength(state: DatasetState): number {
    return state.excludedSet.tableData?.values?.length;
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

  getExcludedSelectedRowData(state: DatasetState): Row[] {
    return state.excludedSet.rowSelectionData;
  },

  getMultiBandCombinations(state: DatasetState): BandCombination[] {
    return state.bands;
  },

  getModelingMetrics(state: DatasetState): Metric[] {
    return state.metrics;
  },
};
