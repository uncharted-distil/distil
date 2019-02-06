import { VariableSummary } from '../dataset/index';

export interface HighlightRoot {
	context: string;
	dataset: string;
	key: string;
	value: any;
}

export interface HighlightValues {
	summaries?: VariableSummary[];
}

export interface Highlight {
	root: HighlightRoot;
	values: HighlightValues;
}

export interface HighlightState {
	highlightValues: HighlightValues;
}

export interface Column {
	key: string;
	value: any;
}

export interface Row {
	index: number;
	d3mIndex: number;
	cols: Column[];
	included: boolean;
}

export interface RowSelection {
	context: string;
	d3mIndices: number[];
}

export const state: HighlightState = {
	// highlighted values fetched from the server
	highlightValues: {
		summaries: []
	}
};
