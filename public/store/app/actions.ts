import axios from 'axios';
import { AppState } from './index';
import { DistilState } from '../store';
import { ActionContext } from 'vuex';
import { mutations } from './module';
import { FilterParams } from '../../util/filters';

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

	exportSolution(context: AppContext, args: { solutionId: string}) {
		return axios.get(`/distil/export/${args.solutionId}`)
			.then(() => {
				console.warn(`User exported solution ${args.solutionId}`);
				mutations.setAborted(context);
			})
			.catch(error => {
				if (error.response) {
					return new Error(error.response.data);
				} else {
					// NOTE: request always fails because we exit on the server
					console.warn(`User exported solution ${args.solutionId}`);
					mutations.setAborted(context);
				}
			});
	},

	exportProblem(context: AppContext, args: { dataset: string, target: string, filterParams: FilterParams, meaningful: string }) {
		if (!args.dataset) {
			console.warn('`dataset` argument is missing');
			return null;
		}
		if (!args.target) {
			console.warn('`target` argument is missing');
			return null;
		}
		if (!args.filterParams) {
			console.warn('`filters` argument is missing');
			return null;
		}
		if (!args.meaningful) {
			console.warn('`meaningful` argument is missing');
			return null;
		}
		return axios.post(`/distil/discovery/${args.dataset}/${args.target}`, { filterParams: args.filterParams, meaningful: args.meaningful})
	},

	fetchVersion(context: AppContext) {
		return axios.get(`/distil/config`)
			.then(response => {
				mutations.setVersionNumber(context, response.data.version);
				mutations.setVersionTimestamp(context, response.data.timestamp);
				mutations.setIsDiscovery(context, response.data.discovery);
			})
			.catch((err: string) => {
				console.warn(err);
			});
	}
};
