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

import { VariableSummary, Variable } from "../store/dataset";
import store from "../store/store";
import { getters as resultGetters } from "../store/results/module";
import { getters as predictionGetters } from "../store/predictions/module";
import { CATEGORICAL_TYPE } from "./types";

export function getIDFromKey(key: string): string {
  return key ? key.split(":")[0] : "";
}
export function getTypeFromKey(key: string): string {
  return key ? key.split(":")[1] : "";
}
export function resultSummariesToVariables(solutionID: string): Variable[] {
  const summaries = [
    getSolutionResultSummary(solutionID),
    getResidualSummary(solutionID),
    getCorrectnessSummary(solutionID),
    getPredictionResultSummary(solutionID),
    getConfidenceSummary(solutionID),
    getRankingSummary(solutionID),
  ];
  const variables = [];
  summaries.forEach((sum) => {
    if (sum) {
      variables.push(summaryToVariable(sum));
    }
  });
  return variables;
}

export function summaryToVariable(summary: VariableSummary): Variable {
  return {
    datasetName: summary.dataset,
    colDisplayName: summary.label,
    key: summary.key,
    colName: summary.label,
    colType: summary.type,
    importance: null,
    colOriginalType: CATEGORICAL_TYPE,
    colDescription: summary.description,
    suggestedTypes: [],
    isColTypeChanged: false,
    grouping: null,
    isColTypeReviewed: false,
    min: summary.baseline?.extrema.min,
    max: summary.baseline?.extrema.max,
    values: summary.baseline?.buckets.map((b) => b.key),
    distilRole: "",
    role: [],
    novelty: null,
  };
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
