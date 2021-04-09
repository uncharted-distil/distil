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
import {
  Extrema,
  TableData,
  TimeSeriesValue,
  VariableSummary,
} from "../dataset/index";

export interface TimeSeries {
  timeseriesData: Dictionary<TimeSeriesValue[]>;
  isDateTime: Dictionary<boolean>;
  info: Dictionary<Extrema>;
}

export interface Forecast {
  forecastData: Dictionary<TimeSeriesValue[]>;
  forecastRange: Dictionary<number[]>;
  isDateTime: Dictionary<boolean>;
}

export interface ResultsState {
  // table data
  includedResultTableData: TableData;
  excludedResultTableData: TableData;
  // baseline data without datasize applied (for the map)
  fullIncludedResultTableData: TableData;
  // areaOfInterest
  areaOfInterestInner: TableData;
  areaOfInterestOuter: TableData;
  // training / target
  trainingSummaries: Dictionary<Dictionary<VariableSummary>>;
  targetSummary: VariableSummary;
  // predicted
  predictedSummaries: VariableSummary[];
  // residuals
  residualSummaries: VariableSummary[];
  residualsExtrema: Extrema;
  // correctness summary (correct vs. incorrect) for predicted categorical data
  correctnessSummaries: VariableSummary[];
  // confidence summary (how sure the system is of it's predictions) for any predicted data
  confidenceSummaries: VariableSummary[];
  // ranking summary
  rankingSummaries: VariableSummary[];
  // timeseries by solutionID, timeseriesID
  timeseries: Dictionary<TimeSeries>;
  // forecasts by solution ID
  forecasts: Dictionary<Forecast>;
  // result IDs
  fittedSolutionId: string;
  produceRequestId: string;
  // variable rankings - maps {solutionID, {featureName: rank value}}
  featureImportanceRanking: Dictionary<Dictionary<number>>;
}

export const state: ResultsState = {
  // table data
  includedResultTableData: null,
  excludedResultTableData: null,
  fullIncludedResultTableData: null,
  // area of interest for tiles clicks in geo map
  areaOfInterestInner: null,
  areaOfInterestOuter: null,
  // training / target
  trainingSummaries: {},
  targetSummary: null,
  // predicted
  predictedSummaries: [],
  // residuals
  residualSummaries: [],
  residualsExtrema: { min: null, max: null },
  // correctness summary (correct vs. incorrect) for predicted categorical data
  correctnessSummaries: [],
  // confidence summary (how sure the system is of it's predictions) for any predicted data
  confidenceSummaries: [],
  // ranking summary (used in tandem with confidence)
  rankingSummaries: [],
  // forecasts
  timeseries: {},
  forecasts: {},
  // result IDs
  fittedSolutionId: null,
  produceRequestId: null,
  // variable rankings
  featureImportanceRanking: {},
};
