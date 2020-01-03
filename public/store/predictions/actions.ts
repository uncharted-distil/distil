import axios from "axios";
import _ from "lodash";
import { ActionContext } from "vuex";
import store, { DistilState } from "../store";
import { EXCLUDE_FILTER } from "../../util/filters";
import {
  getSolutionsByRequestIds,
  getSolutionById
} from "../../util/solutions";
import { Variable, Highlight } from "../dataset/index";
import { mutations } from "./module";
import { PredictionState } from "./index";
import { addHighlightToFilterParams } from "../../util/highlights";
import {
  fetchPredictionResultSummary,
  createPendingSummary,
  createErrorSummary,
  createEmptyTableData,
  fetchSummaryExemplars,
  getTimeseriesAnalysisIntervals
} from "../../util/data";
import { getters as predictionGetters } from "../predictions/module";
import { getters as dataGetters } from "../dataset/module";

export type PredictionContext = ActionContext<PredictionState, DistilState>;

export const actions = {
  // fetches variable summary data for the given dataset and variables
  fetchTrainingSummaries(
    context: PredictionContext,
    args: {
      dataset: string;
      training: Variable[];
      solutionId: string;
      highlight: Highlight;
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
    const solution = getSolutionById(
      context.rootState.solutionModule,
      args.solutionId
    );
    if (!solution.resultId) {
      // no results ready to pull
      return;
    }

    const dataset = args.dataset;
    const solutionId = args.solutionId;

    const promises = [];

    // remove summaries not used to predict the newly selected model
    context.state.trainingSummaries.forEach(v => {
      const isTrainingArg = args.training.reduce((isTrain, variable) => {
        if (!isTrain) {
          isTrain = variable.colName === v.key;
        }
        return isTrain;
      }, false);
      if(v.dataset !== args.dataset || !isTrainingArg) {
        mutations.removeTrainingSummary(context, v);
      };
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
          createPendingSummary(key, label, description, dataset, solutionId)
        );
      }
      // fetch summary
      promises.push(
        actions.fetchTrainingSummary(context, {
          dataset: dataset,
          variable: variable,
          resultID: solution.resultId,
          highlight: args.highlight
        })
      );
    });
    return Promise.all(promises);
  },

  fetchTrainingSummary(
    context: PredictionContext,
    args: {
      dataset: string;
      variable: Variable;
      resultID: string;
      highlight: Highlight;
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

    const timeseries = context.getters.getRouteTimeseriesAnalysis;
    if (timeseries) {
      let interval = context.getters.getRouteTimeseriesBinningInterval;
      if (!interval) {
        const timeVar = context.getters.getTimeseriesAnalysisVariable;
        const range = context.getters.getTimeseriesAnalysisRange;
        const intervals = getTimeseriesAnalysisIntervals(timeVar, range);
        interval = intervals[0].value;
      }

      return axios
        .post(
          `distil/training-timeseries-summary/${args.dataset}/${timeseries}/${args.variable.colName}/${interval}/${args.resultID}`,
          filterParams
        )
        .then(response => {
          const summary = response.data.summary;
          mutations.updateTrainingSummary(context, summary);
        })
        .catch(error => {
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
        });
    }

    return axios
      .post(
        `/distil/training-summary/${args.dataset}/${args.variable.colName}/${args.resultID}`,
        filterParams
      )
      .then(response => {
        const summary = response.data.summary;
        return fetchSummaryExemplars(
          args.dataset,
          args.variable.colName,
          summary
        ).then(() => {
          mutations.updateTrainingSummary(context, summary);
        });
      })
      .catch(error => {
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
      });
  },

  fetchTargetSummary(
    context: PredictionContext,
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
    if (!args.target) {
      console.warn("`variable` argument is missing");
      return null;
    }
    if (!args.solutionId) {
      console.warn("`solutionId` argument is missing");
      return null;
    }
    const solution = getSolutionById(
      context.rootState.solutionModule,
      args.solutionId
    );
    if (!solution.resultId) {
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
        createPendingSummary(
          key,
          label,
          targetVar.colDescription,
          dataset,
          args.solutionId
        )
      );
    }

    let filterParams = {
      highlight: null,
      variables: [],
      filters: []
    };
    filterParams = addHighlightToFilterParams(filterParams, args.highlight);

    const timeseries = context.getters.getRouteTimeseriesAnalysis;
    if (timeseries) {
      let interval = context.getters.getRouteTimeseriesBinningInterval;
      if (!interval) {
        const timeVar = context.getters.getTimeseriesAnalysisVariable;
        const range = context.getters.getTimeseriesAnalysisRange;
        const intervals = getTimeseriesAnalysisIntervals(timeVar, range);
        interval = intervals[0].value;
      }

      return axios
        .post(
          `distil/target-timeseries-summary/${args.dataset}/${timeseries}/${args.target}/${interval}/${solution.resultId}`,
          filterParams
        )
        .then(response => {
          const summary = response.data.summary;
          mutations.updateTargetSummary(context, summary);
        })
        .catch(error => {
          console.error(error);
          mutations.updateTargetSummary(
            context,
            createErrorSummary(key, label, dataset, error)
          );
        });
    }

    return axios
      .post(
        `/distil/target-summary/${args.dataset}/${args.target}/${solution.resultId}`,
        filterParams
      )
      .then(response => {
        const summary = response.data.summary;
        return fetchSummaryExemplars(args.dataset, args.target, summary).then(
          () => {
            mutations.updateTargetSummary(context, summary);
          }
        );
      })
      .catch(error => {
        console.error(error);
        mutations.updateTargetSummary(
          context,
          createErrorSummary(key, label, dataset, error)
        );
      });
  },

  fetchIncludedPredictionTableData(
    context: PredictionContext,
    args: { solutionId: string; dataset: string; highlight: Highlight; produceRequestId: string  }
  ) {
    const solution = getSolutionById(
      context.rootState.solutionModule,
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

    return axios
      .post(
        `distil/prediction-results/${args.dataset}/${encodeURIComponent(args.solutionId)}/${encodeURIComponent(args.produceRequestId)}`,
        filterParams
      )
      .then(response => {
        mutations.setIncludedPredictionTableData(context, response.data);
      })
      .catch(error => {
        console.error(
          `Failed to fetch results from ${args.solutionId} with error ${error}`
        );
        mutations.setIncludedPredictionTableData(context, createEmptyTableData());
      });
  },

  fetchExcludedPredictionTableData(
    context: PredictionContext,
    args: { solutionId: string; dataset: string; highlight: Highlight; produceRequestId: string  }
  ) {
    const solution = getSolutionById(
      context.rootState.solutionModule,
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

    return axios
    .post(
      `distil/prediction-results/${args.dataset}/${encodeURIComponent(args.solutionId)}/${encodeURIComponent(args.produceRequestId)}`,
      filterParams
    )
      .then(response => {
        mutations.setExcludedPredictionTableData(context, response.data);
      })
      .catch(error => {
        console.error(
          `Failed to fetch results from ${args.solutionId} with error ${error}`
        );
        mutations.setExcludedPredictionTableData(context, createEmptyTableData());
      });
  },

  fetchPredictionTableData(
    context: PredictionContext,
    args: { solutionId: string; dataset: string; highlight: Highlight; produceRequestId: string }
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

  const solution = getSolutionById(
    context.rootState.solutionModule,
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

  const timeseries = context.getters.getRouteTimeseriesAnalysis;
  if (timeseries) {
    let interval = context.getters.getRouteTimeseriesBinningInterval;
    if (!interval) {
      const timeVar = context.getters.getTimeseriesAnalysisVariable;
      const range = context.getters.getTimeseriesAnalysisRange;
      const intervals = getTimeseriesAnalysisIntervals(timeVar, range);
      interval = intervals[0].value;
    }

    const endPoint = `distil/forecasting-summary/${args.dataset}/${timeseries}/${args.target}/${interval}`;
    const key = solution.predictedKey;
    const label = "Forecasted";
    return fetchPredictionResultSummary(
      context,
      endPoint,
      solution,
      args.target,
      key,
      label,
      predictionGetters.getPredictionSummaries(context),
      mutations.updatePredictedSummaries,
      filterParams
    );
  }

  const endpoint = `/distil/predicted-summary/${args.dataset}/${args.target}`;
  const key = solution.predictedKey;
  const label = "Predicted";
  return fetchPredictionResultSummary(
    context,
    endpoint,
    solution,
    args.target,
    key,
    label,
    predictionGetters.getPredictionSummaries(context),
    mutations.updatePredictedSummaries,
    filterParams
  );
},

  // fetches result summaries for a given solution create request
  fetchPredictionSummaries(
    context: PredictionContext,
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
    const solutions = getSolutionsByRequestIds(
      context.rootState.solutionModule,
      args.requestIds
    );
    return Promise.all(
      solutions.map(solution => {
        return actions.fetchPredictionSummary(context, {
          dataset: args.dataset,
          target: args.target,
          solutionId: solution.solutionId,
          highlight: args.highlight
        });
      })
    );
  },

  fetchForecastedTimeseries(
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
      context.rootState.solutionModule,
      args.solutionId
    );
    if (!solution.resultId) {
      // no results ready to pull
      return null;
    }

    return axios
      .post(
        `distil/timeseries-forecast/${args.dataset}/${args.timeseriesColName}/${args.xColName}/${args.yColName}/${args.timeseriesID}/${solution.resultId}`,
        {}
      )
      .then(response => {
        mutations.updatePredictedTimeseries(context, {
          solutionId: args.solutionId,
          id: args.timeseriesID,
          timeseries: response.data.timeseries
        });
        mutations.updatePredictedForecast(context, {
          solutionId: args.solutionId,
          id: args.timeseriesID,
          forecast: response.data.forecast
        });
      })
      .catch(error => {
        console.error(error);
      });
  }
};