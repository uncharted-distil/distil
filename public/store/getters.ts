import _ from 'lodash';
import Connection from '../util/ws';
import { DistilState } from './index';
import { GetterTree } from 'vuex';

export const getters: GetterTree<DistilState, any> = {
	getPipelineResults(state: DistilState) {
		return (requestId: string) => {
			return _.concat(
				_.values(state.runningPipelines[requestId]),
				_.values(state.completedPipelines[requestId]));
		};
	},

	getRunningPipelines(state: DistilState) {
		return () => state.runningPipelines;
	},

	getCompletedPipelines(state: DistilState) {
		return () => state.completedPipelines;
	},

	getWebSocketConnection() {
		const conn = new Connection('/ws', (err: string) => {
			if (err) {
				console.warn(err);
				return;
			}
		});
		return () => {
			return conn;
		};
	},

	getPipelineSessionID(state: DistilState) {
		return () => {
			if (!state.pipelineSession) {
				return window.localStorage.getItem('pipeline-session-id');
			}
			return state.pipelineSession.id;
		};
	},

	getPipelineSession(state: DistilState) {
		return () => state.pipelineSession;
	},

	getRecentDatasets() {
		return () => {
			const datasets = window.localStorage.getItem('recent-datasets');
			return (datasets) ? datasets.split(',') : [];
		};
	}
};

