import Vue from 'vue';
import Vuex from 'vuex';
import { Store } from 'vuex';
import { routeModule } from './route/module';
import { Route } from 'vue-router';
import { dataModule } from './data/module';
import { DataState } from './data/index';
import { solutionModule } from './solutions/module';
import { SolutionState } from './solutions/index';
import { viewModule } from './view/module';
import { ViewState } from './view/index';
import { appModule } from './app/module';
import { AppState } from './app/index';

Vue.use(Vuex);

export interface DistilState {
	routeModule: Route;
	dataModule: DataState;
	solutionModule: SolutionState;
	viewModule: ViewState;
	appModule: AppState;
}

export default new Store<DistilState>({
	modules:  {
		routeModule,
		dataModule,
		solutionModule,
		viewModule,
		appModule
	},
	strict: true
});
