import Vue from 'vue';
import { Location } from 'vue-router';
import { ViewState, LAST_STATE } from './index';

export const mutations = {

	pushRoute(state: ViewState, args: { view: string, dataset: string, route: Location }) {
		if (!state.stack[args.view]) {
			Vue.set(state.stack, args.view, {});
		}
		console.log('push state for view', args.view, 'under dataset', args.dataset, ':', args.route);
		Vue.set(state.stack[args.view], args.dataset, args.route);
		Vue.set(state.stack[args.view], LAST_STATE, args.route); // store last as well in case no dataset available
	}
}
