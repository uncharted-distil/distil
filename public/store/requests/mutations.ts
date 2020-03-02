import _ from "lodash";
import Vue from "vue";
import { RequestState, SolutionRequest, Solution, Request } from "./index";

export const mutations = {
  updateSolutions(state: RequestState, solution: Solution) {
    const index = state.solutions.findIndex(
      r => r.solutionId === solution.solutionId
    );
    if (index === -1) {
      state.solutions.push(solution);
    } else {
      Vue.set(state.solutions, index, solution);
    }
  },

  updateSolutionRequests(state: RequestState, request: SolutionRequest) {
    const index = state.searchRequests.findIndex(
      r => r.requestId === request.requestId
    );
    if (index === -1) {
      state.searchRequests.push(request);
    } else {
      Vue.set(state.searchRequests, index, request);
    }
  },

  clearSolutions(state: RequestState) {
    state.solutions = [];
  },

  clearSolutionRequests(state: RequestState) {
    state.searchRequests = [];
  }
};
