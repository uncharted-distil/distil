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
  getTableDataFields,
  getTableDataItems,
  minimumRouteKey,
} from "../../util/data";
import { Dictionary } from "../../util/dict";
import {
  Extrema,
  TableColumn,
  TableData,
  TableRow,
  VariableSummary,
} from "../dataset/index";
import { Forecast, ResultsState, TimeSeries } from "./index";

export const getters = {
  // results

  getTrainingSummaries(state: ResultsState): VariableSummary[] {
    const minKey = minimumRouteKey();
    const summaries = getters.getTrainingSummariesDictionary;
    const variableNames = Object.keys(summaries);
    const trainingVariableSummaries = variableNames.reduce(
      (acc, variableName) => {
        const variableSummary = summaries?.[variableName]?.[minKey];
        if (variableSummary) {
          acc.push(variableSummary);
        }
        return acc;
      },
      []
    );
    return trainingVariableSummaries;
  },

  getTrainingSummariesDictionary(
    state: ResultsState
  ): Dictionary<Dictionary<VariableSummary>> {
    return state.trainingSummaries;
  },

  getTargetSummary(state: ResultsState): VariableSummary {
    return state.targetSummary;
  },

  getResultDataNumRows(state: ResultsState): number {
    return state.includedResultTableData?.numRows ?? 0;
  },

  getFittedSolutionId(state: ResultsState): string {
    return state.includedResultTableData?.fittedSolutionId;
  },

  getProduceRequestId(state: ResultsState): string {
    return state.includedResultTableData?.produceRequestId;
  },

  hasIncludedResultTableData(state: ResultsState): boolean {
    return !!state.includedResultTableData;
  },

  getIncludedResultTableData(state: ResultsState): TableData {
    return state.includedResultTableData;
  },
  getFullIncludedResultTableDataItems(state: ResultsState): TableRow[] {
    return getTableDataItems(state.fullIncludedResultTableData);
  },
  getFullExcludedResultTableDataItems(state: ResultsState): TableRow[] {
    return getTableDataItems(state.fullExcludedResultTableData);
  },
  getNumOfRecords(state: ResultsState) {
    const table = state.includedResultTableData ??
      state.excludedResultTableData ?? { numRows: 0 };
    return table.numRows;
  },
  getAreaOfInterestInnerDataItems(state: ResultsState): TableRow[] {
    return getTableDataItems(state.areaOfInterestInner);
  },
  getAreaOfInterestOuterDataItems(state: ResultsState): TableRow[] {
    return getTableDataItems(state.areaOfInterestOuter);
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
    return state.includedResultTableData?.numRowsFiltered ?? 0;
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
    return state.excludedResultTableData?.numRowsFiltered ?? 0;
  },

  /* Check if any items have a weight property */
  hasResultTableDataItemsWeight(state: ResultsState): boolean {
    const data = getTableDataItems(state.includedResultTableData) ?? [];
    return data.some((item) =>
      Object.keys(item).some(
        (variable) =>
          item[variable] &&
          typeof item[variable] === "object" &&
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
  // ranking
  getRankingSummaries(state: ResultsState): VariableSummary[] {
    return state.rankingSummaries;
  },
  // confidence

  getConfidenceSummaries(state: ResultsState): VariableSummary[] {
    return state.confidenceSummaries;
  },

  // forecasts

  getPredictedTimeseries(state: ResultsState): Dictionary<TimeSeries> {
    return state.timeseries;
  },

  getPredictedForecasts(state: ResultsState): Dictionary<Forecast> {
    return state.forecasts;
  },

  // variable rankings

  getFeatureImportanceRanking(
    state: ResultsState
  ): Dictionary<Dictionary<number>> {
    return state.featureImportanceRanking;
  },
};
