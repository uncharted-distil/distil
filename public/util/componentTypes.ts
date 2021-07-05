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
import VueRouter from "vue-router";
import {
  Highlight,
  RowSelection,
  TableRow,
  Variable,
  VariableSummary,
} from "../store/dataset";
import { Solution } from "../store/requests";
import { ExplorerStateNames } from "./dataExplorer";
import { RouteArgs } from "./routes";
import { BaseState } from "./state/AppStateWrapper";

/**
 * Add any component types needed for typescript
 * Add a path above the interface
 */

// public/views/DataExplorer.vue
export interface DataExplorerRef {
  // computes
  dataset: string;
  isClone: boolean | null;
  highlights: Highlight[];
  isFilteringHighlights: boolean;
  isFilteringSelection: boolean;
  items: TableRow[];
  target: Variable;
  training: string[];
  solution: Solution;
  fittedSolutionId: string;
  rowSelection: RowSelection;
  variables: Variable[];
  summaries: VariableSummary[];
  isRemoteSensing: boolean;
  // data
  labelName: string;
  state: BaseState;
  isBusy: boolean;
  // methods
  isFittedSolutionIdSavedAsModel: (id: string) => boolean;
  updateRoute: (args: RouteArgs) => void;
  changeStatesByName: (name: ExplorerStateNames) => Promise<void>;
  resetHighlightsOrRow: () => void;
  updateTask: () => Promise<void>;
  // globals
  $refs: {
    [key: string]: Vue | Element | Vue[] | Element[];
  };
  $router: VueRouter;
  $bvToast: BvToast;
  $bvModal: BvModal;
}

// public/components/layout/ActionColumn.vue
export interface ActionColumnRef extends Vue {
  toggle: (paneId: string) => void;
}
// public/components/CreateSolutionsForm.vue
export interface CreateSolutionsFormRef extends Vue {
  pending: boolean;
}
// public/components/SaveModal.vue
export interface SaveModalRef extends Vue {
  showSuccessModel: () => void;
}
// public/components/SelectDataTable.vue && public/components/ImageMosaic.vue
export interface DataView extends Vue {
  selectAll: () => void;
}
