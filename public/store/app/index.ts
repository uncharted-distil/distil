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

import { DatasetPendingRequestType } from "../dataset/index";

export interface AppState {
  versionNumber: string;
  helpURL: string;
  versionTimestamp: string;
  problemDataset: string;
  problemTarget: string;
  problemMetrics: string[];
  statusPanelState: StatusPanelState;
  ta2Version: string;
  prototype: boolean;
  trainTestSplit: number;
  trainTestSplitTimeSeries: number;
  shouldScaleImages: boolean;
}

export interface StatusPanelState {
  isOpen: boolean;
  contentType: StatusPanelContentType;
}

// shared data model
export const state: AppState = {
  versionNumber: "unknown",
  versionTimestamp: "unknown",
  problemDataset: "unknown",
  problemTarget: "unknown",
  helpURL: "",
  problemMetrics: [],
  statusPanelState: {
    contentType: undefined,
    isOpen: false,
  },
  ta2Version: "unknown",
  prototype: false,
  trainTestSplit: null,
  trainTestSplitTimeSeries: null,
  shouldScaleImages: false,
};

export type StatusPanelContentType = DatasetPendingRequestType;
