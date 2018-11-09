import { Dictionary } from '../../util/dict';

export enum SummaryType {
	Categorical = "categorical",
	Numerical = "numerical"
}

export const D3M_INDEX_FIELD = 'd3mIndex';

export interface SuggestedType {
	probability: number;
	provenance: string;
	type: string;
}

export interface Variable {
	colDisplayName: string;
	colName: string;
	type: string;
	importance: number;
	novelty: number;
	originalType: string;
	suggestedTypes: SuggestedType[];
}

export interface Dataset {
	name: string;
	description: string;
	summary: string;
	summaryML: string;
	variables: Variable[];
	numBytes: number;
	numRows: number;
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
}

export interface TableRow {
	_key: number;
	_rowVariant: string;
	_cellVariants: Dictionary<string>;
}

export interface DatasetState {
	datasets: Dataset[];
	variables: Variable[];
	variableSummaries: VariableSummary[];
	files: Dictionary<any>;
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

	// selected data entries for the active dataset
	includedTableData: null,

	// excluded data entries for the active dataset
	excludedTableData: null
}
