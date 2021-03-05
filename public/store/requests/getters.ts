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
import moment from "moment";
import { Variable } from "../dataset/index";
import {
  RequestState,
  Solution,
  SolutionRequest,
  SolutionStatus,
  Predictions,
  PredictStatus,
} from "./index";
import { getVarType } from "../../util/types";

export function sortRequestsByTimestamp(
  a: SolutionRequest,
  b: SolutionRequest
): number {
  // descending order
  return moment(b.timestamp).unix() - moment(a.timestamp).unix();
}

function getScoreValue(s: Solution): number {
  if (s.progress === SolutionStatus.SOLUTION_ERRORED) {
    return -1;
  }
  return s.scores && s.scores.length > 0
    ? s.scores[0].value * s.scores[0].sortMultiplier
    : -1;
}

// Sorts in descending order of score
export function sortSolutionsByScore(a: Solution, b: Solution): number {
  return getScoreValue(b) - getScoreValue(a);
}

export const getters = {
  // Returns in-progress search results.
  getRunningSolutions(state: RequestState): Solution[] {
    return state.solutions
      .filter(
        (result) =>
          result.progress === SolutionStatus.SOLUTION_FITTING ||
          result.progress === SolutionStatus.SOLUTION_SCORING ||
          result.progress === SolutionStatus.SOLUTION_PRODUCING
      )
      .sort(sortSolutionsByScore);
  },

  // Returns completed search results.
  getCompletedSolutions(state: RequestState): Solution[] {
    return state.solutions
      .filter(
        (solution) => solution.progress === SolutionStatus.SOLUTION_COMPLETED
      )
      .sort(sortSolutionsByScore);
  },

  // Returns all search results.
  getSolutions(state: RequestState): Solution[] {
    return state.solutions;
  },

  // Returns search results relevant to the current dataset and target.
  getRelevantSolutions(state: RequestState, getters: any): Solution[] {
    const target = <string>getters.getRouteTargetVariable;
    const dataset = <string>getters.getRouteDataset;
    return state.solutions
      .filter(
        (result) => result.dataset === dataset && result.feature === target
      )
      .sort(sortSolutionsByScore);
  },

  // Returns search requests relevant to the current dataset and target.
  getRelevantSolutionRequests(
    state: RequestState,
    getters: any
  ): SolutionRequest[] {
    const target = <string>getters.getRouteTargetVariable;
    const dataset = <string>getters.getRouteDataset;
    // get only matching dataset / target
    return state.solutionRequests
      .filter(
        (request) => request.dataset === dataset && request.feature === target
      )
      .sort(sortRequestsByTimestamp);
  },

  // Returns search requests IDs relevant to the current dataset and target.
  getRelevantSolutionRequestIds(state: RequestState, getters: any): string[] {
    return (<SolutionRequest[]>getters.getRelevantSolutionRequests).map(
      (request) => request.requestId
    );
  },

  // Returns currently selected search result.
  getActiveSolution(state: RequestState, getters: any): Solution {
    const solutionId = <string>getters.getRouteSolutionId;
    const solutions = <Solution[]>getters.getSolutions;
    return _.find(solutions, (solution) => solution.solutionId === solutionId);
  },

  // Returns training variables associated with the currently selected search result.
  getActiveSolutionTrainingVariables(
    state: RequestState,
    getters: any
  ): Variable[] {
    const activeSolution = <Solution>getters.getActiveSolution;
    if (!activeSolution || !activeSolution.features) {
      return [];
    }
    const variables = <Variable[]>getters.getVariablesMap;
    return activeSolution.features
      .filter((f) => f.featureType === "train")
      .map((f) => variables[f.featureName])
      .filter((v) => !!v);
  },

  // Returns target variable associated with the currently selected search result.
  getActiveSolutionTargetVariable(
    state: RequestState,
    getters: any
  ): Variable[] {
    const target = <string>getters.getRouteTargetVariable;
    const variables = <Variable[]>getters.getVariables;
    return variables.filter((variable) => variable.key === target);
  },

  // Returns in-progress predictions.
  getRunningPredictions(state: RequestState): Predictions[] {
    return state.predictions.filter(
      (result) =>
        result.progress === PredictStatus.PREDICT_RUNNING ||
        result.progress === PredictStatus.PREDICT_PENDING
    );
  },

  // Returns completed predictions.
  getCompletedPredictions(state: RequestState): Predictions[] {
    return state.predictions.filter(
      (result) => result.progress !== PredictStatus.PREDICT_COMPLETED
    );
  },

  // Returns all predictions.
  getPredictions(state: RequestState): Predictions[] {
    return state.predictions;
  },

  // Returns predictions relevant to the currently selected fitted solution id.
  getRelevantPredictions(state: RequestState, getters: any): Predictions[] {
    return state.predictions.filter(
      (result) =>
        result.fittedSolutionId === <string>getters.getRouteFittedSolutionId
    );
  },

  // Returns currently selected predictions
  getActivePredictions(state: RequestState, getters: any): Predictions {
    const predictionsId = <string>getters.getRouteProduceRequestId;
    const predictions = <Predictions[]>getters.getPredictions;
    return predictions.find((p) => p.requestId === predictionsId);
  },

  // Returns training variables associated with the currently selected fitted model
  getActivePredictionTrainingVariables(
    state: RequestState,
    getters: any
  ): Variable[] {
    const predictions = <Predictions>getters.getActivePredictions;
    if (!predictions || !predictions.features) {
      return [];
    }
    const variables = <Variable[]>getters.getVariablesMap;
    return predictions.features.map((p) => variables[p.featureName]);
  },
};
