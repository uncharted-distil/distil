import { Dictionary } from '../../util/dict';
import { VariableSummary, Extrema, TableData } from '../dataset/index';

export interface ResultsState {
	// result
	resultSummaries: VariableSummary[];
	targetResultExtrema: Extrema;
	includedResultTableData: TableData;
	excludedResultTableData: TableData;
	// predicted
	predictedSummaries: VariableSummary[];
	predictedExtremas: Dictionary<Extrema>;
	// residuals
	residualSummaries: VariableSummary[];
	residualExtremas: Dictionary<Extrema>;
	// correctness summary (correct vs. incorrect) for predicted categorical data
	correctnessSummaries: VariableSummary[];
}

export const state: ResultsState = {
	// result
	resultSummaries: <VariableSummary[]>[],
	targetResultExtrema: null,
	includedResultTableData: null,
	excludedResultTableData: null,
	// predicted
	predictedSummaries: <VariableSummary[]>[],
	predictedExtremas: {},
	// residuals
	residualSummaries: <VariableSummary[]>[],
	residualExtremas: {},
	// correctness summary (correct vs. incorrect) for predicted categorical data
	correctnessSummaries: <VariableSummary[]>[],
}
