import { Dictionary } from '../../util/dict';

export const CATEGORICAL_SUMMARY = 'categorical';
export const NUMERICAL_SUMMARY = 'numerical';

export const D3M_INDEX_FIELD = 'd3mIndex';

export interface SuggestedType {
	probability: number;
	provenance: string;
	type: string;
}

export interface Variable {
	colDisplayName: string;
	colName: string;
	colType: string;
	importance: number;
	ranking?: number;
	novelty: number;
	colOriginalType: string;
	suggestedTypes: SuggestedType[];
}

export interface Dataset {
	id: string;
	name: string;
	description: string;
	summary: string;
	summaryML: string;
	variables: Variable[];
	numBytes: number;
	numRows: number;
	provenance: string;
	source: string;
}

export interface Extrema {
	min: number;
	max: number;
}

export interface Bucket {
	key: string;
	count: number;
	buckets?: Bucket[];
}

export interface VariableSummary {
	label: string;
	key: string;
	feature: string;
	dataset: string;
	buckets: Bucket[];
	extrema: Extrema;
	numRows: number;
	files?: string[];
	solutionId?: string;
	resultId?: string;
	type?: string;
	varType?: string;
	err?: string;
	pending?: boolean;
	stddev?: number;
	mean?: number;
}

export interface TableData {
	numRows: number;
	columns: TableColumn[];
	values: any[][];
}

export interface TableColumn {
	label: string;
	key: string;
	type: string;
	sortable?: boolean;
	variant?: string;
}

export interface TableRow {
	_key: number;
	_rowVariant: string;
	_cellVariants: Dictionary<string>;
}

export interface TimeseriesExtrema {
	x: Extrema;
	y: Extrema;
}

export interface DatasetState {
	datasets: Dataset[];
	variables: Variable[];
	variableSummaries: VariableSummary[];
	files: Dictionary<any>;
	timeseriesExtrema: Dictionary<TimeseriesExtrema>;
	joinTableData: Dictionary<TableData>;
	includedTableData: TableData;
	excludedTableData: TableData;
}

export const state: DatasetState = {
	// description of matched datasets
	datasets: [],

	// variable list for the active dataset
	variables: [],

	// variable summary data for the active dataset
	variableSummaries: [],

	// linked files
	files: {},

	timeseriesExtrema: {},

	// joined data table data
	joinTableData: {},

	// selected data entries for the active dataset
	includedTableData: null,

	// excluded data entries for the active dataset
	excludedTableData: null
};
