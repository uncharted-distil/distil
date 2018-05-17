import Vue from 'vue';
import Vuex from 'vuex';
import { Store } from 'vuex';
import { routeModule } from './route/module';
import { Route } from 'vue-router';
import { datasetModule } from './dataset/module';
import { DatasetState } from './dataset/index';
import { highlightsModule } from './highlights/module';
import { HighlightState } from './highlights/index';
import { imagesModule } from './images/module';
import { ImageState } from './images/index';
import { resultsModule } from './results/module';
import { ResultsState } from './results/index';
import { solutionModule } from './solutions/module';
import { SolutionState } from './solutions/index';
import { viewModule } from './view/module';
import { ViewState } from './view/index';
import { appModule } from './app/module';
import { AppState } from './app/index';

Vue.use(Vuex);

export interface DistilState {
	routeModule: Route;
	datasetModule: DatasetState;
	highlightsModule: HighlightState,
	solutionModule: SolutionState;
	imagesModule: ImageState;
	viewModule: ViewState;
	resultsModule: ResultsState,
	appModule: AppState;
}

export default new Store<DistilState>({
	modules:  {
		routeModule,
		datasetModule,
		highlightsModule,
		imagesModule,
		resultsModule,
		solutionModule,
		viewModule,
		appModule
	},
	strict: true
});
