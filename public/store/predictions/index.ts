import { Dictionary } from "../../util/dict";
import { VariableSummary, TableData, TimeSeries } from "../dataset/index";
import { Forecast } from "../results";

export interface PredictionState {
  // table data
  includedPredictionTableData: TableData;
  excludedPredictionTableData: TableData;
  // areaOfInterest
  areaOfInterestInner: TableData;
  areaOfInterestOuter: TableData;
  // training / target
  trainingSummaries: Dictionary<Dictionary<VariableSummary>>;
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
  trainingSummaries: {},
  targetSummary: null,
  // predicted
  predictedSummaries: [],
  // forecasts
  timeseries: {},
  forecasts: {},
  // areaOfInterest
  areaOfInterestInner: null,
  areaOfInterestOuter: null,
  fittedSolutionId: null,
  produceRequestId: null,
};
