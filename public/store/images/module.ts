import { Module } from 'vuex';
import { state, ImageState } from './index';
import { getters as moduleGetters } from './getters';
import { actions as moduleActions } from './actions';
import { mutations as moduleMutations } from './mutations';
import { DistilState } from '../store';
import { getStoreAccessors } from 'vuex-typescript';

export const imagesModule: Module<ImageState, DistilState> = {
	getters: moduleGetters,
	actions: moduleActions,
	mutations: moduleMutations,
	state: state
}

const { commit, read, dispatch } = getStoreAccessors<ImageState, DistilState>(null);

// Typed getters
export const getters = {
	getImages: read(moduleGetters.getImages),
}

// Typed actions
export const actions = {
	fetchImage: dispatch(moduleActions.fetchImage)
}

// Typed mutations
export const mutations = {
	setImage: commit(moduleMutations.setImage),

}
