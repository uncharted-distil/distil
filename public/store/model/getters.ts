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
import { isInteger, values } from "lodash";

export const getters = {
  getFilteredModels(state: ModelState): Model[] {
    return state.filteredModelIds.map((m) => state.models[m]);
  },

  getModels(state: ModelState): Model[] {
    return values(state.models);
  },

  getCountOfModels(state: ModelState): number {
    const count = values(state.models).length;
    return isInteger(count) ? count : 0;
  },
};
