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

import { VariableSummary } from "../store/dataset";
import store from "../store/store";
import { getters as resultGetters } from "../store/results/module";
import { getters as predictionGetters } from "../store/predictions/module";

export function getIDFromKey(key: string): string {
  return key ? key.split(":")[0] : "";
}

export function getSolutionResultSummary(solutionID: string): VariableSummary {
  return resultGetters
    .getPredictedSummaries(store)
    .find((s) => getIDFromKey(s.key) === solutionID);
}

export function getResidualSummary(solutionID: string): VariableSummary {
  return resultGetters
    .getResidualsSummaries(store)
    .find((s) => getIDFromKey(s.key) === solutionID);
}

export function getCorrectnessSummary(solutionID: string): VariableSummary {
  return resultGetters
    .getCorrectnessSummaries(store)
    .find((s) => getIDFromKey(s.key) === solutionID);
  return null;
}

export function getPredictionResultSummary(requestId: string): VariableSummary {
  return predictionGetters
    .getPredictionSummaries(store)
    .find((s) => getIDFromKey(s.key) === requestId);
}

export function getConfidenceSummary(solutionID: string): VariableSummary {
  return resultGetters
    .getConfidenceSummaries(store)
    .find((s) => getIDFromKey(s.key) === solutionID);
  return null;
}
export function getRankingSummary(solutionID: string): VariableSummary {
  return resultGetters
    .getRankingSummaries(store)
    .find((s) => getIDFromKey(s.key) === solutionID);
  return null;
}
