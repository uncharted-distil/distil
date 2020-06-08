import _ from "lodash";
import { ViewState } from "./index";
import { ActionContext } from "vuex";
import store, { DistilState } from "../store";
import { mutations as viewMutations, getters as viewGetters } from "./module";
import { Dictionary } from "../../util/dict";
import {
  actions as datasetActions,
  mutations as datasetMutations
} from "../dataset/module";
import {
  actions as requestActions,
  mutations as requestMutations
} from "../requests/module";
import {
  actions as resultActions,
  mutations as resultMutations
} from "../results/module";
import {
  actions as predictionActions,
  mutations as predictionMutations
} from "../predictions/module";
import {
  actions as modelActions,
  mutations as modelMutations
} from "../model/module";
import { getters as routeGetters } from "../route/module";
import { TaskTypes, SummaryMode, Variable, Highlight } from "../dataset";
import { getPredictionsById } from "../../util/predictions";

enum ParamCacheKey {
  VARIABLES = "VARIABLES",
  VARIABLE_SUMMARIES = "VARIABLE_SUMMARIES",
  VARIABLE_RANKINGS = "VARIABLE_RANKINGS",
  SEARCH_REQUESTS = "SEARCH_REQUESTS",
  SOLUTIONS = "SOLUTIONS",
  PREDICTIONS_REQUESTS = "PREDICTIONS_REQUESTS",
  PREDICTIONS = "PREDICTIONS",
  JOIN_SUGGESTIONS = "JOIN_SUGGESTIONS",
  CLUSTERS = "CLUSTERS"
}

function createCacheable(
  key: ParamCacheKey,
  func: (context: ViewContext, args: Dictionary<string>) => any
) {
  return (context: ViewContext, args: Dictionary<string>) => {
    // execute provided function if params are not cached already or changed
    const params = JSON.stringify(args);
    const cachedParams = viewGetters.getFetchParamsCache(store)[key];
    if (cachedParams !== params) {
      viewMutations.setFetchParamsCache(context, {
        key: key,
        value: params
      });
      return Promise.resolve(func(context, args));
    }
    return Promise.resolve();
  };
}

const fetchJoinSuggestions = createCacheable(
  ParamCacheKey.JOIN_SUGGESTIONS,
  (context, args) => {
    return datasetActions.fetchJoinSuggestions(store, {
      dataset: args.dataset,
      searchQuery: args.searchQuery
    });
  }
);

const fetchVariables = createCacheable(
  ParamCacheKey.VARIABLES,
  (context, args) => {
    return datasetActions.fetchVariables(store, {
      dataset: args.dataset
    });
  }
);

const fetchVariableSummaries = createCacheable(
  ParamCacheKey.VARIABLE_SUMMARIES,
  async (context, args) => {
    await fetchVariables(context, args);
    const dataset = args.dataset as string;
    const variables = context.getters.getVariables;
    const filterParams = context.getters.getDecodedSolutionRequestFilterParams;
    const highlight = context.getters.getDecodedHighlight;
    const varModes = context.getters.getDecodedVarModes;
    return Promise.all([
      datasetActions.fetchIncludedVariableSummaries(store, {
        dataset: dataset,
        variables: variables,
        filterParams: filterParams,
        highlight: highlight,
        varModes: varModes
      }),
      datasetActions.fetchExcludedVariableSummaries(store, {
        dataset: dataset,
        variables: variables,
        filterParams: filterParams,
        highlight: highlight,
        varModes: varModes
      })
    ]);
  }
);

const fetchVariableRankings = createCacheable(
  ParamCacheKey.VARIABLE_RANKINGS,
  (context, args) => {
    // if target or dataset has changed, clear previous rankings before re-fetch
    // this is needed because since user decides variable rankings to be updated, re-fetching doesn't always replace the previous data
    datasetActions.updateVariableRankings(store, {
      dataset: args.dataset,
      rankings: {}
    });
    datasetActions.fetchVariableRankings(store, {
      dataset: args.dataset,
      target: args.target
    });
  }
);

const fetchClusters = createCacheable(
  ParamCacheKey.CLUSTERS,
  (context, args) => {
    datasetActions.fetchClusters(store, {
      dataset: args.dataset
    });
  }
);

const fetchSolutionRequests = createCacheable(
  ParamCacheKey.SEARCH_REQUESTS,
  (context, args) => {
    return requestActions.fetchSolutionRequests(store, {
      dataset: args.dataset,
      target: args.target
    });
  }
);

const fetchSolutions = createCacheable(
  ParamCacheKey.SOLUTIONS,
  (context, args) => {
    return requestActions.fetchSolutions(store, {
      dataset: args.dataset,
      target: args.target
    });
  }
);

const fetchPredictions = createCacheable(
  ParamCacheKey.PREDICTIONS,
  (context, args) => {
    return requestActions.fetchPredictions(store, {
      fittedSolutionId: args.fittedSolutionId
    });
  }
);

function clearVariablesParamCache(context: ViewContext) {
  // clear variable param cache to allow re-fetching variables
  viewMutations.setFetchParamsCache(context, {
    key: ParamCacheKey.VARIABLES,
    value: undefined
  });
}

function clearVariableSummaries(context: ViewContext) {
  datasetMutations.clearVariableSummaries(store);

  viewMutations.setFetchParamsCache(context, {
    key: ParamCacheKey.VARIABLE_SUMMARIES,
    value: undefined
  });
}

export type ViewContext = ActionContext<ViewState, DistilState>;

export const actions = {
  async fetchHomeData(context: ViewContext) {
    // clear any previous state
    requestMutations.clearSolutionRequests(store);
    requestMutations.clearSolutions(store);
    modelMutations.setModels(store, []);
    modelMutations.setFilteredModels(store, []);

    // fetch new state
    await modelActions.fetchModels(store);
    await requestActions.fetchSolutions(store, {});
    requestActions.fetchSolutionRequests(store, {});
  },

  async fetchSearchData(context: ViewContext) {
    const terms = context.getters.getRouteTerms;
    const datasetIDs = context.getters.getRouteJoinDatasets;

    // fetch saved models - subsequent calls to
    await modelActions.fetchModels(store);

    const promises = datasetIDs.map((id: string) => {
      return datasetActions.fetchDataset(store, {
        dataset: id
      });
    });

    promises.push(datasetActions.searchDatasets(store, terms));
    promises.push(modelActions.searchModels(store, terms));

    return Promise.all(promises);
  },

  fetchJoinDatasetsData(context: ViewContext) {
    // clear previous state

    const datasetIDs = context.getters.getRouteJoinDatasets;
    const datasetIDA = datasetIDs[0];
    const datasetIDB = datasetIDs[1];
    Promise.all([
      datasetActions.fetchDataset(store, {
        dataset: datasetIDA
      }),
      datasetActions.fetchDataset(store, {
        dataset: datasetIDB
      }),
      datasetActions.fetchJoinDatasetsVariables(store, {
        datasets: datasetIDs
      })
    ]).then(() => {
      return actions.updateJoinDatasetsData(context);
    });
  },

  updateJoinDatasetsData(context: ViewContext) {
    // clear any previous state
    datasetMutations.clearJoinDatasetsTableData(store);

    const datasetIDs = context.getters.getRouteJoinDatasets;
    const highlight = context.getters.getDecodedHighlight;
    const filterParams = context.getters.getDecodedJoinDatasetsFilterParams;
    const datasets = context.getters.getDatasets;
    const varModes = context.getters.getDecodedVarModes;
    const datasetIDA = datasetIDs[0];
    const datasetIDB = datasetIDs[1];

    // fetch new state
    const datasetA = _.find(datasets, d => {
      return d.id === datasetIDA;
    });
    const datasetB = _.find(datasets, d => {
      return d.id === datasetIDB;
    });

    return Promise.all([
      datasetActions.fetchIncludedVariableSummaries(store, {
        dataset: datasetA.id,
        variables: datasetA.variables,
        filterParams: filterParams,
        highlight: highlight,
        varModes: varModes
      }),
      datasetActions.fetchIncludedVariableSummaries(store, {
        dataset: datasetB.id,
        variables: datasetB.variables,
        filterParams: filterParams,
        highlight: highlight,
        varModes: varModes
      }),
      datasetActions.fetchJoinDatasetsTableData(store, {
        datasets: datasetIDs,
        filterParams: filterParams,
        highlight: highlight
      })
    ]);
  },

  async fetchSelectTargetData(context: ViewContext, clearSummaries: boolean) {
    // clear previous state
    if (clearSummaries) {
      clearVariableSummaries(context);
    }

    // fetch new state
    const dataset = context.getters.getRouteDataset;
    const args = {
      dataset: dataset
    };
    await fetchVariables(context, args);
    return fetchVariableSummaries(context, args);
  },

  clearJoinDatasetsData(context) {
    clearVariablesParamCache(context);
    clearVariableSummaries(context);
  },

  async fetchSelectTrainingData(context: ViewContext, clearSummaries: boolean) {
    // clear any previous state
    datasetMutations.setIncludedTableData(store, null);
    datasetMutations.setExcludedTableData(store, null);

    if (clearSummaries) {
      clearVariableSummaries(context);
    }

    const dataset = context.getters.getRouteDataset;
    const target = context.getters.getRouteTargetVariable;

    fetchJoinSuggestions(context, {
      dataset: dataset
    });

    await Promise.all([
      fetchVariables(context, {
        dataset: dataset
      }),
      datasetActions.fetchDataset(store, {
        dataset: dataset
      })
    ]);
    fetchVariableRankings(context, { dataset, target });
    fetchClusters(context, { dataset });
    return actions.updateSelectTrainingData(context);
  },

  updateSelectTrainingData(context: ViewContext) {
    // clear any previous state

    const dataset = context.getters.getRouteDataset;
    const highlight = context.getters.getDecodedHighlight;
    const filterParams = context.getters.getDecodedSolutionRequestFilterParams;
    const varModes = context.getters.getDecodedVarModes;

    return Promise.all([
      fetchVariableSummaries(context, {
        dataset: dataset,
        filterParams: filterParams,
        highlight: highlight,
        varModes: varModes
      }),
      datasetActions.fetchIncludedTableData(store, {
        dataset: dataset,
        filterParams: filterParams,
        highlight: highlight
      }),
      datasetActions.fetchExcludedTableData(store, {
        dataset: dataset,
        filterParams: filterParams,
        highlight: highlight
      })
    ]);
  },

  async fetchResultsData(context: ViewContext) {
    // clear previous state
    resultMutations.clearTargetSummary(store);
    resultMutations.clearTrainingSummaries(store);
    resultMutations.clearResidualsExtrema(store);
    resultMutations.setIncludedResultTableData(store, null);
    resultMutations.setExcludedResultTableData(store, null);
    modelMutations.setModels(store, []);

    const dataset = context.getters.getRouteDataset;
    const target = context.getters.getRouteTargetVariable;

    // fetch new state
    await fetchVariables(context, {
      dataset: dataset
    });
    await modelActions.fetchModels(store); // Fetch saved models.

    // These are long running processes we won't wait on
    fetchVariableRankings(context, {
      dataset: dataset,
      target: target
    });

    fetchClusters(context, { dataset: dataset });

    await Promise.all([
      fetchSolutionRequests(context, {
        dataset: dataset,
        target: target
      }),

      fetchSolutions(context, {
        dataset: dataset,
        target: target
      }),

      datasetActions.searchDatasets(store, "")
    ]);

    return actions.updateResultsSolution(context);
  },

  updateResultsSolution(context: ViewContext) {
    // clear previous state
    resultMutations.clearResidualsExtrema(store);
    resultMutations.setIncludedResultTableData(store, null);
    resultMutations.setExcludedResultTableData(store, null);

    // fetch new state
    const dataset = context.getters.getRouteDataset;
    const target = context.getters.getRouteTargetVariable;
    const requestIds = context.getters.getRelevantSolutionRequestIds;
    const solutionId = context.getters.getRouteSolutionId;
    const trainingVariables =
      context.getters.getActiveSolutionTrainingVariables;
    const highlight = context.getters.getDecodedHighlight;
    const varModes = context.getters.getDecodedVarModes;

    resultActions.fetchResultTableData(store, {
      dataset: dataset,
      solutionId: solutionId,
      highlight: highlight
    });
    resultActions.fetchTargetSummary(store, {
      dataset: dataset,
      target: target,
      solutionId: solutionId,
      highlight: highlight,
      varMode: varModes.has(target) ? varModes.get(target) : SummaryMode.Default
    });
    resultActions.fetchTrainingSummaries(store, {
      dataset: dataset,
      training: trainingVariables,
      solutionId: solutionId,
      highlight: highlight,
      varModes: varModes
    });
    resultActions.fetchPredictedSummaries(store, {
      dataset: dataset,
      target: target,
      requestIds: requestIds,
      highlight: highlight,
      varModes: varModes
    });

    const task = routeGetters.getRouteTask(store);

    if (!task) {
      console.error(`task is ${task}`);
    } else if (
      task.includes(TaskTypes.REGRESSION) ||
      task.includes(TaskTypes.FORECASTING)
    ) {
      resultActions.fetchResidualsExtrema(store, {
        dataset: dataset,
        target: target,
        solutionId: solutionId
      });
      resultActions.fetchResidualsSummaries(store, {
        dataset: dataset,
        target: target,
        requestIds: requestIds,
        highlight: highlight,
        varModes: varModes
      });
    } else if (task.includes(TaskTypes.CLASSIFICATION)) {
      resultActions.fetchCorrectnessSummaries(store, {
        dataset: dataset,
        target: target,
        requestIds: requestIds,
        highlight: highlight
      });
    } else {
      console.error(`unhandled task type ${task}`);
    }
  },

  async fetchPredictionsData(context: ViewContext) {
    // clear previous state
    predictionMutations.clearTrainingSummaries(store);
    predictionMutations.setIncludedPredictionTableData(store, null);

    const produceRequestId = <string>context.getters.getRouteProduceRequestId;
    const fittedSolutionId = context.getters.getRouteFittedSolutionId;

    // fetch the predictions
    await fetchPredictions(context, {
      fittedSolutionId: fittedSolutionId
    });

    // recover the dataset associated with the currently selected predictions set
    const inferenceDataset = getPredictionsById(
      context.getters.getPredictions,
      produceRequestId
    ).dataset;

    // fetch variales for that dataset
    await fetchVariables(context, {
      dataset: inferenceDataset
    });
    return actions.updatePredictions(context);
  },

  updatePredictions(context: ViewContext) {
    // clear previous state
    predictionMutations.setIncludedPredictionTableData(store, null);

    // fetch new state
    const produceRequestId = <string>context.getters.getRouteProduceRequestId;
    const fittedSolutionId = <string>context.getters.getRouteFittedSolutionId;
    const inferenceDataset = getPredictionsById(
      context.getters.getPredictions,
      produceRequestId
    ).dataset;
    const trainingVariables = <Variable[]>(
      context.getters.getActivePredictionTrainingVariables
    );
    const highlight = <Highlight>context.getters.getDecodedHighlight;
    const varModes = <Map<string, SummaryMode>>(
      context.getters.getDecodedVarModes
    );
    predictionActions.fetchPredictionTableData(store, {
      dataset: inferenceDataset,
      highlight: highlight,
      produceRequestId: produceRequestId
    });
    predictionActions.fetchTrainingSummaries(store, {
      dataset: inferenceDataset,
      training: trainingVariables,
      highlight: highlight,
      varModes: varModes,
      produceRequestId: produceRequestId
    });
    predictionActions.fetchPredictedSummaries(store, {
      highlight: highlight,
      fittedSolutionId: fittedSolutionId
    });
  }
};
