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

import { Model, ModelState } from "./index";
import _ from "lodash";

export const mutations = {
  // replace the map of saved models
  setModels(state: ModelState, models: Model[]) {
    state.models = _.keyBy(models, (m) => m.fittedSolutionId);
  },

  // replace the list of filtered models (a subset of the full saved model list)
  setFilteredModels(state: ModelState, models: Model[]) {
    if (!models) return;
    state.filteredModelIds = models.map((m) => m.fittedSolutionId);
  },
};
