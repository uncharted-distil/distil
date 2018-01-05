import { Location } from 'vue-router';
import { ViewState } from './index';
import { Dictionary } from '../../util/dict';

export const getters = {
	getViewStack(state: ViewState): Dictionary<Dictionary<Location>> {
		return state.stack;
	}
}
