import axios from "axios";
import _ from "lodash";
import { Dictionary } from "vue-router/types/router";
import { ActionContext } from "vuex";
import {
  createEmptyTableData,
  createErrorSummary,
  createPendingSummary,
  fetchSolutionResultSummary,
  fetchSummaryExemplars,
  minimumRouteKey,
  validateArgs,
} from "../../util/data";
import { EXCLUDE_FILTER, Filter } from "../../util/filters";
import { addHighlightToFilterParams } from "../../util/highlights";
import {
  getSolutionById,
  getSolutionsBySolutionRequestIds,
} from "../../util/solutions";
import { DataMode, Highlight, SummaryMode, Variable } from "../dataset/index";
import { getters as dataGetters } from "../dataset/module";
import { getters as resultGetters } from "../results/module";
import store, { DistilState } from "../store";
import { ResultsState } from "./index";
import { mutations } from "./module";

export type ResultsContext = ActionContext<ResultsState, DistilState>;

export const actions = {
  // fetches variable summary data for the given dataset and variables
  fetchTrainingSummaries(
    context: ResultsContext,
    args: {
      dataset: string;
      training: Variable[];
      solutionId: string;
      highlight: Highlight;
      dataMode: DataMode;
      varModes: Map<string, SummaryMode>;
    }
  ) {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.training) {
      console.warn("`training` argument is missing");
      return null;
    }
    if (!args.solutionId) {
      console.warn("`solutionId` argument is missing");
      return null;
    }
    if (!args.varModes) {
      console.warn("`varModes` argument is missing");
      return null;
    }

    const solution = getSolutionById(
      context.rootState.requestsModule.solutions,
      args.solutionId
    );
    if (!solution || !solution.resultId) {
      // no results ready to pull
      return;
    }

    const dataMode = args.dataMode ? args.dataMode : DataMode.Default;

    const promises = [];

    const summariesByVariable = context.state.trainingSummaries;
    const routeKey = minimumRouteKey();

    args.training.forEach((variable) => {
      const existingVariableSummary =
        summariesByVariable?.[variable.colName]?.[routeKey];

      if (existingVariableSummary) {
        promises.push(existingVariableSummary);
      } else {
        if (summariesByVariable[variable.colName]) {
          // if we have any saved state for that variable
          // use that as placeholder due to vue lifecycle
          const tempVariableSummaryKey = Object.keys(
            summariesByVariable[variable.colName]
          )[0];
          promises.push(
            summariesByVariable[variable.colName][tempVariableSummaryKey]
          );
        } else {
          // add a loading placeholder if nothing exists for that variable
          createPendingSummary(
            variable.colName,
            variable.colDisplayName,
            variable.colDescription,
            args.dataset
          );
        }

        // fetch summary
        promises.push(
          actions.fetchTrainingSummary(context, {
            dataset: args.dataset,
            variable: variable,
            resultID: solution.resultId,
            highlight: args.highlight,
            dataMode: dataMode,
            varMode: args.varModes.has(variable.colName)
              ? args.varModes.get(variable.colName)
              : SummaryMode.Default,
          })
        );
      }
    });
    return Promise.all(promises);
  },

  async fetchTrainingSummary(
    context: ResultsContext,
    args: {
      dataset: string;
      variable: Variable;
      resultID: string;
      highlight: Highlight;
      dataMode: DataMode;
      varMode: SummaryMode;
    }
  ): Promise<void> {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.variable) {
      console.warn("`variable` argument is missing");
      return null;
    }
    if (!args.resultID) {
      console.warn("`resultID` argument is missing");
      return null;
    }

    const filterParamsBlank = {
      highlight: null,
      variables: [],
      filters: [],
    };
    const filterParams = addHighlightToFilterParams(
      filterParamsBlank,
      args.highlight
    );

    const dataModeDefault = args.dataMode ? args.dataMode : DataMode.Default;
    filterParams.dataMode = dataModeDefault;

    try {
      const response = await axios.post(
        `/distil/training-summary/${args.dataset}/${args.variable.colName}/${args.resultID}/${args.varMode}`,
        filterParams
      );
      const summary = response.data.summary;
      await fetchSummaryExemplars(args.dataset, args.variable.colName, summary);
      mutations.updateTrainingSummary(context, summary);
    } catch (error) {
      console.error(error);
      mutations.updateTrainingSummary(
        context,
        createErrorSummary(
          args.variable.colName,
          args.variable.colDisplayName,
          args.dataset,
          error
        )
      );
    }
  },

  async fetchTargetSummary(
    context: ResultsContext,
    args: {
      dataset: string;
      target: string;
      solutionId: string;
      highlight: Highlight;
      dataMode: DataMode;
      varMode: SummaryMode;
    }
  ) {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.target) {
      console.warn("`variable` argument is missing");
      return null;
    }
    if (!args.solutionId) {
      console.warn("`solutionId` argument is missing");
      return null;
    }
    if (!args.varMode) {
      console.warn("`varMode` argument is missing");
      return null;
    }
    const solution = getSolutionById(
      context.rootState.requestsModule.solutions,
      args.solutionId
    );
    if (!solution || !solution.resultId) {
      // no results ready to pull
      return null;
    }

    const key = args.target;
    const label = args.target;
    const dataset = args.dataset;

    if (!context.state.targetSummary) {
      // fetch the target var so we can pull the description out
      const targetVar = dataGetters.getVariablesMap(store)[args.target];
      mutations.updateTargetSummary(
        context,
        createPendingSummary(key, label, targetVar.colDescription, dataset)
      );
    }

    const filterParamsBlank = {
      highlight: null,
      variables: [],
      filters: [],
    };
    const filterParams = addHighlightToFilterParams(
      filterParamsBlank,
      args.highlight
    );

    const dataModeDefault = args.dataMode ? args.dataMode : DataMode.Default;
    filterParams.dataMode = dataModeDefault;

    try {
      const response = await axios.post(
        `/distil/target-summary/${args.dataset}/${args.target}/${solution.resultId}/${args.varMode}`,
        filterParams
      );
      const summary = response.data.summary;
      await fetchSummaryExemplars(args.dataset, args.target, summary);
      mutations.updateTargetSummary(context, summary);
    } catch (error) {
      console.error(error);
      mutations.updateTargetSummary(
        context,
        createErrorSummary(key, label, dataset, error)
      );
    }
  },

  async fetchIncludedResultTableData(
    context: ResultsContext,
    args: {
      solutionId: string;
      dataset: string;
      highlight: Highlight;
      dataMode: DataMode;
      isMapData: boolean;
      size?: number;
    }
  ) {
    const solution = getSolutionById(
      context.rootState.requestsModule.solutions,
      args.solutionId
    );
    if (!solution || !solution.resultId) {
      // no results ready to pull
      return null;
    }

    const filterParamsBlank = {
      highlight: null,
      variables: [],
      filters: [],
    };
    const filterParams = addHighlightToFilterParams(
      filterParamsBlank,
      args.highlight
    );
    const mutator = args.isMapData
      ? mutations.setFullIncludedResultTableData
      : mutations.setIncludedResultTableData;
    const dataModeDefault = args.dataMode ? args.dataMode : DataMode.Default;
    filterParams.dataMode = dataModeDefault; // Add the size limit to results if provided.
    if (_.isInteger(args.size)) {
      filterParams.size = args.size;
    }

    try {
      const response = await axios.post(
        `/distil/results/${args.dataset}/${encodeURIComponent(
          args.solutionId
        )}`,
        filterParams
      );
      mutator(context, response.data);
    } catch (error) {
      console.error(
        `Failed to fetch results from ${args.solutionId} with error ${error}`
      );
      mutator(context, createEmptyTableData());
    }
  },

  async fetchExcludedResultTableData(
    context: ResultsContext,
    args: {
      solutionId: string;
      dataset: string;
      highlight: Highlight;
      dataMode: DataMode;
      isMapData: boolean;
      size?: number;
    }
  ) {
    const solution = getSolutionById(
      context.rootState.requestsModule.solutions,
      args.solutionId
    );
    if (!solution || !solution.resultId) {
      // no results ready to pull
      return null;
    }

    const filterParamsBlank = {
      highlight: null,
      variables: [],
      filters: [],
    };
    const filterParams = addHighlightToFilterParams(
      filterParamsBlank,
      args.highlight,
      EXCLUDE_FILTER
    );

    const dataModeDefault = args.dataMode ? args.dataMode : DataMode.Default;
    filterParams.dataMode = dataModeDefault;
    const mutator = args.isMapData
      ? mutations.setFullExcludedResultTableData
      : mutations.setExcludedResultTableData;
    // Add the size limit to results if provided.
    if (_.isInteger(args.size)) {
      filterParams.size = args.size;
    }

    try {
      const response = await axios.post(
        `/distil/results/${args.dataset}/${encodeURIComponent(
          args.solutionId
        )}`,
        filterParams
      );
      mutator(context, response.data);
    } catch (error) {
      console.error(
        `Failed to fetch results from ${args.solutionId} with error ${error}`
      );
      mutator(context, createEmptyTableData());
    }
  },
  // fetches
  async fetchAreaOfInterestInner(
    context: ResultsContext,
    args: {
      solutionId: string;
      dataset: string;
      highlight: Highlight;
      dataMode: DataMode;
      size?: number;
      filter: Filter; // the area of interest
    }
  ) {
    const solution = getSolutionById(
      context.rootState.requestsModule.solutions,
      args.solutionId
    );
    if (!solution || !solution.resultId) {
      // no results ready to pull
      return null;
    }

    const filterParamsBlank = {
      highlight: null,
      variables: [],
      filters: [],
    };
    const filterParams = addHighlightToFilterParams(
      filterParamsBlank,
      args.highlight
    );

    const dataModeDefault = args.dataMode ? args.dataMode : DataMode.Default;
    filterParams.dataMode = dataModeDefault; // Add the size limit to results if provided.
    if (_.isInteger(args.size)) {
      filterParams.size = args.size;
    }
    filterParams.filters.push(args.filter);
    try {
      const response = await axios.post(
        `/distil/results/${args.dataset}/${encodeURIComponent(
          args.solutionId
        )}`,
        filterParams
      );
      mutations.setAreaOfInterestInner(context, response.data);
    } catch (error) {
      console.error(
        `Failed to fetch results from ${args.solutionId} with error ${error}`
      );
      mutations.setAreaOfInterestInner(context, createEmptyTableData());
    }
  },
  // fetches the tiles that are within the bounds but are filtered by another highlight
  async fetchAreaOfInterestOuter(
    context: ResultsContext,
    args: {
      solutionId: string;
      dataset: string;
      highlight: Highlight;
      dataMode: DataMode;
      size?: number;
      filter: Filter;
    }
  ) {
    const solution = getSolutionById(
      context.rootState.requestsModule.solutions,
      args.solutionId
    );
    if (!solution || !solution.resultId) {
      // no results ready to pull
      return null;
    }

    const filterParamsBlank = {
      highlight: null,
      variables: [],
      filters: [],
    };
    const filterParams = addHighlightToFilterParams(
      filterParamsBlank,
      args.highlight,
      EXCLUDE_FILTER
    );

    const dataModeDefault = args.dataMode ? args.dataMode : DataMode.Default;
    filterParams.dataMode = dataModeDefault;
    // Add the size limit to results if provided.
    if (_.isInteger(args.size)) {
      filterParams.size = args.size;
    }
    filterParams.filters.push(args.filter);
    // if highlight is null there is nothing to invert so return null
    if (filterParams.highlight === null) {
      mutations.setAreaOfInterestOuter(context, createEmptyTableData());
      return;
    }
    try {
      const response = await axios.post(
        `/distil/results/${args.dataset}/${encodeURIComponent(
          args.solutionId
        )}`,
        filterParams
      );
      mutations.setAreaOfInterestOuter(context, response.data);
    } catch (error) {
      console.error(
        `Failed to fetch results from ${args.solutionId} with error ${error}`
      );
      mutations.setAreaOfInterestOuter(context, createEmptyTableData());
    }
  },
  fetchResultTableData(
    context: ResultsContext,
    args: {
      solutionId: string;
      dataset: string;
      highlight: Highlight;
      dataMode: DataMode;
      isMapData: boolean;
      size?: number;
    }
  ) {
    return Promise.all([
      actions.fetchIncludedResultTableData(context, args),
      actions.fetchExcludedResultTableData(context, args),
    ]);
  },

  async fetchResidualsExtrema(
    context: ResultsContext,
    args: { dataset: string; target: string; solutionId: string }
  ) {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.target) {
      console.warn("`target` argument is missing");
      return null;
    }

    const solution = getSolutionById(
      context.rootState.requestsModule.solutions,
      args.solutionId
    );
    if (!solution || !solution.resultId) {
      // no results ready to pull
      return null;
    }

    try {
      const response = await axios.get(
        `/distil/residuals-extrema/${args.dataset}/${args.target}`
      );
      mutations.updateResidualsExtrema(context, response.data.extrema);
    } catch (error) {
      console.error(error);
    }
  },

  // fetches result summary for a given solution id.
  fetchPredictedSummary(
    context: ResultsContext,
    args: {
      dataset: string;
      target: string;
      solutionId: string;
      highlight: Highlight;
      dataMode: DataMode;
      varMode: SummaryMode;
    }
  ) {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.target) {
      console.warn("`target` argument is missing");
      return null;
    }
    if (!args.solutionId) {
      console.warn("`solutionId` argument is missing");
      return null;
    }
    if (!args.varMode) {
      console.warn("`varMode` argument is missing");
      return null;
    }

    const solution = getSolutionById(
      context.rootState.requestsModule.solutions,
      args.solutionId
    );
    if (!solution || !solution.resultId) {
      // no results ready to pull
      return null;
    }

    const filterParamsBlank = {
      highlight: null,
      variables: [],
      filters: [],
    };
    const filterParams = addHighlightToFilterParams(
      filterParamsBlank,
      args.highlight
    );

    const dataModeDefault = args.dataMode ? args.dataMode : DataMode.Default;
    filterParams.dataMode = dataModeDefault;

    const endpoint = `/distil/solution-result-summary`;
    const key = solution.predictedKey;
    const label = "Predicted";
    return fetchSolutionResultSummary(
      context,
      endpoint,
      solution,
      key,
      label,
      resultGetters.getPredictedSummaries(context),
      mutations.updatePredictedSummaries,
      filterParams,
      args.varMode
    );
  },

  // fetches result summaries for a given solution create request
  fetchPredictedSummaries(
    context: ResultsContext,
    args: {
      dataset: string;
      target: string;
      requestIds: string[];
      highlight: Highlight;
      dataMode: DataMode;
      varModes: Map<string, SummaryMode>;
    }
  ) {
    if (!args.requestIds) {
      console.warn("`requestIds` argument is missing");
      return null;
    }
    const solutions = getSolutionsBySolutionRequestIds(
      context.rootState.requestsModule.solutions,
      args.requestIds
    );
    return Promise.all(
      solutions.map((solution) => {
        return actions.fetchPredictedSummary(context, {
          dataset: args.dataset,
          target: args.target,
          solutionId: solution.solutionId,
          highlight: args.highlight,
          dataMode: args.dataMode,
          varMode: args.varModes.has(args.target)
            ? args.varModes.get(args.target)
            : SummaryMode.Default,
        });
      })
    );
  },

  // fetches result summary for a given solution id.
  fetchResidualsSummary(
    context: ResultsContext,
    args: {
      dataset: string;
      target: string;
      solutionId: string;
      highlight: Highlight;
      dataMode: DataMode;
      varMode: SummaryMode;
    }
  ) {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.target) {
      console.warn("`target` argument is missing");
      return null;
    }
    if (!args.solutionId) {
      console.warn("`solutionId` argument is missing");
      return null;
    }
    if (!args.varMode) {
      console.warn("`varMode` argument is missing");
      return null;
    }

    const solution = getSolutionById(
      context.rootState.requestsModule.solutions,
      args.solutionId
    );
    if (!solution.resultId) {
      // no results ready to pull
      return null;
    }

    const filterParamsBlank = {
      highlight: null,
      variables: [],
      filters: [],
    };
    const filterParams = addHighlightToFilterParams(
      filterParamsBlank,
      args.highlight
    );

    const dataModeDefault = args.dataMode ? args.dataMode : DataMode.Default;
    filterParams.dataMode = dataModeDefault;

    const endPoint = `/distil/residuals-summary/${args.dataset}/${args.target}`;
    const key = solution.errorKey;
    const label = "Error";
    return fetchSolutionResultSummary(
      context,
      endPoint,
      solution,
      key,
      label,
      resultGetters.getResidualsSummaries(context),
      mutations.updateResidualsSummaries,
      filterParams,
      args.varMode
    );
  },

  // fetches result summaries for a given solution create request
  fetchResidualsSummaries(
    context: ResultsContext,
    args: {
      dataset: string;
      target: string;
      requestIds: string[];
      highlight: Highlight;
      dataMode: DataMode;
      varModes: Map<string, SummaryMode>;
    }
  ) {
    if (!args.requestIds) {
      console.warn("`requestIds` argument is missing");
      return null;
    }
    const solutions = getSolutionsBySolutionRequestIds(
      context.rootState.requestsModule.solutions,
      args.requestIds
    );
    return Promise.all(
      solutions.map((solution) => {
        return actions.fetchResidualsSummary(context, {
          dataset: args.dataset,
          target: args.target,
          solutionId: solution.solutionId,
          highlight: args.highlight,
          dataMode: args.dataMode,
          varMode: args.varModes.has(args.target)
            ? args.varModes.get(args.target)
            : SummaryMode.Default,
        });
      })
    );
  },

  // fetches result summary for a given pipeline id.
  fetchCorrectnessSummary(
    context: ResultsContext,
    args: {
      dataset: string;
      solutionId: string;
      highlight: Highlight;
      dataMode: DataMode;
      varMode: SummaryMode;
    }
  ) {
    if (!validateArgs(args, ["dataset", "solutionId", "varMode"])) {
      return null;
    }

    const solution = getSolutionById(
      context.rootState.requestsModule.solutions,
      args.solutionId
    );
    if (!solution || !solution.resultId) {
      // no results ready to pull
      return null;
    }

    const filterParamsBlank = {
      highlight: null,
      variables: [],
      filters: [],
    };
    const filterParams = addHighlightToFilterParams(
      filterParamsBlank,
      args.highlight
    );

    const dataModeDefault = args.dataMode ? args.dataMode : DataMode.Default;
    filterParams.dataMode = dataModeDefault;

    const endPoint = `/distil/correctness-summary/${args.dataset}`;
    const key = solution.errorKey;
    const label = "Error";
    return fetchSolutionResultSummary(
      context,
      endPoint,
      solution,
      key,
      label,
      resultGetters.getCorrectnessSummaries(context),
      mutations.updateCorrectnessSummaries,
      filterParams,
      args.varMode
    );
  },

  // fetches result summaries for a given pipeline create request
  fetchCorrectnessSummaries(
    context: ResultsContext,
    args: {
      dataset: string;
      target: string;
      requestIds: string[];
      highlight: Highlight;
      dataMode: DataMode;
      varModes: Map<string, SummaryMode>;
    }
  ) {
    if (!validateArgs(args, ["dataset", "target", "requestIds"])) {
      return null;
    }

    const solutions = getSolutionsBySolutionRequestIds(
      context.rootState.requestsModule.solutions,
      args.requestIds
    );
    return Promise.all(
      solutions.map((solution) => {
        return actions.fetchCorrectnessSummary(context, {
          dataset: args.dataset,
          solutionId: solution.solutionId,
          highlight: args.highlight,
          dataMode: args.dataMode,
          varMode: args.varModes.has(args.target)
            ? args.varModes.get(args.target)
            : SummaryMode.Default,
        });
      })
    );
  },

  // fetches result summary for a given solution id.
  fetchConfidenceSummary(
    context: ResultsContext,
    args: {
      dataset: string;
      solutionId: string;
      highlight: Highlight;
      dataMode: DataMode;
      varMode: SummaryMode;
    }
  ) {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.solutionId) {
      console.warn("`solutionId` argument is missing");
      return null;
    }
    if (!args.varMode) {
      console.warn("`varMode` argument is missing");
      return null;
    }

    const solution = getSolutionById(
      context.rootState.requestsModule.solutions,
      args.solutionId
    );
    if (!solution || !solution.resultId) {
      // no results ready to pull
      return null;
    }

    const filterParamsBlank = {
      highlight: null,
      variables: [],
      filters: [],
    };
    const filterParams = addHighlightToFilterParams(
      filterParamsBlank,
      args.highlight
    );

    const dataModeDefault = args.dataMode ? args.dataMode : DataMode.Default;
    filterParams.dataMode = dataModeDefault;

    const endpoint = `/distil/confidence-summary/${args.dataset}`;
    const key = solution.confidenceKey;
    const label = "Confidence";
    return fetchSolutionResultSummary(
      context,
      endpoint,
      solution,
      key,
      label,
      resultGetters.getConfidenceSummaries(context),
      mutations.updateConfidenceSummaries,
      filterParams,
      args.varMode
    );
  },

  // fetches result summaries for a given solution create request
  fetchConfidenceSummaries(
    context: ResultsContext,
    args: {
      dataset: string;
      target: string;
      requestIds: string[];
      highlight: Highlight;
      dataMode: DataMode;
      varModes: Map<string, SummaryMode>;
    }
  ) {
    if (!args.requestIds) {
      console.warn("`requestIds` argument is missing");
      return null;
    }
    const solutions = getSolutionsBySolutionRequestIds(
      context.rootState.requestsModule.solutions,
      args.requestIds
    );
    return Promise.all(
      solutions.map((solution) => {
        return actions.fetchConfidenceSummary(context, {
          dataset: args.dataset,
          solutionId: solution.solutionId,
          highlight: args.highlight,
          dataMode: args.dataMode,
          varMode: args.varModes.has(args.target)
            ? args.varModes.get(args.target)
            : SummaryMode.Default,
        });
      })
    );
  },

  async fetchForecastedTimeseries(
    context: ResultsContext,
    args: {
      dataset: string;
      xColName: string;
      yColName: string;
      timeseriesColName: string;
      solutionId: string;
      timeseriesIds: string[];
      uniqueTrail?: string;
    }
  ) {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.xColName) {
      console.warn("`xColName` argument is missing");
      return null;
    }
    if (!args.yColName) {
      console.warn("`yColName` argument is missing");
      return null;
    }
    if (!args.timeseriesIds) {
      console.warn("`timeseriesIds` argument is missing");
      return null;
    }
    if (!args.timeseriesColName) {
      console.warn("`timeseriesColName` argument is missing");
      return null;
    }
    if (!args.solutionId) {
      console.warn("`solutionId` argument is missing");
      return null;
    }

    const solution = getSolutionById(
      context.rootState.requestsModule.solutions,
      args.solutionId
    );
    if (!solution || !solution.resultId) {
      // no results ready to pull
      return null;
    }

    try {
      const response = await axios.post(
        `distil/timeseries-forecast/` +
          `${encodeURIComponent(args.dataset)}/` +
          `${encodeURIComponent(args.dataset)}/` +
          `${encodeURIComponent(args.timeseriesColName)}/` +
          `${encodeURIComponent(args.xColName)}/` +
          `${encodeURIComponent(args.yColName)}/` +
          `${encodeURIComponent(solution.resultId)}`,
        {
          timeseriesUris: args.timeseriesIds,
        }
      );
      const responseMap = new Map(
        Object.keys(response.data).map((k) => {
          return [k + (args.uniqueTrail ?? ""), response.data[k]];
        })
      );
      mutations.bulkUpdatePredictedTimeseries(context, {
        solutionId: args.solutionId,
        map: responseMap,
      });
      mutations.bulkUpdatePredictedForecast(context, {
        solutionId: args.solutionId,
        map: responseMap,
      });
    } catch (error) {
      console.error(error);
    }
  },

  // Fetch variable rankings associated with a computed solution.  If the solution results are
  // available, then the rankings will have been computed.
  async fetchFeatureImportanceRanking(
    context: ResultsContext,
    args: { solutionID: string }
  ) {
    const response = await axios.get(
      `/distil/solution-variable-rankings/${args.solutionID}`
    );
    const rankings = <Dictionary<number>>response.data;
    mutations.setFeatureImportanceRanking(store, {
      solutionID: args.solutionID,
      rankings: _.pickBy(rankings, (ranking) => ranking !== null),
    });
  },
};
