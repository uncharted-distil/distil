import _ from 'lodash';
import { VariableSummary, Extrema, TableData, TableRow, TableColumn } from '../dataset/index';
import { ResultsState } from './index';
import { getTableDataItems, getTableDataFields } from '../../util/data';
import { Dictionary } from '../../util/dict';

export const getters = {

	// results

	getResultSummaries(state: ResultsState): VariableSummary[] {
		return state.resultSummaries;
	},

	getResultDataNumRows(state: ResultsState): number {
		return state.includedResultTableData ? state.includedResultTableData.numRows : 0;
	},

	hasIncludedResultTableData(state: ResultsState): boolean {
		return !!state.includedResultTableData;
	},

	getIncludedResultTableData(state: ResultsState): TableData {
		return state.includedResultTableData;
	},

	getIncludedResultTableDataItems(state: ResultsState, getters: any): TableRow[] {
		return getTableDataItems(state.includedResultTableData);
	},

	getIncludedResultTableDataFields(state: ResultsState): Dictionary<TableColumn> {
		return getTableDataFields(state.includedResultTableData);
	},

	hasExcludedResultTableData(state: ResultsState): boolean {
		return !!state.excludedResultTableData;
	},

	getExcludedResultTableData(state: ResultsState): TableData {
		return state.excludedResultTableData;
	},

	getExcludedResultTableDataItems(state: ResultsState, getters: any): TableRow[] {
		return getTableDataItems(state.excludedResultTableData);
	},

	getExcludedResultTableDataFields(state: ResultsState): Dictionary<TableColumn> {
		return getTableDataFields(state.excludedResultTableData);
	},

	// predicted

	getPredictedSummaries(state: ResultsState): VariableSummary[] {
		return state.predictedSummaries;
	},

	getPredictedExtrema(state: ResultsState): Extrema {
		if (_.isEmpty(state.predictedExtremas) && !state.targetResultExtrema) {
			return {
				min: null,
				max: null
			};
		}
		const res = { min: Infinity, max: -Infinity };
		_.forIn(state.predictedExtremas, extrema => {
			res.min = Math.min(res.min, extrema.min);
			res.max = Math.max(res.max, extrema.max);
		});
		if (state.targetResultExtrema) {
			res.min = Math.min(res.min, state.targetResultExtrema.min);
			res.max = Math.max(res.max, state.targetResultExtrema.max);
		}
		return res;
	},

	// residual

	getResidualsSummaries(state: ResultsState): VariableSummary[] {
		return state.residualSummaries;
	},

	getResidualExtrema(state: ResultsState): Extrema {
		if (_.isEmpty(state.residualExtremas)) {
			return {
				min: null,
				max: null
			};
		}
		const res = { min: Infinity, max: -Infinity };
		_.forIn(state.residualExtremas, extrema => {
			res.min = Math.min(res.min, extrema.min);
			res.max = Math.max(res.max, extrema.max);
		});
		return res;
	},

	// correctness

	getCorrectnessSummaries(state: ResultsState): VariableSummary[] {
		return state.correctnessSummaries;
	}
}
