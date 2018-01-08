import { Location } from 'vue-router';
import { ViewState } from './index';
import { Dictionary } from '../../util/dict';

export const getters = {
	getPrevView(state: ViewState): Dictionary<Dictionary<Location>> {
		return state.stack;
	}
}
