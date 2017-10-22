import Vue from 'vue';
import Vuex from 'vuex';
import { Store, StoreOptions } from 'vuex';
import { state } from './index';
import {actions } from './actions';
import { getters } from './getters';
import { mutations } from './mutations';

Vue.use(Vuex);

export default new Store({
	state,
	getters,
	actions,
	mutations,
	strict: true
});
