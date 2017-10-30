export interface Variable {
	name: string;
	type: string;
	suggestedTypes: string;
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
}

export interface VariableSummary {
	name: string;
	feature: string;
	buckets: Bucket[];
	extrema: Extrema;
	type?: string;
	err?: string;
	pending?: string;
}

export interface Data {
	name: string;
	columns: string[];
	types: string[];
	values: any[][];
}

export interface Range {
	[name: string]: {
		from: number,
		to: number
	}
}

export type Dictionary<T> = { [key: string]: T }

export interface DataState {
	datasets: Datasets[];
	variables: Variable[];
	variableSummaries: VariableSummary[];
	resultsSummaries: VariableSummary[];
	resultData: Data;
	filteredData: Data;
	selectedData: Data;
	highlightedFeatureRanges: Range;
	highlightedFeatureValues: Dictionary<any>;
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
