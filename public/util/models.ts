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
  fittedSolutionId: string | null,
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
  fittedSolutionId: string | null,
): boolean {
  return !!getModelByFittedSolutionId(fittedSolutionId);
}
