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
import { DistilState } from "../store";
import { state, AppState } from "./index";
import { getters as moduleGetters } from "./getters";
import { actions as moduleActions } from "./actions";
import { mutations as moduleMutations } from "./mutations";
import { getStoreAccessors } from "vuex-typescript";

export const appModule: Module<AppState, DistilState> = {
  state: state,
  getters: moduleGetters,
  actions: moduleActions,
  mutations: moduleMutations,
};

const { commit, read, dispatch } = getStoreAccessors<AppState, DistilState>(
  null
);

// typed getters
export const getters = {
  getVersionNumber: read(moduleGetters.getVersionNumber),
  getHelpURL: read(moduleGetters.getHelpURL),
  getVersionTimestamp: read(moduleGetters.getVersionTimestamp),
  getProblemDataset: read(moduleGetters.getProblemDataset),
  getProblemTarget: read(moduleGetters.getProblemTarget),
  getProblemMetrics: read(moduleGetters.getProblemMetrics),
  getStatusPanelState: read(moduleGetters.getStatusPanelState),
  getTA2VersionNumber: read(moduleGetters.getTA2VersionNumber),
  getAllSystemVersions: read(moduleGetters.getAllSystemVersions),
  isPrototype: read(moduleGetters.isPrototype),
  getTrainTestSplit: read(moduleGetters.getTrainTestSplit),
  getTrainTestSplitTimeSeries: read(moduleGetters.getTrainTestSplitTimeSeries),
  getShouldScaleImages: read(moduleGetters.getShouldScaleImages),
};

// typed actions
export const actions = {
  saveModel: dispatch(moduleActions.saveModel),
  exportSolution: dispatch(moduleActions.exportSolution),
  exportProblem: dispatch(moduleActions.exportProblem),
  fetchConfig: dispatch(moduleActions.fetchConfig),
  openStatusPanelWithContentType: dispatch(
    moduleActions.openStatusPanelWithContentType
  ),
  closeStatusPanel: dispatch(moduleActions.closeStatusPanel),
  logUserEvent: dispatch(moduleActions.logUserEvent),
  togglePrototype: dispatch(moduleActions.togglePrototype),
};

// type mutators
export const mutations = {
  setVersionNumber: commit(moduleMutations.setVersionNumber),
  setHelpURL: commit(moduleMutations.setHelpURL),
  setVersionTimestamp: commit(moduleMutations.setVersionTimestamp),
  setProblemDataset: commit(moduleMutations.setProblemDataset),
  setProblemTarget: commit(moduleMutations.setProblemTarget),
  setProblemMetrics: commit(moduleMutations.setProblemMetrics),
  setStatusPanelContentType: commit(moduleMutations.setStatusPanelContentType),
  setTA2VersionNumber: commit(moduleMutations.setTA2VersionNumber),
  openStatusPanel: commit(moduleMutations.openStatusPanel),
  closeStatusPanel: commit(moduleMutations.closeStatusPanel),
  togglePrototype: commit(moduleMutations.togglePrototype),
  setTrainTestSplit: commit(moduleMutations.setTrainTestSplit),
  setTrainTestSplitTimeSeries: commit(
    moduleMutations.setTrainTestSplitTimeSeries
  ),
  setShouldScaleImages: commit(moduleMutations.setShouldScaleImages),
};
