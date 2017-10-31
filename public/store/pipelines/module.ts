import { Module } from 'vuex';
import { state, PipelineState } from './index';
import { getters } from './getters';
import { actions } from './actions';
import { mutations } from './mutations';

export const pipelineModule: Module<PipelineState, any> = {
	state: state,
	getters: getters,
	actions: actions,
	mutations: mutations
}
