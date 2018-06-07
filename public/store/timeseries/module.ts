import { Module } from 'vuex';
import { state, TimeSeriesState } from './index';
import { getters as moduleGetters } from './getters';
import { actions as moduleActions } from './actions';
import { mutations as moduleMutations } from './mutations';
import { DistilState } from '../store';
import { getStoreAccessors } from 'vuex-typescript';

export const timeSeriesModule: Module<TimeSeriesState, DistilState> = {
	getters: moduleGetters,
	actions: moduleActions,
	mutations: moduleMutations,
	state: state
}

const { commit, read, dispatch } = getStoreAccessors<TimeSeriesState, DistilState>(null);

// Typed getters
export const getters = {
	getTimeSeries: read(moduleGetters.getTimeSeries),
}

// Typed actions
export const actions = {
	fetchTimeSeries: dispatch(moduleActions.fetchTimeSeries)
}

// Typed mutations
export const mutations = {
	setTimeSeries: commit(moduleMutations.setTimeSeries),

}
