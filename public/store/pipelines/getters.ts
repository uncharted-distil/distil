import _ from 'lodash';
import { Variable } from '../data/index';
import { PipelineState, PipelineInfo, PIPELINE_RUNNING, PIPELINE_UPDATED, PIPELINE_COMPLETED } from './index';
import { Dictionary } from '../../util/dict';
import localStorage from 'store';

function sortPipelines(a: PipelineInfo, b: PipelineInfo): number {
	if (a.pipelineId < b.pipelineId) {
		return -1
	}
	if (a.pipelineId > b.pipelineId) {
		return 1
	}
	return 0;
}

export const getters = {

	getPipelineSessionID(state: PipelineState): string {
		if (!state.sessionID) {
			const id = localStorage.get('pipeline-session-id');
			if (id) {
				console.log(`Loading session id ${id} from localStorage`);
			}
			return id;
		}
		return state.sessionID;
	},

	hasActiveSession(state: PipelineState): boolean {
		return state.sessionIsActive;
	},

	// Returns a dictionary of dictionaries, where the first key is the pipeline create request ID, and the second
	// key is the pipeline ID.
	getRunningPipelines(state: PipelineState): PipelineInfo[] {
		return state.pipelineRequests.filter(pipeline => pipeline.progress === PIPELINE_RUNNING).sort(sortPipelines);
	},

	// Returns a dictionary of dictionaries, where the first key is the pipeline create request ID, and the second
	// key is the pipeline ID.
	getCompletedPipelines(state: PipelineState): PipelineInfo[] {
		return state.pipelineRequests.filter(pipeline => pipeline.progress === PIPELINE_UPDATED || pipeline.progress === PIPELINE_COMPLETED).sort(sortPipelines);
	},

	getPipelines(state: PipelineState): PipelineInfo[] {
		return Array.from(state.pipelineRequests).slice().sort(sortPipelines);
	},

	getPipelineRequestIds(state: PipelineState): string[] {
		const ids = [];
		state.pipelineRequests.forEach(pipeline => {
			if (ids.indexOf(pipeline.requestId) === -1) {
				ids.push(pipeline.requestId);
			}
		});
		return ids;
	},

	getActivePipeline(state: PipelineState, getters: any): PipelineInfo {
		const pipelineId = getters.getRoutePipelineId;
		return _.find(state.pipelineRequests, pipeline => pipeline.pipelineId === pipelineId);
	},

	getActivePipelineTrainingMap(state: PipelineState, getters: any): Dictionary<boolean> {
		const activePipeline = getters.getActivePipeline;
		if (!activePipeline || !activePipeline.features) {
			return {};
		}
		const training = activePipeline.features.filter(f => f.featureType === 'train').map(f => f.featureName);
		const trainingMap = {};
		training.forEach(t => {
			trainingMap[t] = true;
		});
		return trainingMap;
	},

	getActivePipelineVariables(state: PipelineState, getters: any): Variable[] {
		const trainingMap = getters.getActivePipelineTrainingMap;
		const target = getters.getRouteTargetVariable;
		const variables = getters.getVariables;
		return variables.filter(variable => trainingMap[variable.name] || variable.name === target);
	}


}
