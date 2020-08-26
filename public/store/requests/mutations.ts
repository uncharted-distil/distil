import _ from "lodash";
import Vue from "vue";
import { RequestState, SolutionRequest, Solution, Predictions } from "./index";

export const mutations = {
  updateSolutions(state: RequestState, solution: Solution) {
    const index = state.solutions.findIndex(
      (r) => r.solutionId === solution.solutionId
    );
    if (index === -1) {
      state.solutions.push(solution);
    } else {
      Vue.set(state.solutions, index, solution);
    }
  },

  updateSolutionRequests(state: RequestState, request: SolutionRequest) {
    const index = state.solutionRequests.findIndex(
      (r) => r.requestId === request.requestId
    );
    if (index === -1) {
      state.solutionRequests.push(request);
    } else {
      Vue.set(state.solutionRequests, index, request);
    }
  },

  clearSolutions(state: RequestState) {
    state.solutions = [];
  },

  clearSolutionRequests(state: RequestState) {
    state.solutionRequests = [];
  },

  updatePredictions(state: RequestState, predictions: Predictions) {
    const index = state.predictions.findIndex(
      (r) => r.requestId === predictions.requestId
    );
    if (index === -1) {
      state.predictions.push(predictions);
    } else {
      Vue.set(state.predictions, index, predictions);
    }
  },

  clearPredictions(state: RequestState) {
    state.predictions = [];
  },
};
