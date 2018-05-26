import _ from 'lodash';
import moment from 'moment';
import Vue from 'vue';
import { SolutionState, SolutionRequest } from './index';
import { sortSolutions, sortRequests } from './getters';
import { Stream } from '../../util/ws';

export const mutations = {

	updateSolutionRequests(state: SolutionState, request: SolutionRequest) {
		const index = _.findIndex(state.requests, r => {
			return r.requestId === request.requestId;
		});
		if (index === -1) {
			// add if it does not exist already
			state.requests.push(request);
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
					if (moment(solution.timestamp) > moment(existing.solutions[solutionIndex].timestamp)) {
						Vue.set(existing.solutions, solutionIndex, solution);
					}
				}
			});
		}

		// sort requests and solutions
		state.requests.forEach(request => {
			request.solutions.sort(sortSolutions);
		});
		state.requests.sort(sortRequests);
	},

	clearSolutionRequests(state: SolutionState) {
		state.requests = [];
	},

	addRequestStream(state: SolutionState, args: { requestId: string, stream: Stream }) {
		Vue.set(state.streams, args.requestId, args.stream);
	},

	removeRequestStream(state: SolutionState, args: { requestId: string } ) {
		Vue.delete(state.streams, args.requestId);
	}
}
