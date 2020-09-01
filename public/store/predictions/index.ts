import { Dictionary } from "../../util/dict";
import { VariableSummary, TableData, TimeSeries } from "../dataset/index";
import { Forecast } from "../results";

export interface WorkingSet {
  variableSummaries: VariableSummary[];
  variableSummariesByKey: Dictionary<Dictionary<VariableSummary>>;
}

export interface PredictionState {
  // table data
  includedPredictionTableData: TableData;
  excludedPredictionTableData: TableData;
  // training / target
  trainingSummaries: WorkingSet;
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
  trainingSummaries: {
    variableSummaries: [],
    variableSummariesByKey: {},
  },
  targetSummary: null,
  // predicted
  predictedSummaries: [],
  // forecasts
  timeseries: {},
  forecasts: {},
  fittedSolutionId: null,
  produceRequestId: null,
};
