import Vue from 'vue';
import { Location } from 'vue-router';
import { ViewState, LAST_STATE } from './index';
import localStorage from 'store';

export const mutations = {
	setFetchParamsCache(state: ViewState, args: { key: string, value: string }) {
		state.fetchParamsCache[args.key] = args.value;
	}
};
