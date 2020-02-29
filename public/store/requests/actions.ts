import axios from "axios";
import {
  RequestState,
  SOLUTION_PENDING,
  SOLUTION_COMPLETED,
  SOLUTION_ERRORED,
  REQUEST_PENDING,
  REQUEST_RUNNING,
  REQUEST_COMPLETED,
  REQUEST_ERRORED,
  SOLUTION_FITTING,
  SOLUTION_PRODUCING,
  SOLUTION_SCORING,
  SearchRequest,
  Solution
} from "./index";
import { ActionContext } from "vuex";
import store, { DistilState } from "../store";
import { mutations } from "./module";
import { getWebSocketConnection, getStreamById } from "../../util/ws";
import { FilterParams } from "../../util/filters";
import { actions as resultsActions } from "../results/module";
import { getters as routeGetters } from "../route/module";
import { TaskTypes, SummaryMode } from "../dataset";

const CREATE_SOLUTIONS = "CREATE_SOLUTIONS";
const STOP_SOLUTIONS = "STOP_SOLUTIONS";

// Search request message used in web socket context
interface SearchRequestMsg {
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

export type SolutionContext = ActionContext<RequestState, DistilState>;

function updateCurrentSolutionResults(
  context: SolutionContext,
  req: SearchRequestMsg,
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

  const varModes = context.getters.getDecodedVarModes;

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
      target: req.target,
      solutionId: res.solutionId,
      highlight: context.getters.getDecodedHighlight
    });
  }
}

function updateSolutionResults(
  context: SolutionContext,
  req: SearchRequestMsg,
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
      target: req.target,
      solutionId: res.solutionId,
      highlight: context.getters.getDecodedHighlight
    });
  }
}

function handleRequestProgress(
  context: SolutionContext,
  request: SearchRequestMsg,
  response: SolutionStatusMsg
) {
  // no-op
}

function handleSolutionProgress(
  context: SolutionContext,
  request: SearchRequestMsg,
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

function isRequestResponse(response: SolutionStatusMsg) {
  const progress = response.progress;
  return (
    progress === REQUEST_PENDING ||
    progress === REQUEST_RUNNING ||
    progress === REQUEST_COMPLETED ||
    progress === REQUEST_ERRORED
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
  context: SolutionContext,
  request: SearchRequestMsg,
  response: SolutionStatusMsg
) {
  if (isRequestResponse(response)) {
    // request
    console.log(
      `Progress for request ${response.requestId} updated to ${response.progress}`
    );
    await actions.fetchSearchRequest(context, {
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

// parse returned server data into a solution that can be added to the index
function parseSolutionResponse(responseData: any): Solution {
  return {
    requestId: responseData.requestId,
    solutionId: responseData.solutionId,
    fittedSolutionId: responseData.fittedSolutionId,
    resultId: responseData.resultId,
    dataset: responseData.dataset,
    feature: responseData.feature,
    scores: responseData.scores,
    timestamp: responseData.timestamp,
    progress: responseData.progress,
    features: responseData.features,
    filters: responseData.filters,
    predictedKey: responseData.predictedKey,
    errorKey: responseData.errorKey,
    isBad: false
  };
}

// parse returned server data into a solution request that can be added to the index
function parseRequestResponse(responseData: any): SearchRequest {
  return {
    requestId: responseData.requestId,
    dataset: responseData.dataset,
    feature: responseData.feature,
    features: responseData.features,
    filters: responseData.filters,
    timestamp: responseData.timestamp,
    progress: responseData.progress
  };
}

export const actions = {
  async fetchSearchRequests(
    context: SolutionContext,
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
      const requestResponse = await axios.get(
        `/distil/solution-requests/${args.dataset}/${args.target}`
      );
      const requests = requestResponse.data;
      for (const request of requests) {
        // update request data
        const searchRequest = parseRequestResponse(request);
        mutations.updateSearchRequests(context, searchRequest);
      }
    } catch (error) {
      console.error(error);
    }
  },

  async fetchSearchRequest(
    context: SolutionContext,
    args: { requestId: string }
  ) {
    if (!args.requestId) {
      args.requestId = null;
    }

    try {
      // fetch and uddate the search data
      const requestResponse = await axios.get(
        `/distil/solution-request/${args.requestId}`
      );
      // update request data
      const searchRequest = parseRequestResponse(requestResponse.data);
      mutations.updateSearchRequests(context, searchRequest);
    } catch (error) {
      console.error(error);
    }
  },

  async fetchSolutions(
    context: SolutionContext,
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
      const solutionResponse = await axios.get(
        `/distil/solutions/${args.dataset}/${args.target}`
      );
      if (!solutionResponse.data) {
        return;
      }
      for (const solution of solutionResponse.data) {
        const searchResult = parseSolutionResponse(solution);
        mutations.updateSolutions(context, searchResult);
      }
    } catch (error) {
      console.error(error);
    }
  },

  async fetchSolution(context: SolutionContext, args: { solutionId: string }) {
    try {
      // fetch update the solution data
      const solutionResponse = await axios.get(
        `/distil/solution/${args.solutionId}`
      );
      if (!solutionResponse.data) {
        return;
      }
      const searchResult = parseSolutionResponse(solutionResponse.data);
      mutations.updateSolutions(context, searchResult);
    } catch (error) {
      console.error(error);
    }
  },

  createSearchRequest(context: any, request: SearchRequestMsg) {
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
          // close stream
          stream.close();
          // close the socket
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

  stopSearchRequest(context: any, args: { requestId: string }) {
    const stream = getStreamById(args.requestId);
    if (!stream) {
      console.warn(`No request stream found for requestId: ${args.requestId}`);
      return;
    }
    stream.send({
      type: STOP_SOLUTIONS,
      requestId: args.requestId
    });
  }
};
