import { HighlightState } from './index';
import { VariableSummary } from '../dataset/index';

export const getters = {

	getHighlightedSummaries(state: HighlightState): VariableSummary[] {
		return state.highlightValues ? state.highlightValues.summaries : null;
	}
};
