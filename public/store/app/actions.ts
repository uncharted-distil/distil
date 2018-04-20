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

	exportPipeline(context: AppContext, args: { pipelineId: string}) {
		return axios.get(`/distil/export/${args.pipelineId}`)
			.then(() => {
				console.warn(`User exported pipeline ${args.pipelineId}`);
				mutations.setAborted(context);
			})
			.catch(error => {
				if (error.response) {
					return new Error(error.response.data);
				} else {
					// NOTE: request always fails because we exit on the server
					console.warn(`User exported pipeline ${args.pipelineId}`);
					mutations.setAborted(context);
				}
			});
	},

	fetchVersion(context: AppContext) {
		return axios.get(`/distil/version`)
			.then(response => {
				mutations.setVersionNumber(context, response.data.version);
				mutations.setVersionTimestamp(context, response.data.timestamp);
			})
			.catch((err: string) => {
				console.warn(err);
			});
	}
};
