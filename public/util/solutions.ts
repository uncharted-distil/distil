import _, { Dictionary } from "lodash";

import { sortSolutionsByScore } from "../store/requests/getters";
import { getters as solutionGetters } from "../store/requests/module";
import {
  RequestState,
  Solution,
  SOLUTION_PENDING,
  SOLUTION_FITTING,
  SOLUTION_SCORING,
  SOLUTION_PRODUCING,
  SOLUTION_COMPLETED,
  SOLUTION_ERRORED
} from "../store/requests/index";
import store from "../store/store";

export const SOLUTION_LABELS: Dictionary<string> = {
  [SOLUTION_PENDING]: "PENDING",
  [SOLUTION_FITTING]: "FITTING",
  [SOLUTION_SCORING]: "SCORING",
  [SOLUTION_PRODUCING]: "PREDICTING",
  [SOLUTION_COMPLETED]: "COMPLETED",
  [SOLUTION_ERRORED]: "ERRORED"
};

export const SOLUTION_PROGRESS: Dictionary<number> = {
  [SOLUTION_PENDING]: 0,
  [SOLUTION_FITTING]: 25,
  [SOLUTION_SCORING]: 75,
  [SOLUTION_PRODUCING]: 80,
  [SOLUTION_COMPLETED]: 100
};

export interface NameInfo {
  displayName: string;
  schemaName: string;
}

export function getSolutionIndex(solutionId: string) {
  const solutions = solutionGetters.getRelevantSolutions(store);
  const index = _.findIndex(solutions, solution => {
    return solution.solutionId === solutionId;
  });
  return solutions.length - index - 1;
}

export function getRequestIndex(requestId: string) {
  const requests = solutionGetters.getRelevantSearchRequests(store);
  const index = _.findIndex(requests, req => {
    return req.requestId === requestId;
  });
  return requests.length - index - 1;
}

// Utility function to return all solution results associated with a given request ID
export function getSolutionsByRequestIds(
  state: RequestState,
  requestIds: string[]
): Solution[] {
  const ids = new Set(requestIds);
  return state.solutions.filter(result => ids.has(result.requestId));
}

// Returns a specific solution result given a request and its solution id.
export function getSolutionById(
  state: RequestState,
  solutionId: string
): Solution {
  if (!solutionId) {
    return null;
  }
  return state.solutions.find(result => result.solutionId === solutionId);
}

export function isTopSolutionByScore(
  state: RequestState,
  requestId: string,
  solutionId: string,
  n: number
): boolean {
  if (!solutionId) {
    return null;
  }
  const topN = state.solutions
    .filter(req => req.requestId === requestId)
    .slice()
    .sort(sortSolutionsByScore)
    .slice(0, n);
  return !!topN.find(result => result.solutionId === solutionId);
}
