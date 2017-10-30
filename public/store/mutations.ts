import _ from 'lodash';
import Vue from 'vue';
import { MutationTree } from 'vuex';
import { DistilState, Session, PipelineInfo } from './index';

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

	// adds a running pipeline or replaces an existing one if the ids match
	addRunningPipeline(state: DistilState, pipelineData: PipelineInfo) {
		if (!_.has(state.runningPipelines, pipelineData.requestId)) {
			Vue.set(state.runningPipelines, pipelineData.requestId, {});
		}
		Vue.set(state.runningPipelines[pipelineData.requestId], pipelineData.pipelineId, pipelineData);
	},

	// removes a running pipeline
	removeRunningPipeline(state: DistilState, args: { requestId: string, pipelineId: string }) {
		if (_.has(state.runningPipelines, args.requestId)) {
			// delete the pipeline from the request
			if (_.has(state.runningPipelines[args.requestId], args.pipelineId)) {
				Vue.delete(state.runningPipelines[args.requestId], args.pipelineId);
				// delete the request if empty
				if (_.size(state.runningPipelines[args.requestId]) === 0) {
					Vue.delete(state.runningPipelines, args.requestId);
				}
				return true;
			}
		}
		return false;
	},

	// adds a completed pipeline or replaces an existing one if the ids match
	addCompletedPipeline(state: DistilState, pipelineData: PipelineInfo) {
		if (!_.has(state.completedPipelines, pipelineData.requestId)) {
			Vue.set(state.completedPipelines, pipelineData.requestId, {});
		}
		Vue.set(state.completedPipelines[pipelineData.requestId], pipelineData.pipelineId, pipelineData);
	},

	// removes a completed pipeline
	removeCompletedPipeline(state: DistilState, args: { requestId: string, pipelineId: string }) {
		if (_.has(state.runningPipelines, args.requestId)) {
			// delete the pipeline from the request
			if (_.has(state.completedPipelines[args.requestId], args.pipelineId)) {
				// delete the request if empty
				Vue.delete(state.completedPipelines[args.requestId], args.pipelineId);
				if (_.size(state.completedPipelines[args.requestId]) === 0) {
					Vue.delete(state.completedPipelines, args.requestId);
				}
				return true;
			}
		}
		return false;
	},

	addRecentDataset(state: DistilState, dataset: string) {
		const datasetsStr = window.localStorage.getItem('recent-datasets');
		const datasets = (datasetsStr) ? datasetsStr.split(',') : [];
		datasets.unshift(dataset);
		window.localStorage.setItem('recent-datasets', datasets.join(','));
	}
};

