import axios from 'axios';
import { AppState } from './index';
import { DistilState } from '../store';
import { ActionContext } from 'vuex';

export type AppContext = ActionContext<AppState, DistilState>;

export const actions = {

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
