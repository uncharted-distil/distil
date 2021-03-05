/**
 *
 *    Copyright © 2021 Uncharted Software Inc.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

import Vue from "vue";
import Vuex, { Store } from "vuex";
import { routeModule } from "./route/module";
import { Route } from "vue-router";
import { datasetModule } from "./dataset/module";
import { DatasetState } from "./dataset/index";
import { resultsModule } from "./results/module";
import { ResultsState } from "./results/index";
import { requestsModule } from "./requests/module";
import { RequestState } from "./requests/index";
import { predictionsModule } from "./predictions/module";
import { PredictionState } from "./predictions/index";
import { modelModule } from "./model/module";
import { ModelState } from "./model/index";
import { viewModule } from "./view/module";
import { ViewState } from "./view/index";
import { appModule } from "./app/module";
import { AppState } from "./app/index";

Vue.use(Vuex);

export interface DistilState {
  routeModule: Route;
  datasetModule: DatasetState;
  requestsModule: RequestState;
  resultsModule: ResultsState;
  predictionsModule: PredictionState;
  modelModule: ModelState;
  viewModule: ViewState;
  appModule: AppState;
}

const store = new Store<DistilState>({
  modules: {
    routeModule,
    datasetModule,
    requestsModule,
    resultsModule,
    predictionsModule,
    modelModule,
    viewModule,
    appModule,
  },
  strict: true,
});

export default store;
