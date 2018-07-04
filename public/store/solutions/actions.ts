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
	task: string;
	maxSolutions: number;
	maxTime: number;
	metrics: string[];
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
		highlightRoot: context.getters.getDecodedHighlightRoot
	});
	context.dispatch('fetchPredictedSummary', {
		dataset: req.dataset,
		target: req.target,
		solutionId: res.solutionId
	});
	context.dispatch('fetchTrainingSummaries', {
		dataset: req.dataset,
		training: context.getters.getResultsPaginatedVariables,
		solutionId: res.solutionId,
	});
	context.dispatch('fetchTargetSummary', {
		dataset: req.dataset,
		target: req.target,
		solutionId: res.solutionId
	});
	context.dispatch('fetchResultHighlightValues', {
		dataset: req.dataset,
		target: req.target,
		training: context.getters.getResultsPaginatedVariables,
		highlightRoot: context.getters.getDecodedHighlightRoot,
		solutionId: res.solutionId,
		requestIds: context.getters.getSolutionRequestIds,
		includeCorrectness: isClassification,
		includeResidual: isRegression
	});

	if (isRegression) {
		context.dispatch('fetchResidualsExtrema', {
			dataset: req.dataset,
			target: req.target
		});
		context.dispatch('fetchResidualsSummary', {
			dataset: req.dataset,
			target: req.target,
			solutionId: res.solutionId
		});
	} else if (isClassification) {
		context.dispatch('fetchCorrectnessSummary', {
			dataset: req.dataset,
			solutionId: res.solutionId
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
		solutionId: res.solutionId
	});
	context.dispatch('fetchResultHighlightValues', {
		dataset: req.dataset,
		target: req.target,
		training: context.getters.getResultsPaginatedVariables,
		highlightRoot: context.getters.getDecodedHighlightRoot,
		solutionId: res.solutionId,
		requestIds: context.getters.getSolutionRequestIds,
		includeCorrectness: isClassification,
		includeResidual: isRegression
	});

	if (isRegression) {
		context.dispatch('fetchResidualsExtrema', {
			dataset: req.dataset,
			target: req.target
		});
		context.dispatch('fetchResidualsSummary', {
			dataset: req.dataset,
			target: req.target,
			solutionId: res.solutionId
		});
	} else if (isClassification) {
		context.dispatch('fetchCorrectnessSummary', {
			dataset: req.dataset,
			solutionId: res.solutionId
		});
	}
}

function handleRequestProgress(context: SolutionContext, request: CreateSolutionRequest, response: SolutionStatus) {

	console.log(`Progress for request ${response.requestId} updated to ${response.progress}`);

	switch (response.progress) {
		case REQUEST_PENDING:
		case REQUEST_RUNNING:
		case REQUEST_COMPLETED:
		case REQUEST_ERRORED:
			break;
	}
}

function handleSolutionProgress(context: SolutionContext, request: CreateSolutionRequest, response: SolutionStatus) {

	console.log(`Progress for solution ${response.solutionId} updated to ${response.progress}`);

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

export const actions = {

	fetchSolutions(context: SolutionContext, args: { dataset?: string, target?: string, solutionId?: string }) {
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

			let receivedFirstResponse = false;

			const stream = conn.stream(response => {

				if (response.error) {
					console.error(response.error);
					return;
				}

				// close stream on complete
				if (response.complete) {
					console.log('Solution request has completed, closing stream');
					stream.close();
					mutations.removeRequestStream(context, { requestId: response.requestId });
				}

				// pull updated solution info

				context.dispatch('fetchSolutions', {
					dataset: request.dataset,
					target: request.target,
					solutionId: response.solutionId,
				}).then(() => {
					// handle response
					switch (response.progress) {
						case REQUEST_PENDING:
						case REQUEST_RUNNING:
						case REQUEST_COMPLETED:
						case REQUEST_ERRORED:
							handleRequestProgress(context, request, response);
							break;
						case SOLUTION_PENDING:
						case SOLUTION_RUNNING:
						case SOLUTION_COMPLETED:
						case SOLUTION_ERRORED:
							// resolve promise on first solution response
							if (!receivedFirstResponse) {
								receivedFirstResponse = true;
								// add the request stream
								mutations.addRequestStream(context, { requestId: response.requestId, stream: stream });
								resolve(response);
							}
							handleSolutionProgress(context, request, response);
							break;
					}
				});
			});

			// send create solutions request
			stream.send({
				type: CREATE_SOLUTIONS,
				dataset: request.dataset,
				target: request.target,
				task: request.task,
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
}
