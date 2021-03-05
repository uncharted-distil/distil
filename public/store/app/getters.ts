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

import { AppState, StatusPanelState } from "./index";

export const getters = {
  getVersionNumber(state: AppState): string {
    return state.versionNumber;
  },

  getHelpURL(state: AppState): string {
    return state.helpURL;
  },

  getVersionTimestamp(state: AppState): string {
    return state.versionTimestamp;
  },

  getProblemDataset(state: AppState): string {
    return state.problemDataset;
  },

  getProblemTarget(state: AppState): string {
    return state.problemTarget;
  },

  getProblemMetrics(state: AppState): string[] {
    return state.problemMetrics;
  },

  getStatusPanelState(state: AppState): StatusPanelState {
    return state.statusPanelState;
  },

  getTA2VersionNumber(state: AppState): string {
    return state.ta2Version;
  },

  getAllSystemVersions(state: AppState): string {
    const ta2Version = state.ta2Version;
    const ta3Version = state.versionNumber;
    const ta3Timestamp =
      state.versionTimestamp !== "unset" ? state.versionTimestamp : "";
    return `TA2 version ${ta2Version}\nTA3 version ${ta3Version} ${ta3Timestamp}`.trim();
  },

  isPrototype(state: AppState) {
    return state.prototype;
  },

  getTrainTestSplit(state: AppState): number {
    return state.trainTestSplit;
  },
  getTrainTestSplitTimeSeries(state: AppState): number {
    return state.trainTestSplitTimeSeries;
  },
  getShouldScaleImages(state: AppState): boolean {
    return state.shouldScaleImages;
  },
};
