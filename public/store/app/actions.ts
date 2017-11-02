import axios from 'axios';
import { AppState } from './index';
import { DistilState } from '../store';
import { ActionContext } from 'vuex';
import { getWebSocketConnection } from '../../util/ws';
import { mutations } from '../app/mutations';

export type AppContext = ActionContext<AppState, DistilState>;

export const actions = {

	// starts a pipeline session.
	getPipelineSession(context: AppContext, args: { sessionId: string } ) {
		const sessionId = args.sessionId;
		const conn = getWebSocketConnection();
		return conn.send({
			type: 'GET_SESSION',
			session: sessionId
		}).then(res => {
			if (sessionId && res.created) {
				console.warn('previous session', sessionId, 'could not be resumed, new session created');
			}
			mutations.setPipelineSession(context.rootState.appModule, {
				id: res.session,
				uuids: res.uuids
			});
		}).catch((err: string) => {
			console.warn(err);
		});
	},

	// end a pipeline session.
	endPipelineSession(context: AppContext, args: { sessionId: string }) {
		const sessionId = args.sessionId;
		const conn = getWebSocketConnection();
		if (!sessionId) {
			return;
		}
		return conn.send({
			type: 'END_SESSION',
			session: sessionId
		}).then(() => {
			context.commit('setPipelineSession', null);
		}).catch(err => {
			console.warn(err);
		});
	},

	abort() {
		return axios.get('/distil/abort')
		.then(() => {
			console.warn('User initiated session abort');
		})
		.catch(error => {
			console.error(`Failed to abort with error ${error}`);
		});
	},

	exportPipeline(context: AppContext, args: { sessionId: string, pipelineId: string}) {
		return axios.get(`/distil/export/${args.sessionId}/${args.pipelineId}`)
		.then(() => {
			console.warn(`User exported pipeline ${args.pipelineId}`);
		})
		.catch(error => {
			console.error(`Failed to export with error ${error}`);
		});
	}
};


