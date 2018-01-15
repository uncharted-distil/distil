import Vue from 'vue';
import { Location } from 'vue-router';
import { ViewState, LAST_STATE } from './index';
import localStorage from 'store';

export const mutations = {

	saveView(state: ViewState, args: { view: string, dataset: string, route: Location }) {
		if (!state.stack[args.view]) {
			Vue.set(state.stack, args.view, {});
		}
		const value = {
			path: args.route.path,
			query: args.route.query
		};
		// store under dataset
		Vue.set(state.stack[args.view], args.dataset, value);
		localStorage.set(`${args.view}:${args.dataset}`, value);
		// store last as well in case no dataset available
		Vue.set(state.stack[args.view], LAST_STATE, value);
		localStorage.set(`${args.view}:${LAST_STATE}`, value);
	}
}
