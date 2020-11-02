import _ from "lodash";
import { ActionContext } from "vuex";
import {
  filterArrayByPage,
  NUM_PER_PAGE,
  NUM_PER_TARGET_PAGE,
  searchVariables,
  sortVariablesByImportance,
} from "../../util/data";
import { Dictionary } from "../../util/dict";
import { EXCLUDE_FILTER, Filter, invertFilter } from "../../util/filters";
import { getPredictionsById } from "../../util/predictions";
import {
  DataMode,
  Highlight,
  SummaryMode,
  TaskTypes,
  Variable,
} from "../dataset";
import {
  actions as datasetActions,
  mutations as datasetMutations,
} from "../dataset/module";
import {
  actions as modelActions,
  mutations as modelMutations,
} from "../model/module";
import {
  actions as predictionActions,
  mutations as predictionMutations,
} from "../predictions/module";
import {
  actions as requestActions,
  getters as requestGetters,
  mutations as requestMutations,
} from "../requests/module";
import {
  actions as resultActions,
  mutations as resultMutations,
} from "../results/module";
import { SELECT_TARGET_ROUTE } from "../route";
import { getters as routeGetters } from "../route/module";
import store, { DistilState } from "../store";
import { ViewState } from "./index";
import { getters as viewGetters, mutations as viewMutations } from "./module";

enum ParamCacheKey {
  VARIABLES = "VARIABLES",
  VARIABLE_SUMMARIES = "VARIABLE_SUMMARIES",
  VARIABLE_RANKINGS = "VARIABLE_RANKINGS",
  SOLUTION_VARIABLE_RANKINGS = "SOLUTION_VARIABLE_RANKINGS",
  SEARCH_REQUESTS = "SEARCH_REQUESTS",
  SOLUTIONS = "SOLUTIONS",
  PREDICTIONS_REQUESTS = "PREDICTIONS_REQUESTS",
  PREDICTIONS = "PREDICTIONS",
  JOIN_SUGGESTIONS = "JOIN_SUGGESTIONS",
  CLUSTERS = "CLUSTERS",
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
        value: params,
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
      searchQuery: args.searchQuery,
    });
  }
);

const fetchVariables = createCacheable(
  ParamCacheKey.VARIABLES,
  (context, args) => {
    return datasetActions.fetchVariables(store, {
      dataset: args.dataset,
    });
  }
);

const fetchVariableSummaries = async (context, args) => {
  await fetchVariables(context, args);
  const dataset = args.dataset as string;
  const variables = context.getters.getVariables as Variable[];
  const filterParams = context.getters.getDecodedSolutionRequestFilterParams;
  const highlight = context.getters.getDecodedHighlight;
  const varModes = context.getters.getDecodedVarModes;
  const dataMode = context.getters.getDataMode;

  const currentRoute = routeGetters.getRoutePath(store);
  const ranked = routeGetters.getRouteIsTrainingVariablesRanked(store);
  const pages = routeGetters.getAllRoutePages(store);
  const targetVariable = routeGetters.getTargetVariable(store);

  const currentPageIndexes = pages[currentRoute];
  const mainPageIndex = currentPageIndexes[0];
  const trainingIndex = currentPageIndexes?.[1];

  const pageLength =
    currentRoute === SELECT_TARGET_ROUTE ? NUM_PER_TARGET_PAGE : NUM_PER_PAGE;

  const searches = routeGetters.getAllSearchesByRoute(store);
  const currentPageSearches = searches[currentRoute];
  const currentSearch = currentPageSearches[0];
  const trainingSearch = currentPageSearches[1];

  const allTrainingVariables = routeGetters.getTrainingVariables(store);

  const searchedTrainingVariables = searchVariables(
    allTrainingVariables,
    trainingSearch
  );

  const activeTrainingVariables = trainingIndex
    ? filterArrayByPage(trainingIndex, pageLength, searchedTrainingVariables)
    : [];

  const allTargetTrainingVariables = targetVariable
    ? [targetVariable, ...searchedTrainingVariables]
    : [];
  const activeTargetTrainingVariables = targetVariable
    ? [targetVariable, ...activeTrainingVariables]
    : [];

  const allTargetTrainingVariableNames = allTargetTrainingVariables.map((sv) =>
    sv.colDisplayName.toLowerCase()
  );

  const presortedVariables = ranked
    ? sortVariablesByImportance(variables.slice())
    : variables;

  const searchedPresortedVariables = searchVariables(
    presortedVariables,
    currentSearch
  );

  const mainPageVariables = searchedPresortedVariables
    .filter(
      (v) =>
        allTargetTrainingVariableNames.indexOf(v.colDisplayName.toLowerCase()) <
        0
    )
    .slice((mainPageIndex - 1) * pageLength, mainPageIndex * pageLength);

  const allActiveVariables = [
    ...activeTargetTrainingVariables,
    ...mainPageVariables,
  ];

  return Promise.all([
    datasetActions.fetchIncludedVariableSummaries(store, {
      dataset: dataset,
      variables: allActiveVariables,
      filterParams: filterParams,
      highlight: highlight,
      dataMode: dataMode,
      varModes: varModes,
    }),
    datasetActions.fetchExcludedVariableSummaries(store, {
      dataset: dataset,
      variables: allActiveVariables,
      filterParams: filterParams,
      highlight: highlight,
      dataMode: dataMode,
      varModes: varModes,
    }),
  ]);
};

const fetchVariableRankings = createCacheable(
  ParamCacheKey.VARIABLE_RANKINGS,
  (context, args) => {
    // if target or dataset has changed, clear previous rankings before re-fetch
    // this is needed because since user decides variable rankings to be updated, re-fetching doesn't always replace the previous data
    datasetActions.updateVariableRankings(store, {
      dataset: args.dataset,
      rankings: {},
    });
    datasetActions.fetchVariableRankings(store, {
      dataset: args.dataset,
      target: args.target,
    });
  }
);

const fetchSolutionVariableRankings = createCacheable(
  ParamCacheKey.SOLUTION_VARIABLE_RANKINGS,
  (context, args) => {
    resultActions.fetchFeatureImportanceRanking(store, {
      solutionID: args.solutionID,
    });
  }
);

const fetchClusters = createCacheable(
  ParamCacheKey.CLUSTERS,
  (context, args) => {
    datasetActions.fetchClusters(store, {
      dataset: args.dataset,
    });
  }
);

const fetchSolutionRequests = createCacheable(
  ParamCacheKey.SEARCH_REQUESTS,
  (context, args) => {
    return requestActions.fetchSolutionRequests(store, {
      dataset: args.dataset,
      target: args.target,
    });
  }
);

const fetchSolutions = createCacheable(
  ParamCacheKey.SOLUTIONS,
  (context, args) => {
    return requestActions.fetchSolutions(store, {
      dataset: args.dataset,
      target: args.target,
    });
  }
);

const fetchPredictions = createCacheable(
  ParamCacheKey.PREDICTIONS,
  (context, args) => {
    return requestActions.fetchPredictions(store, {
      fittedSolutionId: args.fittedSolutionId,
    });
  }
);

function clearVariablesParamCache(context: ViewContext) {
  // clear variable param cache to allow re-fetching variables
  viewMutations.setFetchParamsCache(context, {
    key: ParamCacheKey.VARIABLES,
    value: undefined,
  });
}

function clearVariableSummaries(context: ViewContext) {
  datasetMutations.clearVariableSummaries(store);

  viewMutations.setFetchParamsCache(context, {
    key: ParamCacheKey.VARIABLE_SUMMARIES,
    value: undefined,
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
        dataset: id,
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
        dataset: datasetIDA,
      }),
      datasetActions.fetchDataset(store, {
        dataset: datasetIDB,
      }),
      datasetActions.fetchJoinDatasetsVariables(store, {
        datasets: datasetIDs,
      }),
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
    const dataMode = context.getters.getDataMode as DataMode;
    const varModes = context.getters.getDecodedVarModes;
    const datasetIDA = datasetIDs[0];
    const datasetIDB = datasetIDs[1];

    // fetch new state
    const datasetA = _.find(datasets, (d) => {
      return d.id === datasetIDA;
    });
    const datasetB = _.find(datasets, (d) => {
      return d.id === datasetIDB;
    });

    return Promise.all([
      datasetActions.fetchIncludedVariableSummaries(store, {
        dataset: datasetA.id,
        variables: datasetA.variables,
        filterParams: filterParams,
        highlight: highlight,
        dataMode: dataMode,
        varModes: varModes,
      }),
      datasetActions.fetchIncludedVariableSummaries(store, {
        dataset: datasetB.id,
        variables: datasetB.variables,
        filterParams: filterParams,
        highlight: highlight,
        dataMode: dataMode,
        varModes: varModes,
      }),
      datasetActions.fetchJoinDatasetsTableData(store, {
        datasets: datasetIDs,
        filterParams: filterParams,
        highlight: highlight,
      }),
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
      dataset: dataset,
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
      dataset: dataset,
    });

    await Promise.all([
      fetchVariables(context, {
        dataset: dataset,
      }),
      datasetActions.fetchDataset(store, {
        dataset: dataset,
      }),
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
    const dataMode = context.getters.getDataMode;
    const varModes = context.getters.getDecodedVarModes;

    return Promise.all([
      fetchVariableSummaries(context, {
        dataset: dataset,
        filterParams: filterParams,
        highlight: highlight,
        varModes: varModes,
      }),
      datasetActions.fetchIncludedTableData(store, {
        dataset: dataset,
        filterParams: filterParams,
        highlight: highlight,
        dataMode: dataMode,
      }),
      datasetActions.fetchExcludedTableData(store, {
        dataset: dataset,
        filterParams: filterParams,
        highlight: highlight,
        dataMode: dataMode,
      }),
    ]);
  },
  updateHighlight(context: ViewContext) {
    const dataset = context.getters.getRouteDataset;
    const highlight = context.getters.getDecodedHighlight;
    const filterParams = context.getters.getDecodedSolutionRequestFilterParams;
    const dataMode = context.getters.getDataMode;
    return Promise.all([
      datasetActions.fetchHighlightedTableData(store, {
        dataset: dataset,
        filterParams: filterParams,
        highlight: highlight,
        dataMode: dataMode,
        include: true,
      }), // include
      datasetActions.fetchHighlightedTableData(store, {
        dataset: dataset,
        filterParams: filterParams,
        highlight: highlight,
        dataMode: dataMode,
        include: false,
      }), // exclude
    ]);
  },
  async updateAreaOfInterest(context: ViewContext, filter: Filter) {
    const dataset = context.getters.getRouteDataset;
    const highlight = context.getters.getDecodedHighlight;
    const filterParams = context.getters.getDecodedSolutionRequestFilterParams;
    const dataMode = context.getters.getDataMode;
    // artificially add filter but dont add it to the url
    // this is a hack to avoid adding an extra field just for the area of interest
    const clonedFilterParams = _.cloneDeep(filterParams);
    clonedFilterParams.filters.push(filter);
    const clonedExcludeFilter = _.cloneDeep(filter);
    // the exclude has to invert all the filters -- the route does a collective NOT() and
    // for areaOfInterest we need compounded ands so therefore we invert client side pass in
    // as an include and that removes the collective NOT
    const clonedFilterParamsExclude = _.cloneDeep(filterParams);
    clonedFilterParamsExclude.filters.forEach((f) => {
      f.mode = invertFilter(f.mode);
    });
    clonedFilterParamsExclude.filters.push(clonedExcludeFilter);
    const invertedHighlight =
      highlight === null ? null : { ...highlight, include: EXCLUDE_FILTER };
    return Promise.all([
      datasetActions.fetchAreaOfInterestData(store, {
        dataset: dataset,
        filterParams: clonedFilterParams,
        highlight: highlight,
        dataMode: dataMode,
        include: true,
        mutatorIsInclude: true,
        isExclude: false,
      }), // include inner tiles
      datasetActions.fetchAreaOfInterestData(store, {
        dataset: dataset,
        filterParams: clonedFilterParams,
        highlight: invertedHighlight,
        dataMode: dataMode,
        include: true,
        mutatorIsInclude: false,
        isExclude: false,
      }), // include outer tiles
      datasetActions.fetchAreaOfInterestData(store, {
        dataset: dataset,
        filterParams: clonedFilterParamsExclude,
        highlight: highlight,
        dataMode: dataMode,
        include: true,
        mutatorIsInclude: true,
        isExclude: true,
      }), // exclude inner tiles
      datasetActions.fetchAreaOfInterestData(store, {
        dataset: dataset,
        filterParams: clonedFilterParamsExclude,
        highlight: invertedHighlight,
        dataMode: dataMode,
        include: true,
        mutatorIsInclude: false,
        isExclude: true,
      }), // include outer tiles
    ]);
  },
  clearHighlight(context: ViewContext) {
    datasetMutations.setHighlightedIncludeTableData(store, null);
    datasetMutations.setHighlightedExcludeTableData(store, null);
  },
  async fetchResultsData(context: ViewContext) {
    // clear previous state
    resultMutations.clearTargetSummary(store);
    resultMutations.clearTrainingSummaries(store);
    resultMutations.clearResidualsExtrema(store);
    resultMutations.setIncludedResultTableData(store, null);
    resultMutations.setExcludedResultTableData(store, null);
    modelMutations.setModels(store, []);

    const dataset = routeGetters.getRouteDataset(store);
    const target = routeGetters.getRouteTargetVariable(store);
    const solutionID = routeGetters.getRouteSolutionId(store);

    // fetch new state
    await fetchVariables(context, {
      dataset: dataset,
    });
    await modelActions.fetchModels(store); // Fetch saved models.

    // These are long running processes we won't wait on
    fetchClusters(context, { dataset: dataset });

    await Promise.all([
      fetchSolutionVariableRankings(context, { solutionID: solutionID }),

      fetchSolutionRequests(context, {
        dataset: dataset,
        target: target,
      }),

      fetchSolutions(context, {
        dataset: dataset,
        target: target,
      }),

      datasetActions.searchDatasets(store, ""),
    ]);

    return actions.updateResultsSolution(context);
  },

  updateResultsSummaries(context: ViewContext) {
    const dataset = routeGetters.getRouteDataset(store);
    const trainingVariables = requestGetters.getActiveSolutionTrainingVariables(
      store
    );
    const highlight = routeGetters.getDecodedHighlight(store);
    const dataMode = context.getters.getDataMode;
    const varModes: Map<string, SummaryMode> = routeGetters.getDecodedVarModes(
      store
    );
    const solutionId = routeGetters.getRouteSolutionId(store);
    const page = routeGetters.getRouteResultTrainingVarsPage(store);
    const pageSize = NUM_PER_PAGE;
    const activeTrainingVariables = filterArrayByPage(
      page,
      pageSize,
      trainingVariables
    );

    resultActions.fetchTrainingSummaries(store, {
      dataset: dataset,
      training: activeTrainingVariables,
      solutionId: solutionId,
      highlight: highlight,
      dataMode: dataMode,
      varModes: varModes,
    });
  },

  updateResultsSolution(context: ViewContext) {
    // clear previous state
    resultMutations.clearResidualsExtrema(store);
    resultMutations.setIncludedResultTableData(store, null);
    resultMutations.setExcludedResultTableData(store, null);

    // fetch new state
    const dataset = routeGetters.getRouteDataset(store);
    const target = routeGetters.getRouteTargetVariable(store);
    const requestIds = requestGetters.getRelevantSolutionRequestIds(store);
    const solutionId = routeGetters.getRouteSolutionId(store);
    const highlight = routeGetters.getDecodedHighlight(store);
    const dataMode = context.getters.getDataMode;
    const varModes: Map<string, SummaryMode> = routeGetters.getDecodedVarModes(
      store
    );
    const size = routeGetters.getRouteDataSize(store);

    // before fetching narrow

    resultActions.fetchResultTableData(store, {
      dataset: dataset,
      solutionId: solutionId,
      highlight: highlight,
      dataMode: dataMode,
      size,
    });
    resultActions.fetchTargetSummary(store, {
      dataset: dataset,
      target: target,
      solutionId: solutionId,
      highlight: highlight,
      dataMode: dataMode,
      varMode: varModes.has(target)
        ? varModes.get(target)
        : SummaryMode.Default,
    });

    actions.updateResultsSummaries(context);

    resultActions.fetchPredictedSummaries(store, {
      dataset: dataset,
      target: target,
      requestIds: requestIds,
      highlight: highlight,
      dataMode: dataMode,
      varModes: varModes,
    });
    resultActions.fetchFeatureImportanceRanking(store, {
      solutionID: solutionId,
    });

    resultActions.fetchConfidenceSummaries(store, {
      dataset: dataset,
      target: target,
      requestIds: requestIds,
      highlight: highlight,
      dataMode: dataMode,
      varModes: varModes,
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
        solutionId: solutionId,
      });
      resultActions.fetchResidualsSummaries(store, {
        dataset: dataset,
        target: target,
        requestIds: requestIds,
        highlight: highlight,
        dataMode: dataMode,
        varModes: varModes,
      });
    } else if (task.includes(TaskTypes.CLASSIFICATION)) {
      resultActions.fetchCorrectnessSummaries(store, {
        dataset: dataset,
        target: target,
        requestIds: requestIds,
        highlight: highlight,
        dataMode: dataMode,
        varModes: varModes,
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
      fittedSolutionId: fittedSolutionId,
    });

    // recover the dataset associated with the currently selected predictions set
    const inferenceDataset = getPredictionsById(
      context.getters.getPredictions,
      produceRequestId
    ).dataset;

    // fetch variales for that dataset
    await fetchVariables(context, {
      dataset: inferenceDataset,
    });
    return actions.updatePredictions(context);
  },

  updatePredictionTrainingSummaries(context: ViewContext) {
    // fetch new state
    const produceRequestId = <string>context.getters.getRouteProduceRequestId;
    const inferenceDataset = getPredictionsById(
      context.getters.getPredictions,
      produceRequestId
    ).dataset;
    const highlight = <Highlight>context.getters.getDecodedHighlight;
    const varModes = <Map<string, SummaryMode>>(
      context.getters.getDecodedVarModes
    );
    const currentSearch = <string>(
      context.getters.getRouteResultTrainingVarsSearch
    );
    const trainingVariables = <Variable[]>(
      searchVariables(
        context.getters.getActivePredictionTrainingVariables,
        currentSearch
      )
    );
    const page = routeGetters.getRouteResultTrainingVarsPage(store);
    const pageSize = NUM_PER_PAGE;
    const activeTrainingVariables = filterArrayByPage(
      page,
      pageSize,
      trainingVariables
    );

    predictionActions.fetchTrainingSummaries(store, {
      dataset: inferenceDataset,
      training: activeTrainingVariables,
      highlight: highlight,
      varModes: varModes,
      produceRequestId: produceRequestId,
    });
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
    const highlight = <Highlight>context.getters.getDecodedHighlight;
    const size = routeGetters.getRouteDataSize(store);

    predictionActions.fetchPredictionTableData(store, {
      dataset: inferenceDataset,
      highlight: highlight,
      produceRequestId: produceRequestId,
      size,
    });

    actions.updatePredictionTrainingSummaries(context);

    predictionActions.fetchPredictedSummaries(store, {
      highlight: highlight,
      fittedSolutionId: fittedSolutionId,
    });
  },
};
