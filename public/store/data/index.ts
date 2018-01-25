import { Dictionary } from '../../util/dict';

export interface Variable {
	name: string;
	type: string;
	importance: number;
	novelty: number;
}

export interface Datasets {
	name: string;
	description: string;
	variables: Variable[];
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
	feature: string;
	buckets: Bucket[];
	extrema: Extrema;
	pipelineId?: string;
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

export interface Highlights {
	root: HighlightRoot;
	values: Dictionary<string[]>;
}

export interface Range {
	to: number;
	from: number;
}

export interface HighlightRoot {
	context: string;
	key: string;
	value: string | Range;
}

export interface DataState {
	datasets: Datasets[];
	variables: Variable[];
	variableSummaries: VariableSummary[];
	resultsSummaries: VariableSummary[];
	residualSummaries: VariableSummary[];
	resultData: Data;
	filteredData: Data;
	selectedData: Data;
	highlightedFeatureValues: Highlights;
}

export const state = {
	// description of matched datasets
	datasets: <Datasets[]>[],

	// variable list for the active dataset
	variables: <Variable[]>[],

	// variable summary data for the active dataset
	variableSummaries: <VariableSummary[]>[],

	// results summary data for the selected pipeline run
	resultsSummaries: <VariableSummary[]>[],

	// error summary data for the selected pipeline run
	residualSummaries: <VariableSummary[]>[],

	// current set of pipeline results
	resultData: <Data>{},

	// filtered data entries for the active dataset
	filteredData: <Data>{},

	// selected data entries for the active dataset
	selectedData: <Data>{},

	// highlighted features
	highlightedFeatureValues: <Highlights>{}
}
