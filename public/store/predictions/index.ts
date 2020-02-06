import { Dictionary } from "../../util/dict";
import {
  VariableSummary,
  Extrema,
  TableData,
  TimeSeries
} from "../dataset/index";

export interface Forecast {
  forecastData: Dictionary<number[][]>;
  forecastRange: Dictionary<number[]>;
  isDateTime: Dictionary<boolean>;
}

export interface PredictionState {
  // table data
  includedPredictionTableData: TableData;
  excludedPredictionTableData: TableData;
  // training / target
  trainingSummaries: VariableSummary[];
  targetSummary: VariableSummary;
  // predicted
  predictedSummaries: VariableSummary[];
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
  // training / target
  trainingSummaries: [],
  targetSummary: null,
  // predicted
  predictedSummaries: [],
  // forecasts
  timeseries: {},
  forecasts: {},
  fittedSolutionId: null,
  produceRequestId: null
};
