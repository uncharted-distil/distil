import _ from "lodash";
import { Dictionary } from "./dict";
import { sortSolutionsByScore } from "../store/solutions/getters";
import { getters as solutionGetters } from "../store/solutions/module";
import { SolutionState, Solution } from "../store/solutions/index";
import store from "../store/store";

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
  const requests = solutionGetters.getRelevantSolutionRequests(store);
  const index = _.findIndex(requests, req => {
    return req.requestId === requestId;
  });
  return requests.length - index - 1;
}

// Utility function to return all solution results associated with a given request ID
export function getSolutionsByRequestIds(
  state: SolutionState,
  requestIds: string[]
): Solution[] {
  const ids = {};
  requestIds.forEach(id => {
    ids[id] = true;
  });

  let solutions = [];
  const filtered = state.requests.filter(request => ids[request.requestId]);
  filtered.forEach(request => {
    solutions = solutions.concat(request.solutions);
  });
  return solutions;
}

// Returns a specific solution result given a request and its solution id.
export function getSolutionById(
  state: SolutionState,
  solutionId: string
): Solution {
  if (!solutionId) {
    return null;
  }
  let found = null;
  state.requests.forEach(request => {
    request.solutions.forEach(solution => {
      if (solution.solutionId === solutionId) {
        found = solution;
      }
    });
  });
  return found;
}

export function isTopSolutionByScore(
  state: SolutionState,
  requestId: string,
  solutionId: string,
  n: number
): boolean {
  if (!solutionId) {
    return null;
  }
  const request = _.find(state.requests, req => {
    return req.requestId === requestId;
  });

  const sortedByScore = request.solutions
    .slice()
    .sort(sortSolutionsByScore)
    .slice(0, n);

  return !!_.find(sortedByScore, sol => {
    return sol.solutionId === solutionId;
  });
}
