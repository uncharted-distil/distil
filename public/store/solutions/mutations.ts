import _ from "lodash";
import moment from "moment";
import Vue from "vue";
import { SolutionState, SolutionRequest } from "./index";
import { Stream } from "../../util/ws";
import { sortSolutionsByScore } from "./getters";

export const mutations = {
  updateSolutionRequests(state: SolutionState, request: SolutionRequest) {
    const index = _.findIndex(state.requests, r => {
      return r.requestId === request.requestId;
    });
    if (index === -1) {
      // add if it does not exist already
      state.requests.push(request);
      // sort solutions
      request.solutions.sort(sortSolutionsByScore);
    } else {
      const existing = state.requests[index];
      // update progress
      existing.progress = request.progress;
      // update solutions
      request.solutions.forEach(solution => {
        const solutionIndex = _.findIndex(existing.solutions, s => {
          return s.solutionId === solution.solutionId;
        });
        if (solutionIndex === -1) {
          // add if it does not exist already
          existing.solutions.push(solution);
        } else {
          // otherwise replace
          if (
            moment(solution.timestamp) >
            moment(existing.solutions[solutionIndex].timestamp)
          ) {
            Vue.set(existing.solutions, solutionIndex, solution);
          }
        }
      });
      // sort solutions
      existing.solutions.sort(sortSolutionsByScore);
    }
  },

  clearSolutionRequests(state: SolutionState) {
    state.requests = [];
  },

  addRequestStream(
    state: SolutionState,
    args: { requestId: string; stream: Stream }
  ) {
    Vue.set(state.streams, args.requestId, args.stream);
  },

  removeRequestStream(state: SolutionState, args: { requestId: string }) {
    Vue.delete(state.streams, args.requestId);
  }
};
