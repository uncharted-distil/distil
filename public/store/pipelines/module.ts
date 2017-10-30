import { Module } from 'vuex';
import { DistilState } from '../index';
import { state, PipelineState } from './index';
import { getters } from './getters';
import { actions } from './actions';
import { mutations } from './mutations';

export const pipelineModule: Module<PipelineState, DistilState> = {
	state: state,
	getters: getters,
	actions: actions,
	mutations: mutations
}
