import _ from 'lodash';
import moment from 'moment';
import Vue from 'vue';
import { SolutionState, SolutionInfo } from './index';

export const mutations = {

	// adds a solution request or replaces an existing one if the ids match.
	updateSolutionRequests(state: SolutionState, solution: SolutionInfo) {
		const index = _.findIndex(state.solutions, p => {
			return p.solutionId === solution.solutionId;
		});
		if (index === -1) {
			state.solutions.push(solution);
		} else {
			if (moment(solution.timestamp) >= moment(state.solutions[index].timestamp)) {
				Vue.set(state.solutions, index, solution);
			}
		}
	},

	clearSolutionRequests(state: SolutionState) {
		state.solutions = [];
	}
}
