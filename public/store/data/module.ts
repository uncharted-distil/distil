import { Module } from 'vuex';
import { state, DataState } from './index';
import { getters } from './getters';
import { actions } from './actions';
import { mutations } from './mutations';

export const dataModule: Module<DataState, any> = {
	getters: getters,
	actions: actions,
	mutations: mutations,
	state: state
}
