import Vue from 'vue';
import Vuex from 'vuex';
import { Store } from 'vuex';
import { routeModule } from './route/module';
import { dataModule } from './data/module';
import { pipelineModule } from './pipelines/module';
import { appModule } from './app/module';

Vue.use(Vuex);

export default new Store<any>({
	modules:  {
		routeModule,
		dataModule,
		pipelineModule,
		appModule
	},
	strict: true
});
