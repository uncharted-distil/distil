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
    return state.includedPredictionTableData?.numRows ?? 0;
  },

  getFittedSolutionIdFromPrediction(state: PredictionState): string {
    return state.includedPredictionTableData?.fittedSolutionId;
  },

  getProduceRequestIdFromPrediction(state: PredictionState): string {
    return state.includedPredictionTableData?.produceRequestId;
  },

  hasIncludedPredictionTableData(state: PredictionState): boolean {
    return !!state.includedPredictionTableData;
  },

  getIncludedPredictionTableData(state: PredictionState): TableData {
    return state.includedPredictionTableData;
  },
  getAreaOfInterestInnerDataItems(state: PredictionState): TableRow[] {
    return getTableDataItems(state.areaOfInterestInner);
  },
  getAreaOfInterestOuterDataItems(state: PredictionState): TableRow[] {
    return getTableDataItems(state.areaOfInterestOuter);
  },
  getBaselinePredictionTableDataItems(
    state: PredictionState,
    getters: any
  ): TableRow[] {
    return getTableDataItems(state.baselinePredictionTableData);
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

  getTrainingSummariesDictionary(
    state: PredictionState
  ): Dictionary<Dictionary<VariableSummary>> {
    return state.trainingSummaries;
  },

  // forecasts

  getPredictionTimeseries(state: PredictionState): Dictionary<TimeSeries> {
    return state.timeseries;
  },

  getPredictionForecasts(state: PredictionState): Dictionary<Forecast> {
    return state.forecasts;
  },
};
