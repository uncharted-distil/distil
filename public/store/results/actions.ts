import axios from "axios";
import _ from "lodash";
import { ActionContext } from "vuex";
import store, { DistilState } from "../store";
import { EXCLUDE_FILTER } from "../../util/filters";
import {
  getSolutionById,
  getSolutionsBySolutionRequestIds
} from "../../util/solutions";
import { Variable, Highlight, SummaryMode } from "../dataset/index";
import { mutations } from "./module";
import { ResultsState } from "./index";
import { addHighlightToFilterParams } from "../../util/highlights";
import {
  fetchSolutionResultSummary,
  createPendingSummary,
  createErrorSummary,
  createEmptyTableData,
  fetchSummaryExemplars
} from "../../util/data";
import { getters as resultGetters } from "../results/module";
import { getters as dataGetters } from "../dataset/module";

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

    const dataset = args.dataset;
    const solutionId = args.solutionId;
    const varModes = args.varModes;

    const promises = [];

    // remove summaries not used to predict the newly selected model
    context.state.trainingSummaries.forEach(v => {
      const isTrainingArg = args.training.reduce((isTrain, variable) => {
        if (!isTrain) {
          isTrain = variable.colName === v.key;
        }
        return isTrain;
      }, false);
      if (v.dataset !== args.dataset || !isTrainingArg) {
        mutations.removeTrainingSummary(context, v);
      }
    });

    args.training.forEach(variable => {
      const key = variable.colName;
      const label = variable.colDisplayName;
      const description = variable.colDescription;
      const exists = _.find(context.state.trainingSummaries, v => {
        return v.dataset === args.dataset && v.key === variable.colName;
      });
      if (!exists) {
        // add placeholder
        mutations.updateTrainingSummary(
          context,
          createPendingSummary(key, label, description, dataset)
        );
      }
      // fetch summary
      promises.push(
        actions.fetchTrainingSummary(context, {
          dataset: dataset,
          variable: variable,
          resultID: solution.resultId,
          highlight: args.highlight,
          varMode: varModes.has(variable.colName)
            ? varModes.get(variable.colName)
            : SummaryMode.Default
        })
      );
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

    let filterParams = {
      highlight: null,
      variables: [],
      filters: []
    };
    filterParams = addHighlightToFilterParams(filterParams, args.highlight);
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

    let filterParams = {
      highlight: null,
      variables: [],
      filters: []
    };
    filterParams = addHighlightToFilterParams(filterParams, args.highlight);
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
    args: { solutionId: string; dataset: string; highlight: Highlight }
  ) {
    const solution = getSolutionById(
      context.rootState.requestsModule.solutions,
      args.solutionId
    );
    if (!solution || !solution.resultId) {
      // no results ready to pull
      return null;
    }

    let filterParams = {
      highlight: null,
      variables: [],
      filters: []
    };
    filterParams = addHighlightToFilterParams(filterParams, args.highlight);

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
    args: { solutionId: string; dataset: string; highlight: Highlight }
  ) {
    const solution = getSolutionById(
      context.rootState.requestsModule.solutions,
      args.solutionId
    );
    if (!solution || !solution.resultId) {
      // no results ready to pull
      return null;
    }

    let filterParams = {
      highlight: null,
      variables: [],
      filters: []
    };
    filterParams = addHighlightToFilterParams(
      filterParams,
      args.highlight,
      EXCLUDE_FILTER
    );

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
    args: { solutionId: string; dataset: string; highlight: Highlight }
  ) {
    return Promise.all([
      actions.fetchIncludedResultTableData(context, {
        dataset: args.dataset,
        solutionId: args.solutionId,
        highlight: args.highlight
      }),
      actions.fetchExcludedResultTableData(context, {
        dataset: args.dataset,
        solutionId: args.solutionId,
        highlight: args.highlight
      })
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

    let filterParams = {
      highlight: null,
      variables: [],
      filters: []
    };
    filterParams = addHighlightToFilterParams(filterParams, args.highlight);
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
      solutions.map(solution => {
        return actions.fetchPredictedSummary(context, {
          dataset: args.dataset,
          target: args.target,
          solutionId: solution.solutionId,
          highlight: args.highlight,
          varMode: args.varModes.has(args.target)
            ? args.varModes.get(args.target)
            : SummaryMode.Default
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

    let filterParams = {
      highlight: null,
      variables: [],
      filters: []
    };
    filterParams = addHighlightToFilterParams(filterParams, args.highlight);

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
      solutions.map(solution => {
        return actions.fetchResidualsSummary(context, {
          dataset: args.dataset,
          target: args.target,
          solutionId: solution.solutionId,
          highlight: args.highlight,
          varMode: args.varModes.has(args.target)
            ? args.varModes.get(args.target)
            : SummaryMode.Default
        });
      })
    );
  },

  // fetches result summary for a given pipeline id.
  fetchCorrectnessSummary(
    context: ResultsContext,
    args: {
      dataset: string;
      target: string;
      solutionId: string;
      highlight: Highlight;
    }
  ) {
    if (!args.dataset) {
      console.warn("`dataset` argument is missing");
      return null;
    }
    if (!args.solutionId) {
      console.warn("`pipelineId` argument is missing");
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

    let filterParams = {
      highlight: null,
      variables: [],
      filters: []
    };
    filterParams = addHighlightToFilterParams(filterParams, args.highlight);

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
      SummaryMode.Default
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
      solutions.map(solution => {
        return actions.fetchCorrectnessSummary(context, {
          dataset: args.dataset,
          target: args.target,
          solutionId: solution.solutionId,
          highlight: args.highlight
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
      timeseriesID: any;
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
    if (!args.timeseriesID) {
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
        `distil/timeseries-forecast/${encodeURIComponent(
          args.dataset
        )}/${encodeURIComponent(args.timeseriesColName)}/${encodeURIComponent(
          args.xColName
        )}/${encodeURIComponent(args.yColName)}/${encodeURIComponent(
          args.timeseriesID
        )}/${encodeURIComponent(solution.resultId)}`,
        {}
      );
      mutations.updatePredictedTimeseries(context, {
        solutionId: args.solutionId,
        id: args.timeseriesID,
        timeseries: response.data.timeseries,
        isDateTime: response.data.isDateTime
      });
      mutations.updatePredictedForecast(context, {
        solutionId: args.solutionId,
        id: args.timeseriesID,
        forecast: response.data.forecast,
        forecastTestRange: response.data.forecastTestRange,
        isDateTime: response.data.isDateTime
      });
    } catch (error) {
      console.error(error);
    }
  }
};
