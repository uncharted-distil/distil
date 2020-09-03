import axios from "axios";
import _ from "lodash";
import { ActionContext } from "vuex";
import { DistilState } from "../store";
import { FilterParams } from "../../util/filters";
import { Variable, Highlight, SummaryMode } from "../dataset/index";
import { mutations } from "./module";
import { PredictionState } from "./index";
import { addHighlightToFilterParams } from "../../util/highlights";
import {
  fetchPredictionResultSummary,
  createErrorSummary,
  createEmptyTableData,
  fetchSummaryExemplars,
} from "../../util/data";
import {
  getters as predictionGetters,
  mutations as predictionMutations,
} from "../predictions/module";
import { getPredictionsById } from "../../util/predictions";

export type PredictionContext = ActionContext<PredictionState, DistilState>;

export const actions = {
  // fetches variable summary data for the given dataset and variables
  async fetchTrainingSummaries(
    context: PredictionContext,
    args: {
      dataset: string;
      training: Variable[];
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

    context.state.trainingSummaries
      .filter(
        (summary) =>
          !args.training.find(
            (variable) =>
              variable.colName === summary.key &&
              args.dataset === summary.dataset
          )
      )
      .forEach((summary) =>
        predictionMutations.removeTrainingSummary(context, summary)
      );

    args.training.forEach((variable) => {
      const key = variable.colName;
      const label = variable.colDisplayName;
      const description = variable.colDescription;

      // fetch summary
      promises.push(
        actions.fetchTrainingSummary(context, {
          dataset: dataset,
          variable: variable,
          resultID: resultId,
          highlight: args.highlight,
          varMode: args.varModes.has(variable.colName)
            ? args.varModes.get(variable.colName)
            : SummaryMode.Default,
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
      filters: [],
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

  // TODO: shouldn't need solutionID as an arg
  async fetchIncludedPredictionTableData(
    context: PredictionContext,
    args: {
      dataset: string;
      highlight: Highlight;
      produceRequestId: string;
      size?: number;
    }
  ) {
    let filterParams: FilterParams = {
      highlight: null,
      variables: [],
      filters: [],
    };
    filterParams = addHighlightToFilterParams(filterParams, args.highlight);

    // Add the size limit to results if provided.
    if (_.isInteger(args.size)) {
      filterParams.size = args.size;
    }

    try {
      const response = await axios.post(
        `distil/prediction-results/${encodeURIComponent(
          args.produceRequestId
        )}`,
        filterParams
      );
      mutations.setIncludedPredictionTableData(context, response.data);
    } catch (error) {
      console.error(
        `Failed to fetch results from ${args.produceRequestId} with error ${error}`
      );
      mutations.setIncludedPredictionTableData(context, createEmptyTableData());
    }
  },

  fetchPredictionTableData(
    context: PredictionContext,
    args: {
      dataset: string;
      highlight: Highlight;
      produceRequestId: string;
      size?: number;
    }
  ) {
    return Promise.all([
      actions.fetchIncludedPredictionTableData(context, args),
    ]);
  },

  // fetches result summary for prediction request id.
  fetchPredictionSummary(
    context: PredictionContext,
    args: {
      highlight: Highlight;
      varMode: SummaryMode;
      produceRequestId: string;
    }
  ) {
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
      filters: [],
    };
    filterParams = addHighlightToFilterParams(filterParams, args.highlight);

    const endpoint = `/distil/prediction-result-summary`;
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

  // fetches all result summaries for a fitted solution id.
  fetchPredictionSummaries(
    context: PredictionContext,
    args: {
      highlight: Highlight;
      fittedSolutionId: string;
    }
  ) {
    const predictions = context.rootState.requestsModule.predictions.filter(
      (p) => p.fittedSolutionId === args.fittedSolutionId
    );
    return Promise.all(
      predictions.map((p) =>
        actions.fetchPredictionSummary(context, {
          highlight: args.highlight,
          varMode: SummaryMode.Default,
          produceRequestId: p.requestId,
        })
      )
    );
  },

  async fetchForecastedTimeseries(
    context: PredictionContext,
    args: {
      truthDataset: string;
      forecastDataset: string;
      xColName: string;
      yColName: string;
      timeseriesColName: string;
      timeseriesId: any;
      predictionsId: string;
    }
  ) {
    if (!args.truthDataset) {
      console.warn("`truthDataset` argument is missing");
      return null;
    }
    if (!args.forecastDataset) {
      console.warn("`forecastDataset` argument is missing");
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
    if (!args.predictionsId) {
      console.warn("`solutionId` argument is missing");
      return null;
    }

    const predictions = getPredictionsById(
      context.rootState.requestsModule.predictions,
      args.predictionsId
    );
    if (!predictions.resultId) {
      // no results ready to pull
      return null;
    }

    try {
      const response = await axios.post(
        `distil/timeseries-forecast/${args.truthDataset}/${args.forecastDataset}` +
          `/${args.timeseriesColName}/${args.xColName}/${args.yColName}/${args.timeseriesId}` +
          `/${predictions.resultId}`,
        {}
      );
      mutations.updatePredictedTimeseries(context, {
        predictionsId: args.predictionsId,
        id: args.timeseriesId,
        timeseries: response.data.timeseries,
        isDateTime: response.data.isDateTime,
        min: <number>response.data.min,
        max: <number>response.data.max,
        mean: <number>response.data.mean,
      });
      mutations.updatePredictedForecast(context, {
        predictionsId: args.predictionsId,
        id: args.timeseriesId,
        forecast: response.data.forecast,
        forecastTestRange: response.data.forecastRange,
        isDateTime: response.data.isDateTime,
      });
    } catch (error) {
      console.error(error);
    }
  },
};
