import { Location } from 'vue-router';
import { Store } from 'vuex';
import { LAST_STATE } from '../store/view/index';
import { getters as viewGetters } from '../store/view/module';

export function restoreView(store: Store<any>, view: string, dataset: string): Location {
	const prev = viewGetters.getPrevView(store);
	if (!prev[view]) {
		return null;
	}
	if (!dataset) {
		return prev[view][LAST_STATE];
	}
	if (!prev[view][dataset]) {
		return null;
	}
	return prev[view][dataset];
}
