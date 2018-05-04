import _ from 'lodash';
import moment from 'moment';
import Vue from 'vue';
import { SolutionState, SolutionInfo } from './index';

export const mutations = {

	// adds a solution request or replaces an existing one if the ids match.
	updateSolutionRequests(state: SolutionState, solution: SolutionInfo) {
		const index = _.findIndex(state.solutionRequests, p => {
			return p.solutionId === solution.solutionId;
		});
		if (index === -1) {
			state.solutionRequests.push(solution);
		} else {
			if (moment(solution.timestamp) >= moment(state.solutionRequests[index].timestamp)) {
				Vue.set(state.solutionRequests, index, solution);
			}
		}
	},

	clearSolutionRequests(state: SolutionState) {
		state.solutionRequests = [];
	}
}
