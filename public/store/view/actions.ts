/**
 *
 *    Copyright © 2021 Uncharted Software Inc.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

import _ from "lodash";
import { ActionContext } from "vuex";
import {
  createEmptyTableData,
  filterArrayByPage,
  NUM_PER_PAGE,
  NUM_PER_TARGET_PAGE,
  searchVariables,
  sortVariablesByImportance,
} from "../../util/data";
import { Dictionary } from "../../util/dict";
import {
  EXCLUDE_FILTER,
  Filter,
  FilterParams,
  invertFilter,
} from "../../util/filters";
import { getPredictionsById } from "../../util/predictions";
import { filterBadRequests } from "../../util/solutions";
import {
  DataMode,
  Highlight,
  SummaryMode,
  TaskTypes,
  Variable,
  VariableSummaryKey,
} from "../dataset";
import {
  actions as datasetActions,
  getters as datasetGetters,
  mutations as datasetMutations,
} from "../dataset/module";
import {
  actions as modelActions,
  mutations as modelMutations,
} from "../model/module";
import { actions as predictionActions } from "../predictions/module";
import { Predictions } from "../requests";
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
  OUTLIERS = "OUTLIERS",
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
  const variables =
    args.variables ?? (context.getters.getVariables as Variable[]);
  const filterParams = context.getters
    .getDecodedSolutionRequestFilterParams as FilterParams;
  const highlights = context.getters.getDecodedHighlights as Highlight[];
  const varModes = context.getters.getDecodedVarModes;
  const dataMode = context.getters.getDataMode;

  const currentRoute = routeGetters.getRoutePath(store);
  const ranked = routeGetters.getRouteIsTrainingVariablesRanked(store);
  const targetVariable = routeGetters.getTargetVariable(store);

  const pages = routeGetters.getAllRoutePages(store);
  let currentPageIndexes = [];
  if (pages[currentRoute]) {
    currentPageIndexes = pages[currentRoute];
  } else {
    const errorMessage = `
      The store/route/getters getAllRoutePages() method does not have
      a definition for the ${currentRoute} route.`;
    console.error(errorMessage);
  }

  const mainPageIndex = currentPageIndexes?.[0];
  const trainingIndex = currentPageIndexes?.[1];

  let pageLength = NUM_PER_PAGE;
  if (currentRoute === SELECT_TARGET_ROUTE) {
    pageLength = NUM_PER_TARGET_PAGE;
  }

  const searches = routeGetters.getAllSearchesByRoute(store);
  let currentPageSearches = [];
  if (searches[currentRoute]) {
    currentPageSearches = searches[currentRoute];
  } else {
    const errorMessage = `
      The store/route/getters getAllSearchesByRoute() method does not have
      a definition for the ${currentRoute} route.`;
    console.error(errorMessage);
  }

  const currentSearch = currentPageSearches?.[0];
  const trainingSearch = currentPageSearches?.[1];

  const allTrainingVariables = routeGetters.getTrainingVariables(store);

  const sortedAllTrainingVariables = ranked
    ? sortVariablesByImportance(allTrainingVariables.slice())
    : allTrainingVariables;

  const searchedTrainingVariables = searchVariables(
    sortedAllTrainingVariables,
    trainingSearch
  );

  const activeTrainingVariables = trainingIndex
    ? filterArrayByPage(trainingIndex, pageLength, searchedTrainingVariables)
    : [];
  const activeTargetTrainingVariables = targetVariable
    ? [targetVariable, ...activeTrainingVariables]
    : [];

  const presortedVariables = sortVariablesByImportance(variables.slice());

  const searchedPresortedVariables = searchVariables(
    presortedVariables,
    currentSearch
  );

  const mainPageVariables = searchedPresortedVariables.slice(
    (mainPageIndex - 1) * pageLength,
    mainPageIndex * pageLength
  );

  const allActiveVariables = [
    ...activeTargetTrainingVariables,
    ...mainPageVariables,
  ];

  const fetchArgs = {
    dataset: dataset,
    variables: allActiveVariables,
    filterParams: filterParams,
    highlights: highlights,
    dataMode: dataMode,
    varModes: varModes,
  };

  return Promise.all([
    datasetActions.fetchIncludedVariableSummaries(store, fetchArgs),
    datasetActions.fetchExcludedVariableSummaries(store, fetchArgs),
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

const fetchOutliers = createCacheable(
  ParamCacheKey.OUTLIERS,
  (context, args) => {
    datasetActions.fetchOutliers(store, args.dataset);
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
  async fetchHomeData() {
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
      datasetActions.fetchMultiBandCombinations(store, { dataset: datasetIDA }),
      datasetActions.fetchMultiBandCombinations(store, { dataset: datasetIDB }),
    ]).then(() => {
      return actions.updateJoinDatasetsData(context);
    });
  },
  updateJoinDatasetsData(context: ViewContext) {
    const datasetIDs = context.getters.getRouteJoinDatasets;
    const highlights = context.getters.getDecodedJoinDatasetsHighlight;
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
    const groupingA = datasetA.variables.reduce((a, v) => {
      const hiddenVars = v.grouping?.hidden as string[];
      if (hiddenVars) {
        a = a.concat(hiddenVars);
      }
      return a;
    }, []);
    const groupingB = datasetB.variables.reduce((a, v) => {
      const hiddenVars = v.grouping?.hidden as string[];
      if (hiddenVars) {
        a = a.concat(hiddenVars);
      }
      return a;
    }, []);

    return Promise.all([
      datasetActions.fetchIncludedVariableSummaries(store, {
        dataset: datasetA.id,
        variables: datasetA.variables.filter(
          (v) => groupingA.indexOf(v.key) < 0
        ),
        filterParams: filterParams[datasetA.id],
        highlights: highlights[datasetA.id],
        dataMode: dataMode,
        varModes: varModes,
      }),
      datasetActions.fetchIncludedVariableSummaries(store, {
        dataset: datasetB.id,
        variables: datasetB.variables.filter(
          (v) => groupingB.indexOf(v.key) < 0
        ),
        filterParams: filterParams[datasetB.id],
        highlights: highlights[datasetB.id],
        dataMode: dataMode,
        varModes: varModes,
      }),
      datasetActions.fetchJoinDatasetsTableData(store, {
        datasets: datasetIDs,
        filterParams: filterParams,
        highlights: highlights,
      }),
    ]);
  },

  clearAllData() {
    datasetMutations.clearVariableSummaries(store);
    datasetMutations.setIncludedTableData(store, createEmptyTableData());
    datasetMutations.setExcludedTableData(store, createEmptyTableData());
  },

  clearDatasetTableData() {
    datasetMutations.setIncludedTableData(store, createEmptyTableData());
    datasetMutations.setExcludedTableData(store, createEmptyTableData());
  },

  async fetchSelectTargetData(context: ViewContext, clearSummaries: boolean) {
    const dataset = context.getters.getRouteDataset;
    // clear previous state
    if (clearSummaries) {
      clearVariableSummaries(context);
      datasetMutations.setVariables(store, []);
      await datasetActions.fetchVariables(store, { dataset });
    }
    // fetch new state
    return fetchVariableSummaries(context, { dataset });
  },

  async fetchDataExplorerData(context: ViewContext, variables: Variable[]) {
    // fetch new state
    const dataset = context.getters.getRouteDataset;
    await fetchVariableSummaries(context, { dataset, variables });
    fetchClusters(context, { dataset });
    fetchOutliers(context, { dataset });
    fetchJoinSuggestions(context, {
      dataset: dataset,
    });
  },

  updateDataExplorerData(context: ViewContext) {
    const args = {
      dataset: context.getters.getRouteDataset as string,
      filterParams: context.getters
        .getDecodedSolutionRequestFilterParams as FilterParams,
      highlights: context.getters.getDecodedHighlights as Highlight[],
    };
    const variableArgs = {
      ...args,
      varModes: context.getters.getDecodedVarModes,
    };
    const tableDataArgs = {
      ...args,
      dataMode: context.getters.getDataMode,
    };

    return Promise.all([
      fetchVariableSummaries(context, variableArgs),
      datasetActions.fetchIncludedTableData(store, tableDataArgs),
      datasetActions.fetchExcludedTableData(store, tableDataArgs),
    ]);
  },

  clearJoinDatasetsData(context) {
    clearVariablesParamCache(context);
    clearVariableSummaries(context);
  },

  async fetchSelectTrainingData(context: ViewContext, clearSummaries: boolean) {
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
    if (target) {
      fetchVariableRankings(context, { dataset, target });
    }
    fetchClusters(context, { dataset });
    fetchOutliers(context, { dataset });

    return actions.updateSelectTrainingData(context);
  },
  updateSelectVariables(context: ViewContext) {
    const args = {
      dataset: context.getters.getRouteDataset,
      filterParams: context.getters
        .getDecodedSolutionRequestFilterParams as FilterParams,
      highlights: context.getters.getDecodedHighlights as Highlight[],
    };
    const variableArgs = {
      ...args,
      varModes: context.getters.getDecodedVarModes,
    };
    return fetchVariableSummaries(context, variableArgs);
  },
  updateSelectTrainingData(context: ViewContext) {
    const args = {
      dataset: context.getters.getRouteDataset,
      filterParams: context.getters
        .getDecodedSolutionRequestFilterParams as FilterParams,
      highlights: context.getters.getDecodedHighlights as Highlight[],
    };
    const variableArgs = {
      ...args,
      varModes: context.getters.getDecodedVarModes,
    };
    const tableDataArgs = {
      ...args,
      dataMode: context.getters.getDataMode,
    };

    return Promise.all([
      fetchVariableSummaries(context, variableArgs),
      datasetActions.fetchIncludedTableData(store, tableDataArgs),
      datasetActions.fetchExcludedTableData(store, tableDataArgs),
    ]);
  },

  async updateVariableSummaries(context: ViewContext) {
    const args = {
      dataset: context.getters.getRouteDataset,
      filterParams: context.getters
        .getDecodedSolutionRequestFilterParams as FilterParams,
      highlights: context.getters.getDecodedHighlights as Highlight[],
      varModes: context.getters.getDecodedVarModes,
    };
    await fetchVariableSummaries(context, args);
  },

  updateLabelData(context: ViewContext) {
    const dataset = context.getters.getRouteDataset;
    const highlights = context.getters.getDecodedHighlights as Highlight[];
    const filterParams = context.getters
      .getDecodedSolutionRequestFilterParams as FilterParams;
    const numRows = datasetGetters.getNumberOfRecords(store);
    filterParams.size = numRows;
    const dataMode = context.getters.getDataMode;
    const variables = datasetGetters.getVariables(store);
    const varModes = context.getters.getDecodedVarModes;
    const orderBy = routeGetters.getOrderBy(store);
    filterParams.variables = variables.map((v) => v.key);
    const label = routeGetters.getRouteLabel(store);
    if (
      highlights.some((h) => {
        return h.key === label;
      })
    ) {
      datasetMutations.clearVariableSummaries(store);
    } else {
      datasetMutations.setVariableSummary(store, {
        key: VariableSummaryKey(label, dataset),
        summary: null,
      });
    }

    return Promise.all([
      datasetActions.fetchIncludedVariableSummaries(store, {
        dataset,
        variables,
        filterParams,
        highlights,
        dataMode,
        varModes,
      }),
      datasetActions.fetchIncludedTableData(store, {
        dataset,
        filterParams,
        highlights,
        dataMode,
        orderBy,
      }),
    ]);
  },
  updateHighlight(context: ViewContext) {
    const dataset = context.getters.getRouteDataset;
    const variables = datasetGetters.getVariables(store);
    let variableNames = variables.map((v) => {
      return v.colName;
    });
    // upon refresh variables are not available but they exist in route url so fetch those
    if (!variables.length) {
      variableNames = routeGetters.getDecodedSolutionRequestFilterParams(store)
        .variables;
    }
    const dataMode = context.getters.getDataMode;
    const baseline = {
      highlights: { list: [] },
      filters: {
        list: [],
      },
      variables: variableNames,
      size: Number.MAX_SAFE_INTEGER,
    } as FilterParams;
    return datasetActions.fetchBaselineTableData(store, {
      dataset: dataset,
      filterParams: baseline,
      highlights: [],
      dataMode: dataMode,
    });
  },
  async updateAreaOfInterest(context: ViewContext, filter: Filter) {
    const dataset = context.getters.getRouteDataset;
    const filterParams = context.getters
      .getDecodedSolutionRequestFilterParams as FilterParams;
    const highlights = context.getters.getDecodedHighlights as Highlight[];
    const dataMode = context.getters.getDataMode;
    const variables = datasetGetters.getAllVariables(store);
    // artificially add filter but dont add it to the url
    // this is a hack to avoid adding an extra field just for the area of interest
    const clonedFilterParams = _.cloneDeep(filterParams);
    clonedFilterParams.filters.list.push(filter);
    clonedFilterParams.variables = variables.map((v) => {
      return v.key;
    });

    // the exclude has to invert all the filters -- the route does a collective NOT() and
    // for areaOfInterest we need compounded ands so therefore we invert client side pass in
    // as an include and that removes the collective NOT
    const clonedFilterParamsExclude = _.cloneDeep(filterParams);
    const setMap = new Map(
      clonedFilterParamsExclude.filters.list.map((f) => {
        return [f.set, true];
      })
    );
    clonedFilterParamsExclude.filters.list.forEach((f) => {
      f.mode = invertFilter(f.mode);
    });
    setMap.forEach((v, k) => {
      const tmpFilter = _.cloneDeep(filter);
      tmpFilter.set = k;
      clonedFilterParamsExclude.filters.list.push(tmpFilter);
    });

    const baseline = {
      highlights: { list: [] },
      filters: {
        list: [filter],
      },
      size: Number.MAX_SAFE_INTEGER,
      variables: clonedFilterParams.variables,
    } as FilterParams;
    return Promise.all([
      datasetActions.fetchAreaOfInterestData(store, {
        dataset: dataset,
        filterParams: clonedFilterParams,
        highlights: highlights,
        dataMode: dataMode,
        include: true,
        mutatorIsInclude: true,
        isExclude: false,
      }), // include inner tiles
      datasetActions.fetchAreaOfInterestData(store, {
        dataset: dataset,
        filterParams: baseline,
        highlights: [],
        dataMode: dataMode,
        include: true,
        mutatorIsInclude: false,
        isExclude: false,
      }), // include outer tiles
      datasetActions.fetchAreaOfInterestData(store, {
        dataset: dataset,
        filterParams: clonedFilterParamsExclude,
        highlights: highlights,
        dataMode: dataMode,
        include: true,
        mutatorIsInclude: true,
        isExclude: true,
      }), // exclude inner tiles
      datasetActions.fetchAreaOfInterestData(store, {
        dataset: dataset,
        filterParams: baseline,
        highlights: [],
        dataMode: dataMode,
        include: true,
        mutatorIsInclude: false,
        isExclude: true,
      }), // exclude outer tiles
    ]);
  },
  clearHighlight(context: ViewContext) {
    datasetMutations.setBaselineIncludeTableData(store, null);
    datasetMutations.setBaselineExcludeTableData(store, null);
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
    fetchClusters(context, { dataset });
    fetchOutliers(context, { dataset });

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
  async updateResultAreaOfInterest(context: ViewContext, filter: Filter) {
    // fetch new state
    const dataset = routeGetters.getRouteDataset(store);
    const solutionId = routeGetters.getRouteSolutionId(store);
    const highlights = routeGetters.getDecodedHighlights(store);
    const dataMode = context.getters.getDataMode;
    const size = routeGetters.getRouteDataSize(store);

    return Promise.all([
      resultActions.fetchAreaOfInterestInner(store, {
        dataset: dataset,
        solutionId: solutionId,
        highlights: highlights,
        dataMode: dataMode,
        size,
        filter: filter,
      }),
      resultActions.fetchAreaOfInterestOuter(store, {
        dataset: dataset,
        solutionId: solutionId,
        highlights: highlights,
        dataMode: dataMode,
        size,
        filter: filter,
      }),
    ]);
  },
  async updatePredictionAreaOfInterest(context: ViewContext, filter: Filter) {
    // fetch new state
    const dataset = routeGetters.getRouteDataset(store);
    const produceRequestId = routeGetters.getRouteProduceRequestId(store);
    const highlights = routeGetters.getDecodedHighlights(store);
    const size = routeGetters.getRouteDataSize(store);

    return Promise.all([
      predictionActions.fetchAreaOfInterestInner(store, {
        dataset: dataset,
        produceRequestId,
        highlights: highlights,
        size,
        filter: filter,
      }),
      predictionActions.fetchAreaOfInterestOuter(store, {
        dataset: dataset,
        produceRequestId,
        highlights: highlights,
        size,
        filter: filter,
      }),
    ]);
  },
  updateResultsSummaries(context: ViewContext) {
    const dataset = routeGetters.getRouteDataset(store);
    const trainingVariables = requestGetters.getActiveSolutionTrainingVariables(
      store
    );
    const highlights = routeGetters.getDecodedHighlights(store);
    const dataMode = context.getters.getDataMode;
    const varModes: Map<string, SummaryMode> = routeGetters.getDecodedVarModes(
      store
    );
    const solutionId = routeGetters.getRouteSolutionId(store);
    const currentRoute = routeGetters.getRoutePath(store);
    const pages = routeGetters.getAllRoutePages(store);
    let currentPageIndexes = [];
    if (pages[currentRoute]) {
      currentPageIndexes = pages[currentRoute];
    }
    const page = currentPageIndexes?.[0];
    const pageSize = NUM_PER_PAGE;
    const activeTrainingVariables = filterArrayByPage(
      page,
      pageSize,
      sortVariablesByImportance(trainingVariables)
    );

    resultActions.fetchTrainingSummaries(store, {
      dataset: dataset,
      training: activeTrainingVariables,
      solutionId: solutionId,
      highlights: highlights,
      dataMode: dataMode,
      varModes: varModes,
    });
  },
  async updateResultBaseline(context: ViewContext) {
    // fetch new state
    const dataset = routeGetters.getRouteDataset(store);
    const solutionId = routeGetters.getRouteSolutionId(store);
    const dataMode = context.getters.getDataMode;
    // baseline geo data for the map
    const allData = Number.MAX_SAFE_INTEGER;
    resultActions.fetchIncludedResultTableData(store, {
      dataset: dataset,
      solutionId: solutionId,
      highlights: [],
      dataMode: dataMode,
      isMapData: true,
      size: allData,
    });
  },
  updateResultSummaries(context: ViewContext, args: { requestIds: string[] }) {
    // fetch new state
    const dataset = routeGetters.getRouteDataset(store);
    const target = routeGetters.getRouteTargetVariable(store);
    const requestIds = args.requestIds;
    const solutionId = routeGetters.getRouteSolutionId(store);
    const highlights = routeGetters.getDecodedHighlights(store);
    const dataMode = context.getters.getDataMode;
    const varModes: Map<string, SummaryMode> = routeGetters.getDecodedVarModes(
      store
    );

    resultActions.fetchPredictedSummaries(store, {
      dataset: dataset,
      target: target,
      requestIds: requestIds,
      highlights: highlights,
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
        highlights: highlights,
        dataMode: dataMode,
        varModes: varModes,
      });
    } else if (task.includes(TaskTypes.CLASSIFICATION)) {
      resultActions.fetchCorrectnessSummaries(store, {
        dataset: dataset,
        target: target,
        requestIds: requestIds,
        highlights: highlights,
        dataMode: dataMode,
        varModes: varModes,
      });

      resultActions.fetchConfidenceSummaries(store, {
        dataset: dataset,
        target: target,
        requestIds: requestIds,
        highlights: highlights,
        dataMode: dataMode,
        varModes: varModes,
      });
      resultActions.fetchRankingSummaries(store, {
        dataset: dataset,
        target: target,
        requestIds: requestIds,
        highlights: highlights,
        dataMode: dataMode,
        varModes: varModes,
      });
    } else {
      console.error(`unhandled task type ${task}`);
    }
  },
  async updateResultsSolution(context: ViewContext) {
    // fetch new state
    const dataset = routeGetters.getRouteDataset(store);
    const target = routeGetters.getRouteTargetVariable(store);
    const openSolutions = new Map(
      routeGetters.getRouteOpenSolutions(store).map((s) => {
        return [s, true];
      })
    );
    // filters requests out that errored
    const requestIds = filterBadRequests(
      requestGetters.getSolutions(store),
      requestGetters.getRelevantSolutionRequestIds(store).filter((r) => {
        return openSolutions.has(r);
      })
    );
    const solutionId = routeGetters.getRouteSolutionId(store);
    const highlights = routeGetters.getDecodedHighlights(store);
    const dataMode = context.getters.getDataMode;
    const varModes: Map<string, SummaryMode> = routeGetters.getDecodedVarModes(
      store
    );
    const size = routeGetters.getRouteDataSize(store);
    resultActions.fetchResultTableData(store, {
      dataset: dataset,
      solutionId: solutionId,
      highlights: highlights,
      dataMode: dataMode,
      isMapData: false,
      size,
    });
    resultActions.fetchTargetSummary(store, {
      dataset: dataset,
      target: target,
      solutionId: solutionId,
      highlights: highlights,
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
      highlights: highlights,
      dataMode: dataMode,
      varModes: varModes,
    });
    resultActions.fetchFeatureImportanceRanking(store, {
      solutionID: solutionId,
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
        highlights: highlights,
        dataMode: dataMode,
        varModes: varModes,
      });
    } else if (task.includes(TaskTypes.CLASSIFICATION)) {
      resultActions.fetchCorrectnessSummaries(store, {
        dataset: dataset,
        target: target,
        requestIds: requestIds,
        highlights: highlights,
        dataMode: dataMode,
        varModes: varModes,
      });

      resultActions.fetchConfidenceSummaries(store, {
        dataset: dataset,
        target: target,
        requestIds: requestIds,
        highlights: highlights,
        dataMode: dataMode,
        varModes: varModes,
      });
      resultActions.fetchRankingSummaries(store, {
        dataset: dataset,
        target: target,
        requestIds: requestIds,
        highlights: highlights,
        dataMode: dataMode,
        varModes: varModes,
      });
    } else {
      console.error(`unhandled task type ${task}`);
    }
  },

  async fetchPredictionsData(context: ViewContext) {
    const produceRequestId = context.getters.getRouteProduceRequestId as string;
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
    return actions.updatePredictions(context, { isInit: true });
  },

  updatePredictionTrainingSummaries(context: ViewContext) {
    // fetch new state
    const produceRequestId = context.getters.getRouteProduceRequestId as string;
    const inferenceDataset = getPredictionsById(
      context.getters.getPredictions,
      produceRequestId
    ).dataset;
    const highlights = context.getters.getDecodedHighlights as Highlight[];
    const varModes = context.getters.getDecodedVarModes as Map<
      string,
      SummaryMode
    >;
    const currentSearch = context.getters
      .getRouteResultTrainingVarsSearch as string;
    const trainingVariables = searchVariables(
      context.getters.getActivePredictionTrainingVariables,
      currentSearch
    ) as Variable[];
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
      highlights: highlights,
      varModes: varModes,
      produceRequestId: produceRequestId,
    });
  },
  updateBaselinePredictions(context: ViewContext) {
    const produceRequestId = context.getters.getRouteProduceRequestId as string;
    const allData = Number.MAX_SAFE_INTEGER;
    const inferenceDataset = getPredictionsById(
      context.getters.getPredictions,
      produceRequestId
    ).dataset;
    predictionActions.fetchPredictionTableData(store, {
      dataset: inferenceDataset,
      highlights: [],
      produceRequestId: produceRequestId,
      size: allData,
      isBaseline: true,
    });
  },
  updatePredictionSummaries(
    context: ViewContext,
    args: { predictions: Predictions[] }
  ) {
    const fittedSolutionId = context.getters.getRouteFittedSolutionId as string;
    const highlights = context.getters.getDecodedHighlights as Highlight[];
    const dataMode = context.getters.getDataMode;
    const varMode = SummaryMode.Default;
    // this is where rank and confidence should get updated
    predictionActions.fetchPredictedSummaries(store, {
      highlights: highlights,
      fittedSolutionId: fittedSolutionId,
      predictions: args.predictions,
    });

    args.predictions.forEach((p) => {
      predictionActions.fetchConfidenceSummary(store, {
        dataset: p.dataset,
        highlights: highlights,
        solutionId: p.resultId,
        dataMode,
        varMode,
        target: p.feature,
      });
      predictionActions.fetchRankSummary(store, {
        dataset: p.dataset,
        highlights: highlights,
        solutionId: p.resultId,
        dataMode,
        varMode,
        target: p.feature,
      });
    });
  },
  updatePredictions(context: ViewContext, args?: { isInit: boolean }) {
    // fetch new state
    const produceRequestId = context.getters.getRouteProduceRequestId as string;
    const fittedSolutionId = context.getters.getRouteFittedSolutionId as string;
    const pred = getPredictionsById(
      context.getters.getPredictions,
      produceRequestId
    );
    const inferenceDataset = pred.dataset;
    const highlights = context.getters.getDecodedHighlights as Highlight[];
    const size = routeGetters.getRouteDataSize(store);
    const dataMode = context.getters.getDataMode;
    const varMode = SummaryMode.Default;
    const openPredictions = new Map(
      routeGetters.getRouteOpenSolutions(store).map((s) => {
        return [s, true];
      })
    );
    const relPreds = requestGetters
      .getRelevantPredictions(store)
      .filter((p) => {
        return openPredictions.has(p.requestId) || (args?.isInit ?? false);
      });
    // table data
    predictionActions.fetchPredictionTableData(store, {
      dataset: inferenceDataset,
      highlights: highlights,
      produceRequestId: produceRequestId,
      size,
      isBaseline: false,
    });
    // variable summaries
    actions.updatePredictionTrainingSummaries(context);
    // this is where rank and confidence should get updated
    predictionActions.fetchPredictedSummaries(store, {
      highlights: highlights,
      fittedSolutionId: fittedSolutionId,
      predictions: relPreds,
    });

    relPreds.forEach((p) => {
      predictionActions.fetchConfidenceSummary(store, {
        dataset: p.dataset,
        highlights: highlights,
        solutionId: p.resultId,
        dataMode,
        varMode,
        target: p.feature,
      });
      predictionActions.fetchRankSummary(store, {
        dataset: p.dataset,
        highlights: highlights,
        solutionId: p.resultId,
        dataMode,
        varMode,
        target: p.feature,
      });
    });
  },
};
