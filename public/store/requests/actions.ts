import axios from "axios";
import {
  RequestState,
  SOLUTION_PENDING,
  SOLUTION_COMPLETED,
  SOLUTION_ERRORED,
  SOLUTION_REQUEST_PENDING,
  SOLUTION_REQUEST_RUNNING,
  SOLUTION_REQUEST_COMPLETED,
  SOLUTION_REQUEST_ERRORED,
  SOLUTION_FITTING,
  SOLUTION_PRODUCING,
  SOLUTION_SCORING,
  SolutionRequest,
  Solution,
  PREDICT_COMPLETED,
  PREDICT_ERRORED,
  Predictions
} from "./index";
import { ActionContext } from "vuex";
import store, { DistilState } from "../store";
import { mutations } from "./module";
import { getWebSocketConnection, getStreamById } from "../../util/ws";
import { FilterParams } from "../../util/filters";
import { actions as resultsActions } from "../results/module";
import { actions as predictActions } from "../predictions/module";
import { getters as routeGetters } from "../route/module";
import { TaskTypes, SummaryMode } from "../dataset";
import { validateArgs } from "../../util/data";

const CREATE_SOLUTIONS = "CREATE_SOLUTIONS";
const STOP_SOLUTIONS = "STOP_SOLUTIONS";
const CREATE_PREDICTIONS = "PREDICT";
const STOP_PREDICTIONS = "STOP_PREDICTIONS";

// Message definitions for the websocket.  These are only for communication with the
// server while the requests are running, and are not stored in the index.

// Search request message used in web socket context
interface SolutionRequestMsg {
  dataset: string;
  target: string;
  metrics: string[];
  maxSolutions: number;
  maxTime: number;
  filters: FilterParams;
}

// Solution status message used in web socket context
interface SolutionStatusMsg {
  requestId: string;
  solutionId?: string;
  resultId?: string;
  progress: string;
  error: string;
  timestamp: number;
}

interface PredictRequestMsg {
  datasetId: string;
  dataset: string; // base64 encoded version of dataset
  fittedSolutionId: string;
  target: string;
  targetType: string;
}

// Prediction status.
interface PredictStatusMsg {
  solutionId: string;
  resultId: string;
  produceRequestId: string;
  progress: string;
  error: string;
  timestamp: number;
}

export type RequestContext = ActionContext<RequestState, DistilState>;

function updateCurrentSolutionResults(
  context: RequestContext,
  req: SolutionRequestMsg,
  res: SolutionStatusMsg
) {
  const isRegression = routeGetters
    .getRouteTask(store)
    .includes(TaskTypes.REGRESSION);
  const isClassification = routeGetters
    .getRouteTask(store)
    .includes(TaskTypes.CLASSIFICATION);
  const isForecasting = routeGetters
    .getRouteTask(store)
    .includes(TaskTypes.FORECASTING);

  const varModes: Map<string, SummaryMode> = context.getters.getDecodedVarModes;

  resultsActions.fetchResultTableData(store, {
    dataset: req.dataset,
    solutionId: res.solutionId,
    highlight: context.getters.getDecodedHighlight
  });
  resultsActions.fetchPredictedSummary(store, {
    dataset: req.dataset,
    target: req.target,
    solutionId: res.solutionId,
    highlight: context.getters.getDecodedHighlight,
    varMode: varModes.has(req.target)
      ? varModes.get(req.target)
      : SummaryMode.Default
  });
  resultsActions.fetchTrainingSummaries(store, {
    dataset: req.dataset,
    training: context.getters.getActiveSolutionTrainingVariables,
    solutionId: res.solutionId,
    highlight: context.getters.getDecodedHighlight,
    varModes: varModes
  });
  resultsActions.fetchTargetSummary(store, {
    dataset: req.dataset,
    target: req.target,
    solutionId: res.solutionId,
    highlight: context.getters.getDecodedHighlight,
    varMode: varModes.has(req.target)
      ? varModes.get(req.target)
      : SummaryMode.Default
  });

  if (isRegression || isForecasting) {
    resultsActions.fetchResidualsExtrema(store, {
      dataset: req.dataset,
      target: req.target,
      solutionId: res.solutionId
    });
    resultsActions.fetchResidualsSummary(store, {
      dataset: req.dataset,
      target: req.target,
      solutionId: res.solutionId,
      highlight: context.getters.getDecodedHighlight,
      varMode: varModes.has(req.target)
        ? varModes.get(req.target)
        : SummaryMode.Default
    });
  } else if (isClassification) {
    resultsActions.fetchCorrectnessSummary(store, {
      dataset: req.dataset,
      solutionId: res.solutionId,
      highlight: context.getters.getDecodedHighlight,
      varMode: varModes.has(req.target)
        ? varModes.get(req.target)
        : SummaryMode.Default
    });
  }
}

// Updates an in-progress prediction request handled over the web socket.
function updateCurrentPredictResults(
  context: RequestContext,
  req: PredictRequestMsg,
  res: PredictStatusMsg
) {
  const varModes = context.getters.getDecodedVarModes;

  predictActions.fetchPredictionTableData(store, {
    dataset: req.datasetId,
    highlight: context.getters.getDecodedHighlight,
    produceRequestId: res.produceRequestId
  });

  predictActions.fetchPredictedSummary(store, {
    highlight: context.getters.getDecodedHighlight,
    varMode: varModes.has(req.target)
      ? varModes.get(req.target)
      : SummaryMode.Default,
    produceRequestId: res.produceRequestId
  });

  predictActions.fetchTrainingSummaries(store, {
    dataset: req.datasetId,
    training: context.getters.getActiveSolutionTrainingVariables,
    highlight: context.getters.getDecodedHighlight,
    varModes: varModes,
    produceRequestId: res.produceRequestId
  });
}

function updateSolutionResults(
  context: RequestContext,
  req: SolutionRequestMsg,
  res: SolutionStatusMsg
) {
  const taskArgs = routeGetters.getRouteTask(store);
  const isRegression = taskArgs && taskArgs.includes(TaskTypes.REGRESSION);
  const isClassification =
    taskArgs && taskArgs.includes(TaskTypes.CLASSIFICATION);
  const isForecasting = taskArgs && taskArgs.includes(TaskTypes.FORECASTING);

  const varModes = context.getters.getDecodedVarModes;

  // if current solutionId, pull result summaries
  resultsActions.fetchPredictedSummary(store, {
    dataset: req.dataset,
    target: req.target,
    solutionId: res.solutionId,
    highlight: context.getters.getDecodedHighlight,
    varMode: varModes.has(req.target)
      ? varModes.get(req.target)
      : SummaryMode.Default
  });

  if (isRegression || isForecasting) {
    resultsActions.fetchResidualsExtrema(store, {
      dataset: req.dataset,
      target: req.target,
      solutionId: res.solutionId
    });
    resultsActions.fetchResidualsSummary(store, {
      dataset: req.dataset,
      target: req.target,
      solutionId: res.solutionId,
      highlight: context.getters.getDecodedHighlight,
      varMode: varModes.has(req.target)
        ? varModes.get(req.target)
        : SummaryMode.Default
    });
  } else if (isClassification) {
    resultsActions.fetchCorrectnessSummary(store, {
      dataset: req.dataset,
      solutionId: res.solutionId,
      highlight: context.getters.getDecodedHighlight,
      varMode: varModes.has(req.target)
        ? varModes.get(req.target)
        : SummaryMode.Default
    });
  }
}

function handleRequestProgress(
  context: RequestContext,
  request: SolutionRequestMsg,
  response: SolutionStatusMsg
) {
  // no-op
}

function handleSolutionProgress(
  context: RequestContext,
  request: SolutionRequestMsg,
  response: SolutionStatusMsg
) {
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

function isSolutionRequestResponse(response: SolutionStatusMsg) {
  const progress = response.progress;
  return (
    progress === SOLUTION_REQUEST_PENDING ||
    progress === SOLUTION_REQUEST_RUNNING ||
    progress === SOLUTION_REQUEST_COMPLETED ||
    progress === SOLUTION_REQUEST_ERRORED
  );
}

function isSolutionResponse(response: SolutionStatusMsg) {
  const progress = response.progress;
  return (
    progress === SOLUTION_PENDING ||
    progress === SOLUTION_FITTING ||
    progress === SOLUTION_SCORING ||
    progress === SOLUTION_PRODUCING ||
    progress === SOLUTION_COMPLETED ||
    progress === SOLUTION_ERRORED
  );
}

async function handleProgress(
  context: RequestContext,
  request: SolutionRequestMsg,
  response: SolutionStatusMsg
) {
  if (isSolutionRequestResponse(response)) {
    // request
    console.log(
      `Progress for request ${response.requestId} updated to ${response.progress}`
    );
    await actions.fetchSolutionRequest(context, {
      requestId: response.requestId
    });
    handleRequestProgress(context, request, response);
  } else if (isSolutionResponse(response)) {
    // solution
    console.log(
      `Progress for solution ${response.solutionId} updated to ${response.progress}`
    );
    await actions.fetchSolution(context, {
      solutionId: response.solutionId
    });
    handleSolutionProgress(context, request, response);
  }
}

async function handlePredictProgress(
  context: RequestContext,
  request: PredictRequestMsg,
  response: PredictStatusMsg
) {
  // request
  console.log(
    `Progress for request ${response.resultId} updated to ${response.progress}`
  );
  switch (response.progress) {
    case PREDICT_COMPLETED:
    case PREDICT_ERRORED:
      // no waiting for data here - we get single response back when the prediction is complete
      await actions.fetchPrediction(context, {
        requestId: response.produceRequestId
      });
      updateCurrentPredictResults(context, request, response);
      break;
  }
}

export const actions = {
  async fetchSolutionRequests(
    context: RequestContext,
    args: { dataset?: string; target?: string }
  ) {
    if (!args.dataset) {
      args.dataset = null;
    }
    if (!args.target) {
      args.target = null;
    }

    try {
      // fetch and uddate the search data
      const requestResponse = await axios.get<SolutionRequest[]>(
        `/distil/solution-requests/${args.dataset}/${args.target}`
      );
      const requests = requestResponse.data;
      for (const request of requests) {
        // update request data
        mutations.updateSolutionRequests(context, request);
      }
    } catch (error) {
      console.error(error);
    }
  },

  async fetchSolutionRequest(
    context: RequestContext,
    args: { requestId: string }
  ) {
    if (!args.requestId) {
      args.requestId = null;
    }

    try {
      // fetch and uddate the search data
      const requestResponse = await axios.get<SolutionRequest>(
        `/distil/solution-request/${args.requestId}`
      );
      // update request data
      mutations.updateSolutionRequests(context, requestResponse.data);
    } catch (error) {
      console.error(error);
    }
  },

  async fetchSolutions(
    context: RequestContext,
    args: { dataset?: string; target?: string }
  ) {
    if (!args.dataset) {
      args.dataset = null;
    }
    if (!args.target) {
      args.target = null;
    }

    try {
      // fetch update the solution data
      const solutionResponse = await axios.get<Solution[]>(
        `/distil/solutions/${args.dataset}/${args.target}`
      );
      if (!solutionResponse.data) {
        return;
      }
      for (const solution of solutionResponse.data) {
        mutations.updateSolutions(context, solution);
      }
    } catch (error) {
      console.error(error);
    }
  },

  async fetchSolution(context: RequestContext, args: { solutionId: string }) {
    try {
      // fetch update the solution data
      const solutionResponse = await axios.get<Solution>(
        `/distil/solution/${args.solutionId}`
      );
      if (!solutionResponse.data) {
        return;
      }
      mutations.updateSolutions(context, solutionResponse.data);
    } catch (error) {
      console.error(error);
    }
  },

  // Opens up a websocket and initiates the model search.  Updates are returned
  // asynchronously by the server until the request completes.
  createSolutionRequest(context: RequestContext, request: SolutionRequestMsg) {
    return new Promise((resolve, reject) => {
      const conn = getWebSocketConnection();

      let receivedFirstSolution = false;

      const stream = conn.stream(response => {
        // log any error
        if (response.error) {
          console.error(response.error);
        }

        // handle request / solution progress
        if (response.progress) {
          handleProgress(context, request, response);
        }

        if (response.solutionId && !receivedFirstSolution) {
          receivedFirstSolution = true;
          // resolve
          resolve(response);
        }

        // close stream on complete
        if (response.complete) {
          console.log("Solution request has completed, closing stream");
          // check for failure to generate solutions
          if (!receivedFirstSolution) {
            reject(new Error("No valid solutions found"));
          }
          // close streampredict
          conn.close();
        }
      });

      console.log("Sending create solutions request:", request);

      // send create solutions request
      stream.send({
        type: CREATE_SOLUTIONS,
        dataset: request.dataset,
        target: request.target,
        metrics: request.metrics,
        maxSolutions: request.maxSolutions,
        maxTime: request.maxTime,
        filters: request.filters
      });
    });
  },

  stopSolutionRequest(context: any, args: { requestId: string }) {
    const stream = getStreamById(args.requestId);
    if (!stream) {
      console.warn(`No request stream found for requestId: ${args.requestId}`);
      return;
    }
    stream.send({
      type: STOP_SOLUTIONS,
      requestId: args.requestId
    });
  },

  // Opens up a websocket and initiates a prediction request.  Updates are returned until
  // the predctions finish generating.
  createPredictRequest(context: RequestContext, request: PredictRequestMsg) {
    let receivedUpdate = false;

    return new Promise((resolve, reject) => {
      const conn = getWebSocketConnection();
      const stream = conn.stream(response => {
        // log any error
        if (response.error) {
          console.error(response.error);
          resolve(response);
        }

        // handle prediction request progress - this is currently a one-shot operation, rather than
        // one that streams in progress updates like solution processing.  We need to have the prediction
        // data ready in order to move on, so we don't flag the resolve until handling of the predict complete
        // message is finished
        if (response.progress) {
          handlePredictProgress(context, request, response).then(() => {
            // resolve the promise on the first update
            if (!receivedUpdate) {
              receivedUpdate = true;
              resolve(response);
            }
          });
        }

        // close stream on complete
        if (response.complete) {
          console.log("Prediction request has completed, closing stream");
          // close stream
          stream.close();
          // close the socket
          conn.close();
        }
      });

      console.log("Sending predict request:", request);

      // send create solutions request
      stream.send({
        type: CREATE_PREDICTIONS,
        fittedSolutionId: request.fittedSolutionId,
        datasetId: request.datasetId,
        dataset: request.dataset,
        targetType: request.targetType
      });
    });
  },

  // notifies server that prediction request should be halted
  stopPredictRequest(context: RequestContext, args: { requestId: string }) {
    const stream = getStreamById(args.requestId);
    if (!stream) {
      console.warn(`No request stream found for requestId: ${args.requestId}`);
      return;
    }
    stream.send({
      type: STOP_PREDICTIONS,
      requestId: args.requestId
    });
  },

  // fetches all predictions for a given fitted solution
  async fetchPredictions(
    context: RequestContext,
    args: { fittedSolutionId: string }
  ) {
    args.fittedSolutionId = args.fittedSolutionId || "";
    try {
      // fetch and uddate the search data
      const predictionsResponse = await axios.get<Predictions[]>(
        `/distil/predictions/${args.fittedSolutionId}`
      );
      for (const predictions of predictionsResponse.data) {
        mutations.updatePredictions(context, predictions);
      }
    } catch (error) {
      console.error(error);
    }
  },

  // fetches a specific prediction by request ID
  async fetchPrediction(context: RequestContext, args: { requestId: string }) {
    if (!validateArgs(args, ["requestId"])) {
      return;
    }

    try {
      // fetch and uddate the search data
      const requestResponse = await axios.get<Predictions>(
        `/distil/prediction/${args.requestId}`
      );
      // update request data
      mutations.updatePredictions(context, requestResponse.data);
    } catch (error) {
      console.error(error);
    }
  }
};
