import _, { Dictionary } from "lodash";
import moment from "moment";

import { sortSolutionsByScore } from "../store/requests/getters";
import { getters as requestGetters } from "../store/requests/module";
import {
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

export function getSolutionIndex(solutionId: string) {
  // Get the solutions sorted by score.
  const solutions = requestGetters.getRelevantSolutions(store);

  // Sort the solutions by timestamp if they are not part of the same request.
  solutions.sort((a, b) => {
    if (b.requestId !== a.requestId) {
      return moment(b.timestamp).unix() - moment(a.timestamp).unix();
    }

    return -1;
  });

  const index = _.findIndex(solutions, solution => {
    return solution.solutionId === solutionId;
  });

  return solutions.length - index - 1;
}

export function getSolutionRequestIndex(requestId: string) {
  const requests = requestGetters.getRelevantSolutionRequests(store);
  const index = _.findIndex(requests, req => {
    return req.requestId === requestId;
  });
  return requests.length - index - 1;
}

// Utility function to return all solution results associated with a given request ID
export function getSolutionsBySolutionRequestIds(
  solutions: Solution[],
  requestIds: string[]
): Solution[] {
  const ids = new Set(requestIds);
  return solutions.filter(result => ids.has(result.requestId));
}

// Returns a specific solution result given a request and its solution id.
export function getSolutionById(
  solutions: Solution[],
  solutionId: string
): Solution {
  if (!solutionId) {
    return null;
  }
  return solutions.find(result => result.solutionId === solutionId);
}

export function isTopSolutionByScore(
  solutions: Solution[],
  requestId: string,
  solutionId: string,
  n: number
): boolean {
  if (!solutionId) {
    return null;
  }
  const topN = solutions
    .filter(req => req.requestId === requestId)
    .slice()
    .sort(sortSolutionsByScore)
    .slice(0, n);
  return !!topN.find(result => result.solutionId === solutionId);
}
