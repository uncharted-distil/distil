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

import { Dictionary } from "../../util/dict";
import { VariableSummary, TableData, TimeSeries } from "../dataset/index";
import { Forecast } from "../results";

export interface PredictionState {
  // table data
  includedPredictionTableData: TableData;
  excludedPredictionTableData: TableData;
  // baseline tableData
  baselinePredictionTableData: TableData;
  // areaOfInterest
  areaOfInterestInner: TableData;
  areaOfInterestOuter: TableData;
  // training / target
  trainingSummaries: Dictionary<Dictionary<VariableSummary>>;
  targetSummary: VariableSummary;
  // predicted
  predictedSummaries: VariableSummary[];
  confidenceSummaries: VariableSummary[];
  rankSummaries: VariableSummary[];
  // forecasts
  timeseries: Dictionary<TimeSeries>;
  forecasts: Dictionary<Forecast>;
  fittedSolutionId: string;
  produceRequestId: string;
}

export const state: PredictionState = {
  // table data
  includedPredictionTableData: null,
  excludedPredictionTableData: null,
  // baseline table data
  baselinePredictionTableData: null,
  // training / target
  trainingSummaries: {},
  targetSummary: null,
  // predicted
  predictedSummaries: [],
  confidenceSummaries: [],
  rankSummaries: [],
  // forecasts
  timeseries: {},
  forecasts: {},
  // areaOfInterest
  areaOfInterestInner: null,
  areaOfInterestOuter: null,
  fittedSolutionId: null,
  produceRequestId: null,
};
