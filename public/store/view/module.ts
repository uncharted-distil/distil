import { Module } from 'vuex';
import { state, ViewState } from './index';
import { getters as moduleGetters } from './getters';
import { mutations as moduleMutations } from './mutations';
import { DistilState } from '../store';
import { getStoreAccessors } from 'vuex-typescript';

export const viewModule: Module<ViewState, DistilState> = {
	state: state,
	getters: moduleGetters,
	mutations: moduleMutations
};

const { commit, read } = getStoreAccessors<ViewState, DistilState>(null);

export const getters = {
	getPrevView: read(moduleGetters.getPrevView)
};

export const mutations = {
	saveView: commit(moduleMutations.saveView)
};
