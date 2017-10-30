import { MutationTree } from 'vuex';
import { DistilState, Session } from './index';

export const mutations: MutationTree<DistilState> = {
	setWebSocketConnection(state: DistilState, connection: WebSocket) {
		state.wsConnection = connection;
	},

	// sets the active session in the store as well as in the browser local storage
	setPipelineSession(state: DistilState, session: Session) {
		state.pipelineSession = session;
		if (!session) {
			window.localStorage.removeItem('pipeline-session-id');
		} else {
			window.localStorage.setItem('pipeline-session-id', session.id);
		}
	},

	addRecentDataset(state: DistilState, dataset: string) {
		const datasetsStr = window.localStorage.getItem('recent-datasets');
		const datasets = (datasetsStr) ? datasetsStr.split(',') : [];
		datasets.unshift(dataset);
		window.localStorage.setItem('recent-datasets', datasets.join(','));
	}
};

