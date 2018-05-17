import { VariableSummary } from '../dataset/index';
import { Dictionary } from '../../util/dict';

export interface HighlightRoot {
	context: string;
	key: string;
	value: any;
}

export interface HighlightValues {
	summaries?: VariableSummary[];
	samples?: Dictionary<string[]>;
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

export interface RowSelection {
	context: string;
	index: number;
	cols: Column[];
}

export const state: HighlightState = {
	// highlighted values fetched from the server
	highlightValues: {
		summaries: [],
		samples: {}
	}
}
