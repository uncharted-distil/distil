import { AppState, Session } from './index';

export const mutations = {
	setWebSocketConnection(state: AppState, connection: WebSocket) {
		state.wsConnection = connection;
	},

	// sets the active session in the store as well as in the browser local storage
	setPipelineSession(state: AppState, session: Session) {
		state.pipelineSession = session;
		if (!session) {
			window.localStorage.removeItem('pipeline-session-id');
		} else {
			window.localStorage.setItem('pipeline-session-id', session.id);
		}
	},

	addRecentDataset(state: AppState, dataset: string) {
		const datasetsStr = window.localStorage.getItem('recent-datasets');
		const datasets = (datasetsStr) ? datasetsStr.split(',') : [];
		datasets.unshift(dataset);
		window.localStorage.setItem('recent-datasets', datasets.join(','));
	}
};

