import _, { Dictionary } from "lodash";
import moment from "moment";

import { sortSolutionsByScore } from "../store/requests/getters";
import {
  getters as requestGetters,
  actions as requestActions,
} from "../store/requests/module";
import { getters as routeGetters } from "../store/route/module";
import { actions as dataActions } from "../store/dataset/module";
import { createRouteEntry } from "../util/routes";
import {
  Solution,
  SOLUTION_PENDING,
  SOLUTION_FITTING,
  SOLUTION_SCORING,
  SOLUTION_PRODUCING,
  SOLUTION_COMPLETED,
  SOLUTION_ERRORED,
  SOLUTION_CANCELLED,
} from "../store/requests/index";
import { APPLY_MODEL_ROUTE } from "../store/route/index";
import store from "../store/store";
import VueRouter from "vue-router";

export const SOLUTION_LABELS: Dictionary<string> = {
  [SOLUTION_PENDING]: "PENDING",
  [SOLUTION_FITTING]: "FITTING",
  [SOLUTION_SCORING]: "SCORING",
  [SOLUTION_PRODUCING]: "PREDICTING",
  [SOLUTION_COMPLETED]: "COMPLETED",
  [SOLUTION_CANCELLED]: "CANCELLED",
  [SOLUTION_ERRORED]: "ERRORED",
};

export const SOLUTION_PROGRESS: Dictionary<number> = {
  [SOLUTION_PENDING]: 0,
  [SOLUTION_FITTING]: 25,
  [SOLUTION_SCORING]: 75,
  [SOLUTION_PRODUCING]: 80,
  [SOLUTION_COMPLETED]: 100,
};

export function getSolutionIndex(solutionId: string) {
  // Get the solutions sorted by score.
  const solutions = [...requestGetters.getRelevantSolutions(store)];

  // Sort the solutions by timestamp if they are not part of the same request.
  solutions.sort((a, b) => {
    if (b.requestId !== a.requestId) {
      return moment(b.timestamp).unix() - moment(a.timestamp).unix();
    }

    return -1;
  });

  const index = _.findIndex(solutions, (solution) => {
    return solution.solutionId === solutionId;
  });

  return solutions.length - index - 1;
}

export function getSolutionRequestIndex(requestId: string) {
  const requests = requestGetters.getRelevantSolutionRequests(store);
  const index = _.findIndex(requests, (req) => {
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
  return solutions.filter((result) => ids.has(result.requestId));
}

// Returns a specific solution result given a request and its solution id.
export function getSolutionById(
  solutions: Solution[],
  solutionId: string
): Solution {
  if (!solutionId) {
    return null;
  }
  return solutions.find((result) => result.solutionId === solutionId);
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
    .filter((req) => req.requestId === requestId)
    .slice()
    .sort(sortSolutionsByScore)
    .slice(0, n);
  return !!topN.find((result) => result.solutionId === solutionId);
}

export async function openModelSolution(
  router: VueRouter,
  args: {
    datasetId: string;
    targetFeature: string;
    fittedSolutionId?: string;
    solutionId?: string;
    variableFeatures: string[];
  }
) {
  let task = routeGetters.getRouteTask(store);
  if (!task) {
    const taskResponse = await dataActions.fetchTask(store, {
      dataset: args.datasetId,
      targetName: args.targetFeature,
      variableNames: args.variableFeatures, // solution.features.map(f => f.featureName)
    });
    task = taskResponse.data.task.join(",");
  }
  const solutionArgs = {
    dataset: args.datasetId,
    target: args.targetFeature,
  };
  await Promise.all([
    requestActions.fetchSolutionRequests(store, solutionArgs),
    requestActions.fetchSolutions(store, solutionArgs),
  ]);
  const solutionId = args.solutionId
    ? args.solutionId
    : requestGetters
        .getSolutions(store)
        .find((s) => s.fittedSolutionId === args.fittedSolutionId).solutionId;
  const routeDefintion = {
    dataset: args.datasetId,
    target: args.targetFeature,
    task: task,
    solutionId: solutionId,
    singleSolution: true.toString(),
    applyModel: true.toString(),
  };

  const entry = createRouteEntry(APPLY_MODEL_ROUTE, routeDefintion);
  router.push(entry).catch((err) => {
    console.warn(err);
  });
}
