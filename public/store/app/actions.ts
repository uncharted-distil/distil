import axios from 'axios';
import { AppState } from './index';
import { DistilState } from '../store';
import { ActionContext } from 'vuex';
import { mutations } from './module';

export type AppContext = ActionContext<AppState, DistilState>;

export const actions = {

	abort(context: AppContext) {
		return axios.get('/distil/abort')
			.then(() => {
				console.warn('User initiated session abort');
				mutations.setAborted(context);
			})
			.catch(error => {
				// NOTE: request always fails because we exit on the server
				console.warn('User initiated session abort');
				mutations.setAborted(context);
			});
	},

	exportPipeline(context: AppContext, args: { sessionId: string, pipelineId: string}) {
		return axios.get(`/distil/export/${args.sessionId}/${args.pipelineId}`)
			.then(() => {
				console.warn(`User exported pipeline ${args.pipelineId}`);
				mutations.setAborted(context);
			})
			.catch(error => {
				// check for case where target / task doesn't match the problem requests - server returns
				// a bad request staus code along with an error message
				if (error.response && error.response.status === 400) {
					return new Error(error.response.data);
				} else {
					console.warn(`User exported pipeline ${args.pipelineId}`);
					mutations.setAborted(context);
				}
			});
	}
};
