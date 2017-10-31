import { Module } from 'vuex';
import { state, AppState } from './index';
import { getters } from './getters';
import { actions } from './actions';
import { mutations } from './mutations';

export const appModule: Module<AppState, any> = {
	state: state,
	getters: getters,
	actions: actions,
	mutations: mutations
}
