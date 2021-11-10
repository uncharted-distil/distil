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

import { BvModal } from "bootstrap-vue/src/components/modal";
import { BvToast } from "bootstrap-vue/src/components/toast";
import Vue from "vue";
import VueRouter, { Route } from "vue-router";
import {
  Highlight,
  RowSelection,
  TableRow,
  Variable,
  VariableSummary,
} from "../store/dataset";
import { Solution } from "../store/requests";
import ExplorerConfig, {
  Action,
  ActionNames,
  ExplorerStateNames,
} from "./explorer";
import { RouteArgs } from "./routes";
import { BaseState } from "./state/AppStateWrapper";

/**
 * Add any component types needed for typescript
 * Add a path above the interface
 */
// public/views/DataExplorer.vue
export interface DataExplorerRef {
  // computes
  activeVariables: Variable[];
  activeViews: string[];
  allVariables: Variable[];
  availableActions: Action[];
  baselineItems: TableRow[];
  dataset: string;
  explore: string[];
  explorerRouteState: ExplorerStateNames;
  fittedSolutionId: string;
  geoVarExists: boolean;
  highlights: Highlight[];
  inactiveMetaTypes: string[];
  isClone: boolean | null;
  isFilteringHighlights: boolean;
  isFilteringSelection: boolean;
  isRemoteSensing: boolean;
  isSelectState: boolean;
  items: TableRow[];
  numRows: number;
  rowSelection: RowSelection;
  secondaryVariables: Variable[];
  solution: Solution;
  summaries: VariableSummary[];
  target: Variable;
  task: string;
  totalNumRows: number;
  training: string[];
  variables: Variable[];
  variablesPerActions: Record<string, Variable[]>;
  variablesTypes: string[];
  viewComponent: string;

  // data
  activeView: number;
  config: ExplorerConfig;
  busyState: string;
  dataLoading: boolean;
  include: boolean;
  isBusy: boolean;
  labelModalId: string;
  labelName: string;
  labelNameState: boolean;
  metaTypes: string[];
  state: BaseState;
  unsaveModalId: string;
  clusterCount: number;

  // methods
  bindEventHandlers: (eventHandlers: Record<string, Function>) => void;
  changeStatesByName: (name: ExplorerStateNames) => Promise<void>;
  isFittedSolutionIdSavedAsModel: (id: string) => boolean;
  preSelectTopVariables: (num?: number) => void;
  removeEventHandlers: (eventHandlers: Record<string, Function>) => void;
  resetHighlightsOrRow: () => void;
  setConfig: (config: ExplorerConfig) => void;
  setState: (state: BaseState) => void;
  toggleAction: (actionName: ActionNames) => void;
  updateRoute: (args: RouteArgs) => void;
  updateTask: () => Promise<void>;
  setBusyState: (isBusy: boolean, busyState?: string) => void;
  isCurrentDatasetSaved: () => Promise<boolean>;
  // globals
  $bvModal: BvModal;
  $bvToast: BvToast;
  $nextTick(callback: (this: this) => void): void;
  $refs: {
    [key: string]: Vue | Element | Vue[] | Element[];
  };
  $route: Route;
  $router: VueRouter;
}

// public/components/layout/ActionColumn.vue
export interface ActionColumnRef extends Vue {
  toggle: (paneId: string) => void;
}
// public/components/CreateSolutionsForm.vue
export interface CreateSolutionsFormRef extends Vue {
  pending: boolean;
  success: () => void;
  fail: (err: Error) => void;
}
// public/components/SaveModal.vue
export interface SaveModalRef extends Vue {
  hideSaveForm: () => void;
  isSaving: boolean;
}
// public/components/SelectDataTable.vue && public/components/ImageMosaic.vue
export interface DataView extends Vue {
  selectAll: () => void;
}
