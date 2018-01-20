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
}

export interface RangeHighlights extends Highlights {
	ranges: Range;
}

export interface HighlightRoot {
	context: string;
	key: string;
	value: string;
}

export interface ValueHighlights extends Highlights {
	values: Dictionary<string[]>;
}

export type Range = Dictionary<{
	from: number, to: number
}>;

export interface DataState {
	datasets: Datasets[];
	variables: Variable[];
	variableSummaries: VariableSummary[];
	resultsSummaries: VariableSummary[];
	residualSummaries: VariableSummary[];
	resultData: Data;
	filteredData: Data;
	selectedData: Data;
	highlightedFeatureRanges: RangeHighlights;
	highlightedFeatureValues: ValueHighlights;
}

export const state = {
	// description of matched datasets
	datasets: [],

	// variable list for the active dataset
	variables: [],

	// variable summary data for the active dataset
	variableSummaries: [],

	// results summary data for the selected pipeline run
	resultsSummaries: [],

	// error summary data for the selected pipeline run
	residualSummaries: [],

	// current set of pipeline results
	resultData: {} as any,

	// filtered data entries for the active dataset
	filteredData: {} as any,

	// selected data entries for the active dataset
	selectedData: {} as any,

	// highlighted features
	highlightedFeatureRanges: {} as any,

	highlightedFeatureValues: {} as any
}
