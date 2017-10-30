import Vue from 'vue';
import Vuex from 'vuex';
import { Store } from 'vuex';
import { state, DistilState } from './index';
import { actions } from './actions';
import { getters } from './getters';
import { mutations } from './mutations';
import { routeModule } from './route/module';
import { dataModule } from './data/module';

Vue.use(Vuex);

export default new Store<DistilState>({
	state,
	getters,
	actions,
	mutations,
	modules:  {
		routeModule,
		dataModule
	},
	strict: true
});
