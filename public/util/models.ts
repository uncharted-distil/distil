import _ from "lodash";
import store from "../store/store";
import { getters as modelGetters } from "../store/model/module";

/**
 * Name of the model of a fitted solution.
 * @param  {String} fittedSolutionId
 * @return {String}
 */
export function getModelNameByFittedSolutionId(
  fittedSolutionId: string | null
): string {
  const model = modelGetters
    .getModels(store)
    .find(model => model.fittedSolutionId === fittedSolutionId);

  // Return the name if it exist, null otherwise.
  return _.isNil(_.get(model, "modelName")) ? null : model.modelName;
}
