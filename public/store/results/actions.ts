import axios from "axios";
import _ from "lodash";
import { ActionContext } from "vuex";
import store, { DistilState } from "../store";
import { EXCLUDE_FILTER, FilterParams } from "../../util/filters";
import {
  getSolutionById,
  getSolutionsBySolutionRequestIds,
} from "../../util/solutions";
import {
  Variable,
  Highlight,
  SummaryMode,
  DataMode,
  VariableSummary,
} from "../dataset/index";
import { mutations } from "./module";
import { ResultsState } from "./index";
import { addHighlightToFilterParams } from "../../util/highlights";
import {
  fetchSolutionResultSummary,
  createPendingSummary,
  createErrorSummary,
  createEmptyTableData,
  fetchSummaryExemplars,
  validateArgs,
  minimumRouteKey,
} from "../../util/data";
import { getters as resultGetters } from "../results/module";
import { getters as dataGetters } from "../dataset/module";
import { Dictionary } from "vue-router/types/router";

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
      mutations.setIncludedResultTableData(context, response.data);
    } catch (error) {
      console.error(
        `Failed to fetch results from ${args.solutionId} with error ${error}`
      );
      mutations.setIncludedResultTableData(context, createEmptyTableData());
    }
  },

  async fetchExcludedResultTableData(
    context: ResultsContext,
    args: {
      solutionId: string;
      dataset: string;
      highlight: Highlight;
      dataMode: DataMode;
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
      mutations.setExcludedResultTableData(context, response.data);
    } catch (error) {
      console.error(
        `Failed to fetch results from ${args.solutionId} with error ${error}`
      );
      mutations.setExcludedResultTableData(context, createEmptyTableData());
    }
  },

  fetchResultTableData(
    context: ResultsContext,
    args: {
      solutionId: string;
      dataset: string;
      highlight: Highlight;
      dataMode: DataMode;
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

  async fetchForecastedTimeseries(
    context: ResultsContext,
    args: {
      dataset: string;
      xColName: string;
      yColName: string;
      timeseriesColName: string;
      timeseriesId: any;
      solutionId: string;
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
    if (!args.timeseriesColName) {
      console.warn("`timeseriesColName` argument is missing");
      return null;
    }
    if (!args.timeseriesId) {
      console.warn("`timeseriesID` argument is missing");
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
          `${encodeURIComponent(args.timeseriesId)}/` +
          `${encodeURIComponent(solution.resultId)}`,
        {}
      );
      mutations.updatePredictedTimeseries(context, {
        solutionId: args.solutionId,
        id: args.timeseriesId,
        timeseries: response.data.timeseries,
        isDateTime: response.data.isDateTime,
      });
      mutations.updatePredictedForecast(context, {
        solutionId: args.solutionId,
        id: args.timeseriesId,
        forecast: response.data.forecast,
        forecastTestRange: response.data.forecastTestRange,
        isDateTime: response.data.isDateTime,
      });
    } catch (error) {
      console.error(error);
    }
  },

  // Fetch variable rankings associated with a computed solution.  If the solution results are
  // available, then the rankings will have been computed.
  async fetchVariableRankings(
    context: ResultsContext,
    args: { solutionID: string }
  ) {
    const response = await axios.get(
      `/distil/solution-variable-rankings/${args.solutionID}`
    );
    const rankings = <Dictionary<number>>response.data;
    mutations.setVariableRankings(store, {
      solutionID: args.solutionID,
      rankings: _.pickBy(rankings, (ranking) => ranking !== null),
    });
  },
};
