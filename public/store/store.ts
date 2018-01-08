import Vue from 'vue';
import Vuex from 'vuex';
import { Store } from 'vuex';
import { routeModule } from './route/module';
import { Route } from 'vue-router';
import { dataModule } from './data/module';
import { DataState } from './data/index';
import { pipelineModule } from './pipelines/module';
import { PipelineState } from './pipelines/index';
import { viewModule } from './view/module';
import { ViewState } from './view/index';
import { appModule } from './app/module';
import { AppState } from './app/index';

Vue.use(Vuex);

export interface DistilState {
	routeModule: Route;
	dataModule: DataState;
	pipelineModule: PipelineState;
	viewModule: ViewState;
	appModule: AppState;
}

export default new Store<DistilState>({
	modules:  {
		routeModule,
		dataModule,
		pipelineModule,
		viewModule,
		appModule
	},
	strict: true
});
