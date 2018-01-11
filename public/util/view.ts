import { Location } from 'vue-router';
import { Store } from 'vuex';
import { LAST_STATE } from '../store/view/index';
import { getters as viewGetters } from '../store/view/module';
import localStorage from 'store';

export function restoreView(store: Store<any>, view: string, dataset: string): Location {
	const prev = viewGetters.getPrevView(store);
	const key = dataset || LAST_STATE;
	if (!prev[view])
		return localStorage.get(`${view}:${key}`) || null;

	if (!prev[view][key]) {
		return localStorage.get(`${view}:${key}`) || null;
	}
	return prev[view][key];
}
