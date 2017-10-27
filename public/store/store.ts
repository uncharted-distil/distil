import Vue from 'vue';
import Vuex from 'vuex';
import { Store } from 'vuex';
import { state, DistilState } from './index';
import {actions } from './actions';
import { getters } from './getters';
import { mutations } from './mutations';

Vue.use(Vuex);

export default new Store<DistilState>({
	state,
	getters,
	actions,
	mutations,
	strict: true
});
