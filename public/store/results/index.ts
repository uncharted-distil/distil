import { Dictionary } from "../../util/dict";
import {
  VariableSummary,
  Extrema,
  TableData,
  TimeSeriesValue
} from "../dataset/index";

export interface TimeSeries {
  timeseriesData: Dictionary<TimeSeriesValue[]>;
  isDateTime: Dictionary<boolean>;
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
  // training / target
  trainingSummaries: VariableSummary[];
  targetSummary: VariableSummary;
  // predicted
  predictedSummaries: VariableSummary[];
  // residuals
  residualSummaries: VariableSummary[];
  residualsExtrema: Extrema;
  // correctness summary (correct vs. incorrect) for predicted categorical data
  correctnessSummaries: VariableSummary[];
  // timeseries by solutionID, timeseriesID
  timeseries: Dictionary<TimeSeries>;
  // forecasts by solution ID
  forecasts: Dictionary<Forecast>;
  // result IDs
  fittedSolutionId: string;
  produceRequestId: string;
}

export const state: ResultsState = {
  // table data
  includedResultTableData: null,
  excludedResultTableData: null,
  // training / target
  trainingSummaries: [],
  targetSummary: null,
  // predicted
  predictedSummaries: [],
  // residuals
  residualSummaries: [],
  residualsExtrema: { min: null, max: null },
  // correctness summary (correct vs. incorrect) for predicted categorical data
  correctnessSummaries: [],
  // forecasts
  timeseries: {},
  forecasts: {},
  // result IDs
  fittedSolutionId: null,
  produceRequestId: null
};
