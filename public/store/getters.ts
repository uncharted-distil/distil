import Connection from '../util/ws';
import { DistilState } from './index';
import { GetterTree } from 'vuex';

export const getters: GetterTree<DistilState, any> = {
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

