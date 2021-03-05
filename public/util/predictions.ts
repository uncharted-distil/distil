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
