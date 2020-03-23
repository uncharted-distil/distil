import axios from "axios";
import _ from "lodash";
import { ActionContext } from "vuex";
import { DistilState } from "../store";
import { EXCLUDE_FILTER } from "../../util/filters";
import { getSolutionById } from "../../util/solutions";
import { Variable, Highlight, SummaryMode } from "../dataset/index";
import { mutations } from "./module";
import { PredictionState } from "./index";
import { addHighlightToFilterParams } from "../../util/highlights";
import {
  fetchPredictionResultSummary,
  createPendingSummary,
  createErrorSummary,
  createEmptyTableData,
  fetchSummaryExemplars
} from "../../util/data";
import { getters as predictionGetters } from "../predictions/module";
import { RouteArgs } from "../../util/routes";
import { getPredictionsById } from "../../util/predictions";

export type PredictionContext = ActionContext<PredictionState, DistilState>;

export const actions = {
  // fetches variable summary data for the given dataset and variables
  async fetchTrainingSummaries(
    context: PredictionContext,
    args: {
      dataset: string;
      training: Variable[];
      solutionId: string;
      highlight: Highlight;
      varModes: Map<string, SummaryMode>;
      produceRequestId: string;
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
    if (!args.varModes) {
      console.warn("`varModes` argument is missing");
      return null;
    }
    if (!args.solutionId) {
      console.warn("`solutionId` argument is missing");
      return null;
    }
    const predictions = getPredictionsById(
      context.rootState.requestsModule.predictions,
      args.produceRequestId
    );
    if (!predictions) {
      // no results ready to pull
      return;
    }

    const dataset = args.dataset;
    const resultId = predictions.resultId;

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
          resultID: resultId,
          highlight: args.highlight,
          varMode: args.varModes.has(variable.colName)
            ? args.varModes.get(variable.colName)
            : SummaryMode.Default
        })
      );
    });
    return Promise.all(promises);
  },

  async fetchTrainingSummary(
    context: PredictionContext,
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
    if (!args.varMode) {
      console.warn("`varMode` argument is missing");
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

  async fetchIncludedPredictionTableData(
    context: PredictionContext,
    args: {
      solutionId: string;
      dataset: string;
      highlight: Highlight;
      produceRequestId: string;
    }
  ) {
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

    try {
      const response = await axios.post(
        `distil/prediction-results/${args.dataset}/${encodeURIComponent(
          args.solutionId
        )}/${encodeURIComponent(args.produceRequestId)}`,
        filterParams
      );
      mutations.setIncludedPredictionTableData(context, response.data);
    } catch (error) {
      console.error(
        `Failed to fetch results from ${args.solutionId} with error ${error}`
      );
      mutations.setIncludedPredictionTableData(context, createEmptyTableData());
    }
  },

  async fetchExcludedPredictionTableData(
    context: PredictionContext,
    args: {
      solutionId: string;
      dataset: string;
      highlight: Highlight;
      produceRequestId: string;
    }
  ) {
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
    filterParams = addHighlightToFilterParams(
      filterParams,
      args.highlight,
      EXCLUDE_FILTER
    );

    try {
      const response = await axios.post(
        `distil/prediction-results/${args.dataset}/${encodeURIComponent(
          args.solutionId
        )}/${encodeURIComponent(args.produceRequestId)}`,
        filterParams
      );
      mutations.setExcludedPredictionTableData(context, response.data);
    } catch (error) {
      console.error(
        `Failed to fetch results from ${args.solutionId} with error ${error}`
      );
      mutations.setExcludedPredictionTableData(context, createEmptyTableData());
    }
  },

  fetchPredictionTableData(
    context: PredictionContext,
    args: {
      solutionId: string;
      dataset: string;
      highlight: Highlight;
      produceRequestId: string;
    }
  ) {
    return Promise.all([
      actions.fetchIncludedPredictionTableData(context, {
        dataset: args.dataset,
        solutionId: args.solutionId,
        highlight: args.highlight,
        produceRequestId: args.produceRequestId
      }),
      actions.fetchExcludedPredictionTableData(context, {
        dataset: args.dataset,
        solutionId: args.solutionId,
        highlight: args.highlight,
        produceRequestId: args.produceRequestId
      })
    ]);
  },

  // fetches result summary for a given solution id.
  fetchPredictionSummary(
    context: PredictionContext,
    args: {
      dataset: string;
      target: string;
      solutionId: string;
      highlight: Highlight;
      varMode: SummaryMode;
      produceRequestId: string;
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

    const predictions = getPredictionsById(
      context.rootState.requestsModule.predictions,
      args.produceRequestId
    );
    if (!predictions.resultId) {
      return null;
    }

    let filterParams = {
      highlight: null,
      variables: [],
      filters: []
    };
    filterParams = addHighlightToFilterParams(filterParams, args.highlight);

    const endpoint = `/distil/predicted-summary/${args.dataset}/${args.target}`;
    const key = predictions.predictedKey;
    const label = "Predicted";
    return fetchPredictionResultSummary(
      context,
      endpoint,
      predictions,
      key,
      label,
      predictionGetters.getPredictionSummaries(context),
      mutations.updatePredictedSummary,
      filterParams,
      args.varMode
    );
  },

  async fetchForecastedTimeseries(
    context: PredictionContext,
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
    if (!solution.resultId) {
      // no results ready to pull
      return null;
    }

    try {
      const response = await axios.post(
        `distil/timeseries-forecast/${args.dataset}/${args.timeseriesColName}/${args.xColName}/${args.yColName}/${args.timeseriesID}/${solution.resultId}`,
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
        forecastTestRange: response.data.forecastRange,
        isDateTime: response.data.isDateTime
      });
    } catch (error) {
      console.error(error);
    }
  }
};
