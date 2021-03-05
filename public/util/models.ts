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

import _ from "lodash";
import store from "../store/store";
import { Model } from "../store/model/index";
import { getters as modelGetters } from "../store/model/module";

/**
 * Find a saved model from a fitted solution id.
 * @param {String} fittedSolutionId
 * @returns {Model}
 */
function getModelByFittedSolutionId(fittedSolutionId: string | null): Model {
  return modelGetters
    .getModels(store)
    .find((model) => model.fittedSolutionId === fittedSolutionId);
}

/**
 * Name of the model of a fitted solution.
 * @param  {String} fittedSolutionId
 * @return {String}
 */
export function getModelNameByFittedSolutionId(
  fittedSolutionId: string | null
): string {
  const model = getModelByFittedSolutionId(fittedSolutionId);

  // Return the name if it exist, null otherwise.
  return _.isNil(_.get(model, "modelName")) ? null : model.modelName;
}

/**
 * Check if a fitted solution is a saved model.
 * @param  {String} fittedSolutionId
 * @return {Boolean}
 */
export function isFittedSolutionIdSavedAsModel(
  fittedSolutionId: string | null
): boolean {
  return !!getModelByFittedSolutionId(fittedSolutionId);
}
