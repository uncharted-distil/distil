import { Dictionary } from '../../util/dict';

export enum SummaryType {
	Categorical = "categorical",
	Numerical = "numerical"
}

export interface Variable {
	name: string;
	type: string;
	importance: number;
	novelty: number;
}

export interface Datasets {
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
	name: string;
	label: string;
	feature: string;
	dataset: string;
	buckets: Bucket[];
	extrema: Extrema;
	numRows: number;
	solutionId?: string;
	resultId?: string;
	type?: string;
	varType?: string;
	err?: string;
	pending?: boolean;
}

export interface Data {
	name: string;
	numRows: number;
	columns: string[];
	types: string[];
	values: any[][];
}

export interface FieldInfo {
	label: string,
	type: string,
	sortable: boolean
}

export interface TableRow {
	_key: number;
	_rowVariant: string;
}

export interface TargetRow extends TableRow {
	_cellVariants: Dictionary<string>;
}

export interface HighlightRoot {
	context: string;
	key: string;
	value: any;
}

export interface Column {
	key: string;
	value: any;
}

export interface RowSelection {
	context: string;
	index: number;
	cols: Column[];
}

export interface HighlightValues {
	summaries?: VariableSummary[];
	samples?: Dictionary<string[]>;
}

export interface Highlight {
	root: HighlightRoot;
	values: HighlightValues;
}

export interface DataState {
	datasets: Datasets[];
	variables: Variable[];
	variableSummaries: VariableSummary[];
	resultSummaries: VariableSummary[];
	predictedSummaries: VariableSummary[];
	residualSummaries: VariableSummary[];
	correctnessSummaries: VariableSummary[];
	resultExtrema: Extrema;
	predictedExtremas: Dictionary<Extrema>;
	residualExtremas: Dictionary<Extrema>;
	highlightedResultData: Data;
	unhighlightedResultData: Data;
	selectedData: Data;
	excludedData: Data;
	highlightValues: HighlightValues;
	loadedImages: {};
}

export const state = {
	// description of matched datasets
	datasets: <Datasets[]>[],

	// variable list for the active dataset
	variables: <Variable[]>[],

	// variable summary data for the active dataset
	variableSummaries: <VariableSummary[]>[],

	// results summary data for the training dataset
	resultSummaries: <VariableSummary[]>[],

	// results summary data for the predicted data
	predictedSummaries: <VariableSummary[]>[],

	// residuals summary for the predicted numerical data
	residualSummaries: <VariableSummary[]>[],

	// correctness summary (correct vs. incorrect) for predicted categorical data
	correctnessSummaries: <VariableSummary[]>[],

	resultExtrema: null,

	predictedExtremas: {},

	residualExtremas: {},

	// current set of solution results
	highlightedResultData: null,

	unhighlightedResultData: null,

	// selected data entries for the active dataset
	selectedData: null,

	// excluded data entries for the active dataset
	excludedData: null,

	// highlight values
	highlightValues: {
		summaries: [],
		samples: {}
	},

	loadedImages: {}
}
