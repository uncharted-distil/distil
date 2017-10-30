import axios from 'axios';
import { DistilState } from './index';
import { ActionTree } from 'vuex';
import Connection from '../util/ws';

export const actions: ActionTree<DistilState, any> = {

	// starts a pipeline session.
	getPipelineSession(context: any, args: { sessionId: string } ) {
		const sessionId = args.sessionId;
		const conn = context.getters.getWebSocketConnection() as Connection;
		return conn.send({
			type: 'GET_SESSION',
			session: sessionId
		}).then(res => {
			if (sessionId && res.created) {
				console.warn('previous session', sessionId, 'could not be resumed, new session created');
			}
			context.commit('setPipelineSession', {
				id: res.session,
				uuids: res.uuids
			});
		}).catch((err: string) => {
			console.warn(err);
		});
	},

	// end a pipeline session.
	endPipelineSession(context: any, args: { sessionId: string }) {
		const sessionId = args.sessionId;
		const conn = context.getters.getWebSocketConnection();
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

	exportPipeline(context: any, args: { sessionId: string, pipelineId: string}) {
		return axios.get(`/distil/export/${args.sessionId}/${args.pipelineId}`)
		.then(() => {
			console.warn(`User exported pipeline ${args.pipelineId}`);
		})
		.catch(error => {
			console.error(`Failed to export with error ${error}`);
		});
	},

	addRecentDataset(context: any, dataset: string) {
		context.commit('addRecentDataset', dataset);
	}
};


