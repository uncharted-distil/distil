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

import { AppState, StatusPanelContentType } from "./index";
import Vue from "vue";

export const mutations = {
  setVersionNumber(state: AppState, versionNumber: string) {
    state.versionNumber = versionNumber;
  },

  setHelpURL(state: AppState, helpURL: string) {
    state.helpURL = helpURL;
  },

  setVersionTimestamp(state: AppState, versionTimestamp: string) {
    state.versionTimestamp = versionTimestamp;
  },

  setProblemDataset(state: AppState, dataset: string) {
    state.problemDataset = dataset;
  },

  setProblemTarget(state: AppState, target: string) {
    state.problemTarget = target;
  },

  setProblemMetrics(state: AppState, metrics: string[]) {
    state.problemMetrics = metrics;
  },

  setStatusPanelContentType(
    state: AppState,
    contentType: StatusPanelContentType
  ) {
    Vue.set(state.statusPanelState, "contentType", contentType);
  },

  setTA2VersionNumber(state: AppState, ta2Version: string) {
    state.ta2Version = ta2Version;
  },
  setShouldScaleImages(state: AppState, shouldScale: boolean) {
    state.shouldScaleImages = shouldScale;
  },
  setTrainTestSplit(state: AppState, trainTestSplit: string) {
    state.trainTestSplit = parseFloat(trainTestSplit);
  },

  setTrainTestSplitTimeSeries(
    state: AppState,
    trainTestSplitTimeSeries: string
  ) {
    state.trainTestSplitTimeSeries = parseFloat(trainTestSplitTimeSeries);
  },

  openStatusPanel(state: AppState) {
    Vue.set(state.statusPanelState, "isOpen", true);
  },

  closeStatusPanel(state: AppState) {
    Vue.set(state.statusPanelState, "isOpen", false);
  },

  togglePrototype(state: AppState) {
    state.prototype = !state.prototype;
  },
};
