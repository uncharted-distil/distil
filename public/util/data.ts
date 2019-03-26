import _ from 'lodash';
import axios from 'axios';
import Vue from 'vue';
import { Variable, VariableSummary, TimeseriesSummary, TableData, TableRow, TableColumn, Grouping, D3M_INDEX_FIELD } from '../store/dataset/index';
import { Solution, SOLUTION_COMPLETED } from '../store/solutions/index';
import { Dictionary } from './dict';
import { Group } from './facets';
import { FilterParams } from './filters';
import store from '../store/store';
import { getters as datasetGetters, actions as datasetActions } from '../store/dataset/module';
import { formatValue, TIMESERIES_TYPE } from '../util/types';

// Postfixes for special variable names
export const PREDICTED_SUFFIX = '_predicted';
export const ERROR_SUFFIX = '_error';

export const NUM_PER_PAGE = 10;

export const DATAMART_PROVENANCE_NYU = 'datamartNYU';
export const DATAMART_PROVENANCE_ISI = 'datamartISI';
export const ELASTIC_PROVENANCE = 'elastic';
export const FILE_PROVENANCE = 'file';


export function getTimeseriesGroupingsFromFields(variables: Variable[], fields: Dictionary<TableColumn>): Grouping[] {
	return _.map(fields, (field, key) => key)
		.filter(key => {
			const v = variables.find(v => v.colName === key);
			return (v && v.grouping && v.grouping.type === TIMESERIES_TYPE);
		}).map(key => {
			const v = variables.find(v => v.colName === key);
			return v.grouping;
		});
}

export function fetchHistogramExemplars(datasetName: string, variableName: string, histogram: VariableSummary) {

	const variables = datasetGetters.getVariables(store);
	const variable = variables.find(v => v.colName === variableName);

	if (histogram.exemplars) {
		if (variable.grouping) {
			if (variable.grouping.type === 'timeseries') {

				// if there a linked exemplars, fetch those before resolving
				return Promise.all(histogram.exemplars.map(exemplar => {
					return datasetActions.fetchTimeseries(store, {
						dataset: datasetName,
						timeseriesColName: variable.grouping.idCol,
						xColName: variable.grouping.properties.xCol,
						yColName: variable.grouping.properties.yCol,
						timeseriesID: exemplar,
					});
				}));
			}

		} else {
			// if there a linked files, fetch those before resolving
			return datasetActions.fetchFiles(store, {
				dataset: datasetName,
				variable: variableName,
				urls: histogram.exemplars
			});
		}
	}

	return new Promise(res => res());
}

export function updateSummaries(summary: VariableSummary|TimeseriesSummary, summaries: (VariableSummary|TimeseriesSummary)[]) {
	const index = _.findIndex(summaries, s => {
		return s.dataset === summary.dataset && s.key === summary.key;
	});
	if (index >= 0) {
		Vue.set(summaries, index, summary);
	} else {
		summaries.push(summary);
	}
}

export function filterSummariesByDataset(summaries: (VariableSummary|TimeseriesSummary)[], dataset: string): (VariableSummary|TimeseriesSummary)[] {
	return summaries.filter(summary => {
		return summary.dataset === dataset;
	});
}

export function createEmptyTableData(): TableData {
	return {
		numRows: 0,
		columns: [],
		values: []
	};
}

export function createPendingSummary(key: string, label: string, dataset: string, solutionId?: string): VariableSummary {
	return {
		key: key,
		label: label,
		dataset: dataset,
		feature: '',
		pending: true,
		buckets: [],
		extrema: {
			min: null,
			max: null
		},
		numRows: 0,
		solutionId: solutionId
	};
}

export function createErrorSummary(key: string, label: string, dataset: string, error: any): VariableSummary {
	return {
		key: key,
		label: label,
		dataset: dataset,
		feature: '',
		buckets: [],
		extrema: {
			min: null,
			max: null
		},
		err: error.response ? error.response.data : error,
		numRows: 0
	};
}

export function getSummary(
	context: any,
	endpoint: string,
	solution: Solution,
	key: string,
	label: string,
	updateFunction: (arg: any, summary: VariableSummary) => void,
	filterParams: FilterParams): Promise<any> {

	const feature = solution.feature;
	const dataset = solution.dataset;
	const solutionId = solution.solutionId;
	const resultId = solution.resultId;

	// save a placeholder histogram
	updateFunction(context, createPendingSummary(key, label, dataset, solutionId));

	// fetch the results for each solution
	if (solution.progress !== SOLUTION_COMPLETED) {
		// skip
		return;
	}

	// return promise
	return axios.post(`${endpoint}/${resultId}`, filterParams ? filterParams : {})
		.then(response => {
			// save the histogram data
			const histogram = response.data.histogram;
			histogram.feature = feature;
			histogram.solutionId = solutionId;
			histogram.resultId = resultId;
			histogram.dataset = dataset;
			updateFunction(context, histogram);
		})
		.catch(error => {
			console.error(error);
			updateFunction(context, createErrorSummary(key, label, dataset, error));
		});
}

export function filterVariablesByPage<T>(pageIndex: number, numPerPage: number, variables: T[]): T[] {
	if (variables.length > numPerPage) {
		const firstIndex = numPerPage * (pageIndex - 1);
		const lastIndex = Math.min(firstIndex + numPerPage, variables.length);
		return variables.slice(firstIndex, lastIndex);
	}
	return variables;
}

export function getVariableImportance(v: Variable): number {
	return v.ranking !== undefined ? v.ranking : v.importance;
}

export function sortVariablesByImportance(variables: Variable[]): Variable[] {
	variables.sort((a, b) => {
		return getVariableImportance(b) - getVariableImportance(a);
	});
	return variables;
}

export function sortGroupsByImportance(groups: Group[], variables: Variable[]): Group[] {
	// create importance lookup map
	const importance: Dictionary<number> = {};
	variables.forEach(variable => {
		importance[variable.colName] = getVariableImportance(variable);
	});
	// sort by importance
	groups.sort((a, b) => {
		return importance[b.key] - importance[a.key];
	});
	return groups;
}

export function validateData(data: TableData) {
	return !_.isEmpty(data) &&
		!_.isEmpty(data.values) &&
		!_.isEmpty(data.columns);
}

export function getTableDataItems(data: TableData): TableRow[] {
	if (validateData(data)) {
		// convert fetched result data rows into table data rows
		return data.values.map((resultRow, rowIndex) => {
			const row = {} as TableRow;
			resultRow.forEach((colValue, colIndex) => {
				const colName = data.columns[colIndex].key;
				const colType = data.columns[colIndex].type;
				row[colName] = formatValue(colValue, colType);
			});
			row._key = rowIndex;
			return row;
		});
	}
	return !_.isEmpty(data) ? [] : null;
}

export function getTableDataFields(data: TableData) {
	if (validateData(data)) {
		const result = {};
		for (const col of data.columns) {
			if (col.key !== D3M_INDEX_FIELD) {
				result[col.key] = {
					label: col.label,
					key: col.key,
					type: col.type,
					sortable: true
				};
			}
		}
		return result;
	}
	return {};
}

export function isDatamartProvenance(provenance: string): boolean {
	return provenance === DATAMART_PROVENANCE_NYU || provenance === DATAMART_PROVENANCE_ISI;
}
