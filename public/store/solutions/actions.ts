import axios from 'axios';
import { SolutionState,
	SOLUTION_PENDING, SOLUTION_RUNNING, SOLUTION_COMPLETED, SOLUTION_ERRORED,
	REQUEST_PENDING, REQUEST_RUNNING, REQUEST_COMPLETED, REQUEST_ERRORED } from './index';
import { ActionContext } from 'vuex';
import { DistilState } from '../store';
import { mutations } from './module';
import { getWebSocketConnection } from '../../util/ws';
import { FilterParams } from '../../util/filters';

const CREATE_SOLUTIONS = 'CREATE_SOLUTIONS';
const STOP_SOLUTIONS = 'STOP_SOLUTIONS';

interface CreateSolutionRequest {
	dataset: string;
	target: string;
	task?: string;
	subTask?: string;
	timestampField?: string;
	metrics: string[];
	maxSolutions: number;
	maxTime: number;
	filters: FilterParams;
}

interface SolutionStatus {
	requestId: string;
	solutionId?: string;
	resultId?: string;
	progress: string;
	error: string;
	timestamp: number;
}

export type SolutionContext = ActionContext<SolutionState, DistilState>;

function updateCurrentSolutionResults(context: SolutionContext, req: CreateSolutionRequest, res: SolutionStatus) {
	const isRegression = context.getters.isRegression;
	const isClassification = context.getters.isClassification;

	context.dispatch('fetchResultTableData', {
		dataset: req.dataset,
		solutionId: res.solutionId,
		highlight: context.getters.getDecodedHighlight
	});
	context.dispatch('fetchPredictedSummary', {
		dataset: req.dataset,
		target: req.target,
		solutionId: res.solutionId,
		highlight: context.getters.getDecodedHighlight
	});
	context.dispatch('fetchTrainingSummaries', {
		dataset: req.dataset,
		training: context.getters.getActiveSolutionTrainingVariables,
		solutionId: res.solutionId,
		highlight: context.getters.getDecodedHighlight
	});
	context.dispatch('fetchTargetSummary', {
		dataset: req.dataset,
		target: req.target,
		solutionId: res.solutionId,
		highlight: context.getters.getDecodedHighlight
	});

	if (isRegression) {
		context.dispatch('fetchResidualsExtrema', {
			dataset: req.dataset,
			target: req.target,
			solutionId: res.solutionId
		});
		context.dispatch('fetchResidualsSummary', {
			dataset: req.dataset,
			target: req.target,
			solutionId: res.solutionId,
			highlight: context.getters.getDecodedHighlight
		});
	} else if (isClassification) {
		context.dispatch('fetchCorrectnessSummary', {
			dataset: req.dataset,
			solutionId: res.solutionId,
			highlight: context.getters.getDecodedHighlight
		});
	}
}

function updateSolutionResults(context: SolutionContext, req: CreateSolutionRequest, res: SolutionStatus) {
	const isRegression = context.getters.isRegression;
	const isClassification = context.getters.isClassification;

	// if current solutionId, pull result summaries
	context.dispatch('fetchPredictedSummary', {
		dataset: req.dataset,
		target: req.target,
		solutionId: res.solutionId,
		highlight: context.getters.getDecodedHighlight
	});

	if (isRegression) {
		context.dispatch('fetchResidualsExtrema', {
			dataset: req.dataset,
			target: req.target,
			solutionId: res.solutionId
		});
		context.dispatch('fetchResidualsSummary', {
			dataset: req.dataset,
			target: req.target,
			solutionId: res.solutionId,
			highlight: context.getters.getDecodedHighlight
		});
	} else if (isClassification) {
		context.dispatch('fetchCorrectnessSummary', {
			dataset: req.dataset,
			solutionId: res.solutionId,
			highlight: context.getters.getDecodedHighlight
		});
	}
}

function handleRequestProgress(context: SolutionContext, request: CreateSolutionRequest, response: SolutionStatus) {
	// no-op
}

function handleSolutionProgress(context: SolutionContext, request: CreateSolutionRequest, response: SolutionStatus) {
	switch (response.progress) {
		case SOLUTION_COMPLETED:
		case SOLUTION_ERRORED:
			// if current solutionId, pull results
			if (response.solutionId === context.getters.getRouteSolutionId) {
				// current solutionId is selected
				updateCurrentSolutionResults(context, request, response);
			} else {
				// current solutionId is NOT selected
				updateSolutionResults(context, request, response);
			}
			break;
	}
}

function isRequestResponse(response: SolutionStatus) {
	const progress = response.progress;
	return progress === REQUEST_PENDING ||
		progress === REQUEST_RUNNING ||
		progress === REQUEST_COMPLETED ||
		progress === REQUEST_ERRORED;
}

function isSolutionResponse(response: SolutionStatus) {
	const progress = response.progress;
	return progress === SOLUTION_PENDING ||
		progress === SOLUTION_RUNNING ||
		progress === SOLUTION_COMPLETED ||
		progress === SOLUTION_ERRORED;
}

function handleProgress(context: SolutionContext, request: CreateSolutionRequest, response: SolutionStatus) {

	if (isRequestResponse(response)) {
		// request
		console.log(`Progress for request ${response.requestId} updated to ${response.progress}`);
	} else if (isSolutionResponse(response)) {
		// solution
		console.log(`Progress for solution ${response.solutionId} updated to ${response.progress}`);
	}

	context.dispatch('fetchSolutionRequests', {
		dataset: request.dataset,
		target: request.target,
		solutionId: response.solutionId,
	}).then(() => {
		// handle response
		if (isRequestResponse(response)) {
			// request
			handleRequestProgress(context, request, response);
		} else if (isSolutionResponse(response)) {
			// solution
			handleSolutionProgress(context, request, response);
		}
	});
}

export const actions = {

	fetchSolutionRequests(context: SolutionContext, args: { dataset?: string, target?: string, solutionId?: string }) {
		if (!args.dataset) {
			args.dataset = null;
		}
		if (!args.target) {
			args.target = null;
		}
		if (!args.solutionId) {
			args.solutionId = null;
		}

		return axios.get(`/distil/solutions/${args.dataset}/${args.target}/${args.solutionId}`)
			.then(response => {
				if (!response.data) {
					return;
				}
				const requests = response.data;
				requests.forEach(request => {
					// update solution
					mutations.updateSolutionRequests(context, request);
				});
			})
			.catch(error => {
				console.error(error);
			});
	},

	createSolutionRequest(context: any, request: CreateSolutionRequest) {
		return new Promise((resolve, reject) => {

			const conn = getWebSocketConnection();

			let receivedFirstSolution = false;
			let receivedFirstResponse = false;

			const stream = conn.stream(response => {

				// log any error
				if (response.error) {
					console.error(response.error);
				}

				// handle request / solution progress
				if (response.progress) {
					handleProgress(context, request, response);
				}

				if (response.requestId && !receivedFirstResponse) {
					receivedFirstResponse = true;
					// add the request stream
					mutations.addRequestStream(context, { requestId: response.requestId, stream: stream });
				}

				if (response.solutionId && !receivedFirstSolution) {
					receivedFirstSolution = true;
					// resolve
					resolve(response);
				}

				// close stream on complete
				if (response.complete) {
					console.log('Solution request has completed, closing stream');
					// remove request stream
					if (receivedFirstResponse) {
						mutations.removeRequestStream(context, { requestId: response.requestId });
					}
					// check for failure to generate solutions
					if (!receivedFirstSolution) {
						reject(new Error('No valid solutions found'));
					}
					// close stream
					stream.close();
				}

			});

			console.log('Sending create solutions request:', request);

			// send create solutions request
			stream.send({
				type: CREATE_SOLUTIONS,
				dataset: request.dataset,
				target: request.target,
				task: request.task,
				timestampField: request.timestampField,
				subTask: request.subTask,
				metrics: request.metrics,
				maxSolutions: request.maxSolutions,
				maxTime: request.maxTime,
				filters: request.filters
			});
		});
	},

	stopSolutionRequest(context: any, args: { requestId: string }) {
		const streams = context.getters.getRequestStreams;
		const stream = streams[args.requestId];
		if (!stream) {
			console.warn(`No request stream found for requestId: ${args.requestId}`);
			return;
		}
		stream.send({
			type: STOP_SOLUTIONS,
			requestId: args.requestId
		});
	},
};
