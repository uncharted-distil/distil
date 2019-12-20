import _ from "lodash";
import { ViewState } from "./index";
import { ActionContext } from "vuex";
import store, { DistilState } from "../store";
import { mutations as viewMutations } from "./module";
import { Dictionary } from "../../util/dict";
import {
  actions as datasetActions,
  mutations as datasetMutations
} from "../dataset/module";
import {
  actions as solutionActions,
  mutations as solutionMutations
} from "../solutions/module";
import {
  actions as resultActions,
  mutations as resultMutations
} from "../results/module";
import { getters as routeGetters } from "../route/module";
import { TaskTypes } from "../dataset";

enum ParamCacheKey {
  VARIABLES = "VARIABLES",
  VARIABLE_SUMMARIES = "VARIABLE_SUMMARIES",
  VARIABLE_RANKINGS = "VARIABLE_RANKINGS",
  SOLUTION_REQUESTS = "SOLUTION_REQUESTS",
  JOIN_SUGGESTIONS = "JOIN_SUGGESTIONS"
}

function createCacheable(
  key: ParamCacheKey,
  func: (context: ViewContext, args: Dictionary<string>) => any
) {
  return (context: ViewContext, args: Dictionary<string>) => {
    // execute provided function if params are not cached already or changed
    const params = JSON.stringify(args);
    const cachedParams = context.getters.getFetchParamsCache[key];
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
  (context, args) => {
    return fetchVariables(context, args).then(() => {
      const dataset = args.dataset;
      const variables = context.getters.getVariables;
      const filterParams =
        context.getters.getDecodedSolutionRequestFilterParams;
      const highlight = context.getters.getDecodedHighlight;

      return Promise.all([
        datasetActions.fetchIncludedVariableSummaries(store, {
          dataset: dataset,
          variables: variables,
          filterParams: filterParams,
          highlight: highlight
        }),
        datasetActions.fetchExcludedVariableSummaries(store, {
          dataset: dataset,
          variables: variables,
          filterParams: filterParams,
          highlight: highlight
        })
      ]);
    });
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

const fetchSolutionRequests = createCacheable(
  ParamCacheKey.SOLUTION_REQUESTS,
  (context, args) => {
    return solutionActions.fetchSolutionRequests(store, {
      dataset: args.dataset,
      target: args.target
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
  fetchHomeData(context: ViewContext) {
    // clear any previous state
    solutionMutations.clearSolutionRequests(store);

    // fetch new state
    return solutionActions.fetchSolutionRequests(store, {});
  },

  fetchSearchData(context: ViewContext) {
    const terms = context.getters.getRouteTerms;
    const datasetIDs = context.getters.getRouteJoinDatasets;

    const promises = datasetIDs.map((id: string) => {
      return datasetActions.fetchDataset(store, {
        dataset: id
      });
    });
    promises.push(datasetActions.searchDatasets(store, terms));

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
        highlight: highlight
      }),
      datasetActions.fetchIncludedVariableSummaries(store, {
        dataset: datasetB.id,
        variables: datasetB.variables,
        filterParams: filterParams,
        highlight: highlight
      }),
      datasetActions.fetchJoinDatasetsTableData(store, {
        datasets: datasetIDs,
        filterParams: filterParams,
        highlight: highlight
      })
    ]);
  },

  fetchSelectTargetData(context: ViewContext, clearSummaries: boolean) {
    // clear previous state
    if (clearSummaries) {
      clearVariableSummaries(context);
    }

    // fetch new state
    const dataset = context.getters.getRouteDataset;
    const args = {
      dataset: dataset
    };
    return fetchVariables(context, args).then(() => {
      return fetchVariableSummaries(context, args);
    });
  },

  clearJoinDatasetsData(context) {
    clearVariablesParamCache(context);
    clearVariableSummaries(context);
  },

  fetchSelectTrainingData(context: ViewContext, clearSummaries: boolean) {
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

    return Promise.all([
      fetchVariables(context, {
        dataset: dataset
      }),
      datasetActions.fetchDataset(store, {
        dataset: dataset
      })
    ]).then(() => {
      fetchVariableRankings(context, { dataset, target });

      return actions.updateSelectTrainingData(context);
    });
  },

  updateSelectTrainingData(context: ViewContext) {
    // clear any previous state

    const dataset = context.getters.getRouteDataset;
    const highlight = context.getters.getDecodedHighlight;
    const filterParams = context.getters.getDecodedSolutionRequestFilterParams;

    return Promise.all([
      fetchVariableSummaries(context, {
        dataset: dataset,
        filterParams: filterParams,
        highlight: highlight
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

  fetchResultsData(context: ViewContext) {
    // clear previous state
    resultMutations.clearTargetSummary(store);
    resultMutations.clearTrainingSummaries(store);
    resultMutations.clearResidualsExtrema(store);
    resultMutations.setIncludedResultTableData(store, null);
    resultMutations.setExcludedResultTableData(store, null);

    const dataset = context.getters.getRouteDataset;
    const target = context.getters.getRouteTargetVariable;
    // fetch new state
    return fetchVariables(context, {
      dataset: dataset
    })
      .then(() => {
        fetchVariableRankings(context, {
          dataset: dataset,
          target: target
        });
        return fetchSolutionRequests(context, {
          dataset: dataset,
          target: target
        });
      })
      .then(() => {
        return actions.updateResultsSolution(context);
      });
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

    resultActions.fetchResultTableData(store, {
      dataset: dataset,
      solutionId: solutionId,
      highlight: highlight
    });
    resultActions.fetchTargetSummary(store, {
      dataset: dataset,
      target: target,
      solutionId: solutionId,
      highlight: highlight
    });
    resultActions.fetchTrainingSummaries(store, {
      dataset: dataset,
      training: trainingVariables,
      solutionId: solutionId,
      highlight: highlight
    });
    resultActions.fetchPredictedSummaries(store, {
      dataset: dataset,
      target: target,
      requestIds: requestIds,
      highlight: highlight
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
        highlight: highlight
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

  fetchPredictionsData(context: ViewContext) {
    // clear previous state
    resultMutations.clearTargetSummary(store);
    resultMutations.clearTrainingSummaries(store);
    resultMutations.clearResidualsExtrema(store);
    resultMutations.setIncludedResultTableData(store, null);
    resultMutations.setExcludedResultTableData(store, null);

    const dataset = context.getters.getRouteDataset;
    const target = context.getters.getRouteTargetVariable;
    // fetch new state
    return fetchVariables(context, {
      dataset: dataset
    })
      .then(() => {
        fetchVariableRankings(context, {
          dataset: dataset,
          target: target
        });
        return fetchSolutionRequests(context, {
          dataset: dataset,
          target: target
        });
      })
      .then(() => {
        return actions.updatePredictionsSolution(context);
      });
  },

  updatePredictionsSolution(context: ViewContext) {
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

    resultActions.fetchResultTableData(store, {
      dataset: dataset,
      solutionId: solutionId,
      highlight: highlight
    });
    resultActions.fetchTargetSummary(store, {
      dataset: dataset,
      target: target,
      solutionId: solutionId,
      highlight: highlight
    });
    resultActions.fetchTrainingSummaries(store, {
      dataset: dataset,
      training: trainingVariables,
      solutionId: solutionId,
      highlight: highlight
    });
    resultActions.fetchPredictedSummaries(store, {
      dataset: dataset,
      target: target,
      requestIds: requestIds,
      highlight: highlight
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
        highlight: highlight
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
  }
};
