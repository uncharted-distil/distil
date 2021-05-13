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

import _, { Dictionary } from "lodash";
import moment from "moment";

import { sortRequestsByTimestamp } from "../store/requests/getters";
import {
  getters as requestGetters,
  actions as requestActions,
} from "../store/requests/module";
import { getters as routeGetters } from "../store/route/module";
import { actions as dataActions } from "../store/dataset/module";
import { createRouteEntry, overlayRouteEntry } from "../util/routes";
import { Solution, SolutionStatus } from "../store/requests/index";
import { APPLY_MODEL_ROUTE } from "../store/route/index";
import store from "../store/store";
import VueRouter, { Route } from "vue-router";

export const SOLUTION_LABELS: Dictionary<string> = {
  [SolutionStatus.SOLUTION_PENDING]: "PENDING",
  [SolutionStatus.SOLUTION_FITTING]: "FITTING",
  [SolutionStatus.SOLUTION_SCORING]: "SCORING",
  [SolutionStatus.SOLUTION_PRODUCING]: "PREDICTING",
  [SolutionStatus.SOLUTION_COMPLETED]: "COMPLETED",
  [SolutionStatus.SOLUTION_CANCELLED]: "CANCELLED",
  [SolutionStatus.SOLUTION_ERRORED]: "ERRORED",
};

export const SOLUTION_PROGRESS: Dictionary<number> = {
  [SolutionStatus.SOLUTION_PENDING]: 0,
  [SolutionStatus.SOLUTION_FITTING]: 25,
  [SolutionStatus.SOLUTION_SCORING]: 75,
  [SolutionStatus.SOLUTION_PRODUCING]: 80,
  [SolutionStatus.SOLUTION_COMPLETED]: 100,
};

export function filterBadRequests(
  solutions: Solution[],
  requestIds: string[]
): string[] {
  const solutionMap = new Map(
    solutions.map((s) => {
      return [s.requestId, s.progress !== SolutionStatus.SOLUTION_ERRORED];
    })
  );
  return requestIds.filter((r) => {
    return solutionMap.get(r);
  });
}

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

export function isTopSolutionByTime(
  solutions: Solution[],
  solutionId: string,
  n: number
): boolean {
  if (!solutionId) {
    return null;
  }
  const topN = [...solutions].sort(sortRequestsByTimestamp).slice(0, n);
  return !!topN.find((result) => result.solutionId === solutionId);
}
export function reviseOpenSolutions(
  requestId: string,
  route: Route,
  router: VueRouter
) {
  const openSolutions = routeGetters.getRouteOpenSolutions(store);
  const idx = openSolutions.findIndex((s) => {
    return s === requestId;
  });
  if (idx != -1) {
    openSolutions.splice(idx, 1);
  } else {
    openSolutions.push(requestId);
  }
  const entry = overlayRouteEntry(route, {
    openSolutions: JSON.stringify(openSolutions),
  });
  router.push(entry).catch((err) => console.warn(err));
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
