import { Predictions } from "../store/requests";
import { getters as requestGetters } from "../store/requests/module";
import store from "../store/store";
import moment from "moment";
import _ from "lodash";

export function getPredictionsById(
  predictions: Predictions[],
  predictionsId: string
): Predictions {
  const id = predictionsId || "";
  return predictions.find((r) => r.requestId === id);
}

// Finds the index to assign to a given prediction, based on timestamps of prediction execution.
export function getPredictionsIndex(predictionId: string): number {
  // Get the solutions sorted by score.
  const predictions = [...requestGetters.getRelevantPredictions(store)];

  // Sort the solutions by timestamp if they are not part of the same request.
  predictions.sort((a, b) => {
    if (b.requestId !== a.requestId) {
      return moment(b.timestamp).unix() - moment(a.timestamp).unix();
    }
    return -1;
  });

  const index = _.findIndex(predictions, (prediction) => {
    return prediction.requestId === predictionId;
  });

  return predictions.length - index - 1;
}
