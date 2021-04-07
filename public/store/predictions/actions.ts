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

import axios from "axios";
import _ from "lodash";
import { ActionContext } from "vuex";
import { DistilState } from "../store";
import { FilterParams, EXCLUDE_FILTER, Filter } from "../../util/filters";
import {
  Variable,
  Highlight,
  SummaryMode,
  VariableSummary,
} from "../dataset/index";
import { mutations } from "./module";
import { PredictionState } from "./index";
import { addHighlightToFilterParams } from "../../util/highlights";
import {
  fetchPredictionResultSummary,
  createErrorSummary,
  createEmptyTableData,
  fetchSummaryExemplars,
  minimumRouteKey,
  createPendingSummary,
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
      highlights: Highlight[];
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

    const summariesByVariable = context.state.trainingSummaries;
    const routeKey = minimumRouteKey();

    args.training.forEach((variable) => {
      const key = variable.key;
      const label = variable.colDisplayName;
      const description = variable.colDescription;
      const existingVariableSummary =
        summariesByVariable?.[variable.key]?.[routeKey];

      if (existingVariableSummary) {
        promises.push(existingVariableSummary);
      } else {
        if (summariesByVariable[variable.key]) {
          // if we have any saved state for that variable
          // use that as placeholder due to vue lifecycle
          const tempVariableSummaryKey = Object.keys(
            summariesByVariable[variable.key]
          )[0];
          promises.push(
            summariesByVariable[variable.key][tempVariableSummaryKey]
          );
        } else {
          // add a loading placeholder if nothing exists for that variable
          createPendingSummary(
            variable.key,
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
            resultID: resultId,
            highlights: args.highlights,
            varMode: args.varModes.has(variable.key)
              ? args.varModes.get(variable.key)
              : SummaryMode.Default,
          })
        );
      }
    });
    return Promise.all(promises);
  },
  // fetches
  async fetchAreaOfInterestInner(
    context: PredictionContext,
    args: {
      produceRequestId: string;
      dataset: string;
      highlights: Highlight[];
      size?: number;
      filter: Filter; // the area of interest
    }
  ) {
    const filterParamsBlank = {
      highlights: { list: [], invert: false },
      variables: [],
      filters: { list: [], invert: false },
    };
    const filterParams = addHighlightToFilterParams(
      filterParamsBlank,
      args.highlights
    );

    if (_.isInteger(args.size)) {
      filterParams.size = args.size;
    }
    filterParams.filters.list.push(args.filter);
    try {
      const response = await axios.post(
        `/distil/prediction-results/${encodeURIComponent(
          args.produceRequestId
        )}`,
        filterParams
      );
      mutations.setAreaOfInterestInner(context, response.data);
    } catch (error) {
      console.error(
        `Failed to fetch results from ${args.produceRequestId} with error ${error}`
      );
      mutations.setAreaOfInterestInner(context, createEmptyTableData());
    }
  },
  // fetches the tiles that are within the bounds but are filtered by another highlight
  async fetchAreaOfInterestOuter(
    context: PredictionContext,
    args: {
      produceRequestId: string;
      dataset: string;
      highlights: Highlight[];
      size?: number;
      filter: Filter;
    }
  ) {
    const filterParamsBlank = {
      highlights: { list: [], invert: false },
      variables: [],
      filters: { list: [], invert: false },
    };
    const filterParams = addHighlightToFilterParams(
      filterParamsBlank,
      args.highlights,
      EXCLUDE_FILTER
    );
    // Add the size limit to results if provided.
    if (_.isInteger(args.size)) {
      filterParams.size = args.size;
    }
    filterParams.filters.list.push(args.filter);
    // if highlight is null there is nothing to invert so return null
    if (
      filterParams.highlights === null &&
      filterParams.highlights.list.length > 0
    ) {
      mutations.setAreaOfInterestOuter(context, createEmptyTableData());
      return;
    }
    try {
      const response = await axios.post(
        `/distil/prediction-results/${args.dataset}/${encodeURIComponent(
          args.produceRequestId
        )}`,
        filterParams
      );
      mutations.setAreaOfInterestOuter(context, response.data);
    } catch (error) {
      console.error(
        `Failed to fetch results from ${args.produceRequestId} with error ${error}`
      );
      mutations.setAreaOfInterestOuter(context, createEmptyTableData());
    }
  },
  async fetchTrainingSummary(
    context: PredictionContext,
    args: {
      dataset: string;
      variable: Variable;
      resultID: string;
      highlights: Highlight[];
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
      highlights: { list: [], invert: false },
      variables: [],
      filters: { list: [], invert: false },
    } as FilterParams;
    filterParams = addHighlightToFilterParams(filterParams, args.highlights);
    try {
      const response = await axios.post(
        `/distil/training-summary/${args.dataset}/${args.variable.key}/${args.resultID}/${args.varMode}`,
        filterParams
      );
      const summary = response.data.summary;
      await fetchSummaryExemplars(args.dataset, args.variable.key, summary);
      mutations.updateTrainingSummary(context, summary);
    } catch (error) {
      console.error(error);
      mutations.updateTrainingSummary(
        context,
        createErrorSummary(
          args.variable.key,
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
      highlights: Highlight[];
      produceRequestId: string;
      isBaseline: boolean;
      size?: number;
    }
  ) {
    let filterParams = {
      highlights: { list: [], invert: false },
      variables: [],
      filters: { list: [], invert: false },
    } as FilterParams;
    filterParams = addHighlightToFilterParams(filterParams, args.highlights);
    const mutator = args.isBaseline
      ? mutations.setBaselinePredictionTableData
      : mutations.setIncludedPredictionTableData;
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
      mutator(context, response.data);
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
      highlights: Highlight[];
      produceRequestId: string;
      size?: number;
      isBaseline: boolean;
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
      highlights: Highlight[];
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
      highlights: { list: [], invert: false },
      variables: [],
      filters: { list: [], invert: false },
    } as FilterParams;
    filterParams = addHighlightToFilterParams(filterParams, args.highlights);

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
      highlights: Highlight[];
      fittedSolutionId: string;
    }
  ) {
    const predictions = context.rootState.requestsModule.predictions.filter(
      (p) => p.fittedSolutionId === args.fittedSolutionId
    );
    return Promise.all(
      predictions.map((p) =>
        actions.fetchPredictionSummary(context, {
          highlights: args.highlights,
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
      predictionsId: string;
      timeseriesIds: string[];
      uniqueTrail?: string;
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
    if (!args.timeseriesIds) {
      console.warn("`timeseriesIds` argument is missing");
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
          `/${args.timeseriesColName}/${args.xColName}/${args.yColName}` +
          `/${predictions.resultId}`,
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
        predictionsId: args.predictionsId,
        map: responseMap,
      });
      mutations.bulkUpdatePredictedForecast(context, {
        predictionsId: args.predictionsId,
        map: responseMap,
      });
    } catch (error) {
      console.error(error);
    }
  },

  async fetchExportData(
    context: PredictionContext,
    args: {
      produceRequestId: string;
      format: string;
    }
  ): Promise<string> {
    try {
      const endPoint = `/distil/export-results/${args.produceRequestId}/${args.format}`;
      const response = await axios.get(endPoint);
      return response.data;
    } catch (error) {
      console.error(error);
    }
    return null;
  },

  async createDataset(
    context: PredictionContext,
    args: {
      produceRequestId: string;
      newDatasetName: string;
      includeDatasetFeatures?: boolean;
    }
  ): Promise<Error> {
    try {
      const endPoint = "/distil/clone-result/";
      const params = `${args.produceRequestId}`;
      const response = await axios.post(
        `/distil/clone-result/${encodeURIComponent(args.produceRequestId)}`,
        {
          datasetName: args.newDatasetName,
          includeDatasetFeatures: args.includeDatasetFeatures,
        }
      );
    } catch (error) {
      console.error(error);
      return error;
    }
    return null;
  },
};
