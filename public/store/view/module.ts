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

import { Module } from "vuex";
import { state, ViewState } from "./index";
import { getters as moduleGetters } from "./getters";
import { actions as moduleActions } from "./actions";
import { mutations as moduleMutations } from "./mutations";
import { DistilState } from "../store";
import { getStoreAccessors } from "vuex-typescript";

export const viewModule: Module<ViewState, DistilState> = {
  state: state,
  actions: moduleActions,
  getters: moduleGetters,
  mutations: moduleMutations,
};

const { commit, read, dispatch } = getStoreAccessors<ViewState, DistilState>(
  null
);

export const getters = {
  getFetchParamsCache: read(moduleGetters.getFetchParamsCache),
};

export const mutations = {
  setFetchParamsCache: commit(moduleMutations.setFetchParamsCache),
};

export const actions = {
  clearAllData: dispatch(moduleActions.clearAllData),
  clearDatasetTableData: dispatch(moduleActions.clearDatasetTableData),
  fetchHomeData: dispatch(moduleActions.fetchHomeData),
  fetchSearchData: dispatch(moduleActions.fetchSearchData),
  fetchJoinDatasetsData: dispatch(moduleActions.fetchJoinDatasetsData),
  clearJoinDatasetsData: dispatch(moduleActions.clearJoinDatasetsData),
  updateJoinDatasetsData: dispatch(moduleActions.updateJoinDatasetsData),
  fetchDataExplorerData: dispatch(moduleActions.fetchDataExplorerData),
  updateDataExplorerData: dispatch(moduleActions.updateDataExplorerData),
  fetchSelectTargetData: dispatch(moduleActions.fetchSelectTargetData),
  fetchSelectTrainingData: dispatch(moduleActions.fetchSelectTrainingData),
  updateSelectTrainingData: dispatch(moduleActions.updateSelectTrainingData),
  updateLabelData: dispatch(moduleActions.updateLabelData),
  updateHighlight: dispatch(moduleActions.updateHighlight),
  clearHighlight: dispatch(moduleActions.clearHighlight),
  fetchResultsData: dispatch(moduleActions.fetchResultsData),
  updateResultsSummaries: dispatch(moduleActions.updateResultsSummaries),
  updateResultsSolution: dispatch(moduleActions.updateResultsSolution),
  fetchPredictionsData: dispatch(moduleActions.fetchPredictionsData),
  updateResultAreaOfInterest: dispatch(
    moduleActions.updateResultAreaOfInterest
  ),
  updatePredictionAreaOfInterest: dispatch(
    moduleActions.updatePredictionAreaOfInterest
  ),
  updatePredictionTrainingSummaries: dispatch(
    moduleActions.updatePredictionTrainingSummaries
  ),
  updatePrediction: dispatch(moduleActions.updatePredictions),
  updateAreaOfInterest: dispatch(moduleActions.updateAreaOfInterest),
};
