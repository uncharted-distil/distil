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
  // areaOfInteres
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
  // forecasts
  timeseries: {},
  forecasts: {},
  // result IDs
  fittedSolutionId: null,
  produceRequestId: null,
  // variable rankings
  featureImportanceRanking: {},
};
