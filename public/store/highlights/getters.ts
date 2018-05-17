import { HighlightState } from './index';
import { VariableSummary } from '../dataset/index';
import { Dictionary } from '../../util/dict';

export const getters = {

	getHighlightedSamples(state: HighlightState): Dictionary<string[]> {
		return state.highlightValues ? state.highlightValues.samples : {};
	},

	getHighlightedSummaries(state: HighlightState): VariableSummary[] {
		return state.highlightValues ? state.highlightValues.summaries : null;
	}
}
