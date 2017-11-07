import Connection from '../../util/ws';
import { AppState } from './index';

export const getters = {
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

	getPipelineSessionID(state: AppState) {
		return () => {
			if (!state.pipelineSession) {
				return window.localStorage.getItem('pipeline-session-id');
			}
			return state.pipelineSession.id;
		};
	},

	getPipelineSession(state: AppState) {
		return () => state.pipelineSession;
	},

	getRecentDatasets() {
		return () => {
			const datasets = window.localStorage.getItem('recent-datasets');
			return (datasets) ? datasets.split(',') : [];
		};
	}
};
