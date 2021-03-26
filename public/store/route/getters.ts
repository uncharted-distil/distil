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
import { $enum } from "ts-enum-util";
import { Route } from "vue-router";
import { ColorScaleNames, minimumRouteKey } from "../../util/data";
import { Dictionary } from "../../util/dict";
import { decodeFilters, Filter, FilterParams } from "../../util/filters";
import { decodeHighlights } from "../../util/highlights";
import { buildLookup } from "../../util/lookup";
import { decodeRowSelection } from "../../util/row";
import { GEOBOUNDS_TYPE, GEOCOORDINATE_TYPE } from "../../util/types";
import {
  BandID,
  DataMode,
  Highlight,
  RowSelection,
  SummaryMode,
  TaskTypes,
  Variable,
  VariableSummary,
} from "../dataset/index";
import { ModelQuality } from "../requests/index";
import {
  AVAILABLE_TARGET_VARS_INSTANCE_PAGE,
  AVAILABLE_TARGET_VARS_INSTANCE_SEARCH,
  AVAILABLE_TRAINING_VARS_INSTANCE_PAGE,
  AVAILABLE_TRAINING_VARS_INSTANCE_SEARCH,
  DATA_EXPLORER_ROUTE,
  DATA_EXPLORER_VARS_INSTANCE_PAGE,
  DATA_EXPLORER_VARS_INSTANCE_SEARCH,
  DATA_SIZE_DEFAULT,
  DATA_SIZE_REMOTE_SENSING_DEFAULT,
  JOINED_VARS_INSTANCE_PAGE,
  JOINED_VARS_INSTANCE_SEARCH,
  JOIN_DATASETS_ROUTE,
  LABEL_FEATURE_VARS_INSTANCE_PAGE,
  RESULTS_ROUTE,
  RESULT_TRAINING_VARS_INSTANCE_PAGE,
  RESULT_TRAINING_VARS_INSTANCE_SEARCH,
  SELECT_TARGET_ROUTE,
  SELECT_TRAINING_ROUTE,
  TRAINING_VARS_INSTANCE_PAGE,
  TRAINING_VARS_INSTANCE_SEARCH,
} from "../route/index";

export const getters = {
  getRoute(state: Route): Route {
    return state;
  },

  getRoutePath(state: Route): string {
    return state.path;
  },

  getRouteTerms(state: Route): string {
    return state.query.terms as string;
  },

  getRouteDataset(state: Route): string {
    return state.query.dataset as string;
  },

  getRouteInclude(state: Route): boolean {
    if (state.query.include === "false") {
      return false;
    }
    return true;
  },

  getPriorPath(state: Route): string {
    return state.query.priorRoute as string;
  },

  getRouteJoinDatasets(state: Route): string[] {
    return state.query.joinDatasets
      ? (state.query.joinDatasets as string).split(",")
      : [];
  },

  getRouteJoinDatasetsHash(state: Route): string {
    return state.query.joinDatasets as string;
  },

  getJoinDatasetsVariables(state: Route, getters: any): Variable[] {
    const datasetIDs = getters.getRouteJoinDatasets;
    if (datasetIDs.length !== 2) {
      return [];
    }
    const datasets = getters.getDatasets;
    const datasetA = _.find(datasets, (d) => {
      return d.id === datasetIDs[0];
    });
    const datasetB = _.find(datasets, (d) => {
      return d.id === datasetIDs[1];
    });
    let variables = [];
    if (datasetA) {
      datasetA.variables.forEach((v) => {
        v.datasetName = datasetIDs[0];
      });
      variables = variables.concat(datasetA.variables);
    }
    if (datasetB) {
      datasetB.variables.forEach((v) => {
        v.datasetName = datasetIDs[1];
      });
      variables = variables.concat(datasetB.variables);
    }
    return variables;
  },
  getAnnotationHasChanged(state: Route) {
    const hasChanged =
      state.query.annotationHasChanged === "true" ||
      !state.query.annotationHasChanged;
    return hasChanged;
  },

  getJoinDatasetColumnA(state: Route, getters: any): string {
    return state.query.joinColumnA as string;
  },

  getJoinDatasetColumnB(state: Route, getters: any): string {
    return state.query.joinColumnB as string;
  },

  getBaseColumnSuggestions(state: Route, getters: any): string[] {
    return state.query.baseColumnSuggestions as string[];
  },

  getJoinColumnSuggestions(state: Route, getters: any): string[] {
    return state.query.joinColumnSuggestions as string[];
  },

  getJoinAccuracy(state: Route, getters: any): number {
    const accuracy = state.query.joinAccuracy;
    return accuracy ? _.toNumber(accuracy) : 1;
  },
  getDecodedJoinDatasetsHighlight(
    state: Route,
    getters: any
  ): Dictionary<Highlight[]> {
    const datasetIDs = getters.getRouteJoinDatasets as string[];
    if (datasetIDs.length !== 2) {
      return {};
    }
    const highlights = getters.getDecodedHighlights as Highlight[];
    const result = {};
    datasetIDs.forEach((id) => {
      result[id] = [];
    });
    highlights.forEach((highlight) => {
      if (result[highlight.dataset]) {
        result[highlight.dataset].push(highlight);
      }
    });
    return result;
  },
  getDecodedJoinDatasetsFilterParams(
    state: Route,
    getters: any
  ): Dictionary<FilterParams> {
    const datasetIDs = getters.getRouteJoinDatasets;
    if (datasetIDs.length !== 2) {
      return {};
    }
    const datasets = getters.getDatasets;
    const res = {};

    // build filter params for each dataset
    datasetIDs.forEach((datasetID) => {
      const dataset = _.find(datasets, (d) => {
        return d.id === datasetID;
      });
      if (dataset) {
        const filters = getters.getDecodedFilters;

        // only include filters for this dataset
        const lookup = buildLookup(dataset.variables.map((v) => v.key));
        const filtersForDataset = filters.filter((f) => {
          return lookup[f.key];
        });

        const filterParams = _.cloneDeep({
          filters: { list: filtersForDataset },
          variables: dataset.variables.map((v) => v.key),
          highlights: { list: [] },
        });
        res[datasetID] = filterParams;
      }
    });
    return res;
  },

  getRouteTrainingVariables(state: Route): string {
    return state.query.training ? (state.query.training as string) : null;
  },

  // Return the list of variable displayed in the Data Explorer view
  getExploreVariables(state: Route): string[] {
    const explore = state.query?.explore as string;
    return explore?.split(",") ?? [];
  },

  // Returns a boolean to say that the variables for this dataset has been ranked.
  getRouteIsTrainingVariablesRanked(state: Route): boolean {
    return state.query.varRanked && state.query.varRanked === "1"; // Use "1" for truth.
  },

  // Returns a boolean to say that the cluster for this dataset has been generated.
  getRouteIsClusterGenerated(state: Route): boolean {
    return state.query.clustering && state.query.clustering === "1"; // Use "1" for truth.
  },

  // Returns a boolean to say that the outlier detection for this dataset has been applied.
  isOutlierApplied(state: Route): boolean {
    return state.query.outlier && state.query.outlier === "1"; // Use "1" for truth.
  },

  getRouteJoinDatasetsVarsPage(state: Route): number {
    const pageVar = JOINED_VARS_INSTANCE_PAGE;
    return state.query[pageVar] ? _.toNumber(state.query[pageVar]) : 1;
  },

  getRouteAvailableTargetVarsPage(state: Route): number {
    const pageVar = AVAILABLE_TARGET_VARS_INSTANCE_PAGE;
    return state.query[pageVar] ? _.toNumber(state.query[pageVar]) : 1;
  },

  getRouteAvailableTrainingVarsPage(state: Route): number {
    const pageVar = AVAILABLE_TRAINING_VARS_INSTANCE_PAGE;
    return state.query[pageVar] ? _.toNumber(state.query[pageVar]) : 1;
  },
  getLabelFeaturesVarsPage(state: Route): number {
    const pageVar = LABEL_FEATURE_VARS_INSTANCE_PAGE;
    return state.query[pageVar] ? _.toNumber(state.query[pageVar]) : 1;
  },
  getRouteTrainingVarsPage(state: Route): number {
    const pageVar = TRAINING_VARS_INSTANCE_PAGE;
    return state.query[pageVar] ? _.toNumber(state.query[pageVar]) : 1;
  },

  getRouteResultTrainingVarsPage(state: Route): number {
    const pageVar = RESULT_TRAINING_VARS_INSTANCE_PAGE;
    return state.query[pageVar] ? _.toNumber(state.query[pageVar]) : 1;
  },

  /* The Data Explorer uses a different page variables per action */
  getRouteDataExplorerVarsPage(state: Route): any {
    const pageVar = DATA_EXPLORER_VARS_INSTANCE_PAGE;
    return state.query[pageVar] ? _.toNumber(state.query[pageVar]) : 1;
  },

  getAllRoutePages(state: Route, getters: any): Object {
    const pages = {};
    pages[JOIN_DATASETS_ROUTE] = [getters.getRouteJoinDatasetsVarsPage];
    pages[SELECT_TARGET_ROUTE] = [getters.getRouteAvailableTargetVarsPage];
    pages[DATA_EXPLORER_ROUTE] = [getters.getRouteDataExplorerVarsPage];
    pages[SELECT_TRAINING_ROUTE] = [
      getters.getRouteAvailableTrainingVarsPage,
      getters.getRouteTrainingVarsPage,
    ];
    pages[RESULTS_ROUTE] = [getters.getRouteResultTrainingVarsPage];
    return pages;
  },

  getRouteJoinDatasetsVarsSearch(state: Route): string {
    const searchVar = JOINED_VARS_INSTANCE_SEARCH;
    return state.query[searchVar] ? _.toString(state.query[searchVar]) : "";
  },

  getRouteAvailableTargetVarsSearch(state: Route): string {
    const searchVar = AVAILABLE_TARGET_VARS_INSTANCE_SEARCH;
    return state.query[searchVar] ? _.toString(state.query[searchVar]) : "";
  },

  getRouteAvailableTrainingVarsSearch(state: Route): string {
    const searchVar = AVAILABLE_TRAINING_VARS_INSTANCE_SEARCH;
    return state.query[searchVar] ? _.toString(state.query[searchVar]) : "";
  },

  getRouteTrainingVarsSearch(state: Route): string {
    const searchVar = TRAINING_VARS_INSTANCE_SEARCH;
    return state.query[searchVar] ? _.toString(state.query[searchVar]) : "";
  },

  getRouteResultTrainingVarsSearch(state: Route): string {
    const searchVar = RESULT_TRAINING_VARS_INSTANCE_SEARCH;
    return state.query[searchVar] ? _.toString(state.query[searchVar]) : "";
  },

  getRouteDataExplorerVarsSearch(state: Route): string {
    const searchVar = DATA_EXPLORER_VARS_INSTANCE_SEARCH;
    return state.query[searchVar] ? _.toString(state.query[searchVar]) : "";
  },

  getAllSearchesByRoute(state: Route, getters: any): Object {
    const searches = {};
    searches[JOIN_DATASETS_ROUTE] = [getters.getRouteJoinDatasetsVarsSearch];
    searches[SELECT_TARGET_ROUTE] = [getters.getRouteAvailableTargetVarsSearch];
    searches[DATA_EXPLORER_ROUTE] = [getters.getRouteDataExplorerVarsSearch];
    searches[SELECT_TRAINING_ROUTE] = [
      getters.getRouteAvailableTrainingVarsSearch,
      getters.getRouteTrainingVarsSearch,
    ];
    searches[RESULTS_ROUTE] = [getters.getRouteResultTrainingVarsSearch];
    return searches;
  },

  getAllSearchesByQueryString(state: Route, getters: any): Object {
    const searches = {};
    searches[JOINED_VARS_INSTANCE_SEARCH] =
      getters.getRouteJoinDatasetsVarsSearch;
    searches[AVAILABLE_TARGET_VARS_INSTANCE_SEARCH] =
      getters.getRouteAvailableTargetVarsSearch;
    searches[AVAILABLE_TRAINING_VARS_INSTANCE_SEARCH] =
      getters.getRouteAvailableTrainingVarsSearch;
    searches[TRAINING_VARS_INSTANCE_SEARCH] =
      getters.getRouteTrainingVarsSearch;
    searches[RESULT_TRAINING_VARS_INSTANCE_SEARCH] =
      getters.getRouteResultTrainingVarsSearch;
    searches[DATA_EXPLORER_VARS_INSTANCE_SEARCH] =
      getters.getRouteDataExplorerVarsSearch;
    return searches;
  },

  getRouteTargetVariable(state: Route): string {
    return state.query.target ? (state.query.target as string) : null;
  },

  getRouteSolutionId(state: Route): string {
    return state.query.solutionId ? (state.query.solutionId as string) : null;
  },

  getRouteResultId(state: Route): string {
    return state.query.resultId ? (state.query.resultId as string) : null;
  },

  getRouteFilters(state: Route): string {
    return state.query.filters ? (state.query.filters as string) : null;
  },

  getRouteHighlight(state: Route): string {
    return state.query.highlights ? (state.query.highlights as string) : null;
  },

  getRouteRowSelection(state: Route): string {
    return state.query.row ? (state.query.row as string) : null;
  },

  getRouteResultFilters(state: Route): string {
    return state.query.results ? (state.query.results as string) : null;
  },

  getRouteDataSize(state: Route, getters: any): number {
    const dataSize = state.query.dataSize;
    if (dataSize) {
      return _.toInteger(dataSize);
    }

    const isMultiBandImage = getters.isMultiBandImage;
    return isMultiBandImage
      ? DATA_SIZE_REMOTE_SENSING_DEFAULT
      : DATA_SIZE_DEFAULT;
  },

  getRouteProduceRequestId(state: Route): string {
    return state.query.produceRequestId
      ? (state.query.produceRequestId as string)
      : null;
  },

  getRouteResidualThresholdMin(state: Route): string {
    return state.query.residualThresholdMin as string;
  },

  getRouteResidualThresholdMax(state: Route): string {
    return state.query.residualThresholdMax as string;
  },

  getDecodedTrainingVariableNames(state: Route, getters: any): string[] {
    const training = getters.getRouteTrainingVariables;
    return training ? training.split(",") : [];
  },

  getDecodedFilters(state: Route, getters: any): Filter[] {
    return decodeFilters(state.query.filters as string).list;
  },

  getDecodedSolutionRequestFilterParams(
    state: Route,
    getters: any
  ): FilterParams {
    const filters = getters.getDecodedFilters;
    const size = getters.getRouteDataSize;
    const filterParams = _.cloneDeep({
      highlights: { list: [] },
      variables: [],
      filters: { list: filters },
      size,
    });

    // If we have explore variables, we do not show the target & training ones
    const explore = getters.getExploreVariables;
    if (!_.isEmpty(explore)) {
      filterParams.variables = explore;
    }
    // Otherwise, we list the target & training non-null variables
    else {
      const training = getters.getDecodedTrainingVariableNames;
      const target = getters.getRouteTargetVariable as string;
      filterParams.variables = [...training, target].filter((v) => v);
    }

    return filterParams;
  },

  getDecodedHighlights(state: Route): Highlight[] {
    return decodeHighlights(state.query.highlights as string);
  },

  getDecodedRowSelection(state: Route): RowSelection {
    return decodeRowSelection(state.query.row as string);
  },

  getTrainingVariables(state: Route, getters: any): Variable[] {
    const training = getters.getDecodedTrainingVariableNames;
    const lookup = buildLookup(training);
    const variables = getters.getVariables;
    return variables.filter((variable) => lookup[variable.key.toLowerCase()]);
  },

  getTrainingVariableSummaries(state: Route, getters: any): VariableSummary[] {
    const training = getters.getDecodedTrainingVariableNames;
    const include = getters.getRouteInclude;
    const minKey = minimumRouteKey();
    const summaries = include
      ? getters.getIncludedVariableSummariesDictionary
      : getters.getExcludedVariableSummariesDictionary;
    const trainingVariableSummaries = training.reduce((acc, variableName) => {
      const variableSummary = summaries?.[variableName]?.[minKey];
      if (variableSummary) {
        acc.push(variableSummary);
      }
      return acc;
    }, []);
    return trainingVariableSummaries;
  },

  getTargetVariable(state: Route, getters: any): Variable {
    const target = getters.getRouteTargetVariable;
    if (target) {
      const variables = getters.getVariables;
      const found = variables.filter(
        (summary) => target.toLowerCase() === summary.key.toLowerCase()
      );
      if (found) {
        return found[0];
      }
    }
    return null;
  },

  getTargetVariableSummaries(state: Route, getters: any): VariableSummary[] {
    const target = getters.getRouteTargetVariable;
    const include = getters.getRouteInclude;
    const minKey = minimumRouteKey();
    const summaries = include
      ? getters.getIncludedVariableSummariesDictionary
      : getters.getExcludedVariableSummariesDictionary;
    const targetVariableSummary = summaries?.[target]?.[minKey];
    if (targetVariableSummary) {
      return [targetVariableSummary];
    } else {
      const currentVariable = summaries?.[target];
      if (currentVariable) {
        const placeholderKey = Object.keys(currentVariable)[0];
        return [currentVariable[placeholderKey]];
      } else {
        return [];
      }
    }
  },

  getAvailableVariables(state: Route, getters: any): Variable[] {
    const training = getters.getDecodedTrainingVariableNames;
    const target = getters.getRouteTargetVariable;
    const variables = getters.getVariables;
    const lookup =
      training && target ? buildLookup(training.concat([target])) : null;
    if (!lookup) return variables ?? ([] as Variable[]);
    return variables.filter((variable) => !lookup[variable.key.toLowerCase()]);
  },

  getActiveSolutionIndex(state: Route, getters: any): number {
    const solutionId = getters.getRouteSolutionId;
    const solutions = getters.getSolutions;
    return _.findIndex(solutions, (solution: any) => {
      return solution.solutionId === solutionId;
    });
  },
  getColorScale(state: Route, getters: any): ColorScaleNames {
    const colorScale = state.query.colorScale as ColorScaleNames;
    return colorScale ?? ColorScaleNames.viridis; // default to viridis
  },
  getGeoCenter(state: Route, getters: any): number[] {
    const geo = state.query.geo as string;
    if (!geo) {
      return null;
    }
    const split = geo.split(",");
    return [_.toNumber(split[0]), _.toNumber(split[1])];
  },

  getGeoZoom(state: Route, getters: any): number {
    const geo = state.query.geo as string;
    if (!geo) {
      return null;
    }
    const split = geo.split(",");
    return _.toNumber(split[2]);
  },

  getGroupingType(state: Route): string {
    return state.query.groupingType as string;
  },

  getRouteTask(state: Route, getters: any): string {
    const task = state.query.task as string;
    if (!task) {
      return null;
    }
    return task;
  },

  getDataMode(state: Route, getters: any): DataMode {
    const mode = state.query.dataMode as string;
    if (!mode) {
      return null;
    }
    return $enum(DataMode).asValueOrDefault(mode, DataMode.Default);
  },

  getImageAttention(state: Route, getters: any): boolean {
    const imageAttention = state.query.imageAttention === "true";
    return imageAttention;
  },
  // Returns a map of (variable ID, summary mode) tuples that indicated the mode args that should be
  // applied to a given variable when fetched from the server.
  getDecodedVarModes(state: Route, getters: any): Map<string, SummaryMode> {
    const varModes = state.query.varModes as string;
    if (!varModes) {
      return new Map<string, SummaryMode>();
    }
    const modeTuples = varModes.split(",");
    const modeMap: Map<string, SummaryMode> = new Map();
    modeTuples.forEach((m) => {
      const [k, v] = m.split(":");
      modeMap.set(
        k,
        $enum(SummaryMode).asValueOrDefault(v, SummaryMode.Default)
      );
    });
    return modeMap;
  },

  getRouteFittedSolutionId(state: Route, getters: any): string {
    const id = <string>state.query.fittedSolutionId;
    if (!id) {
      return null;
    }
    return id;
  },

  getRoutePredictionsDataset(state: Route, getters: any): string {
    const dataset = <string>state.query.predictionsDataset;
    if (!dataset) {
      return null;
    }
    return dataset;
  },

  isSingleSolution(state: Route, getters: any): boolean {
    const isSingleSolution = <string>state.query.singleSolution;
    return !!isSingleSolution;
  },

  /* Check if the model should be open using the 'Apply model' navigation. */
  isApplyModel(state: Route, getters: any): boolean {
    const isApplyModel = <string>state.query.applyModel;
    return !!isApplyModel;
  },
  getOrderBy(state: Route): string {
    const orderBy = state.query.orderBy as string;
    if (!orderBy) {
      return null;
    }
    return orderBy;
  },
  hasOrderBy(state: Route): boolean {
    const orderBy = state.query.orderBy as string;
    return !!orderBy;
  },

  /* Check if the current task includes Remote Sensing. */
  isMultiBandImage(state: Route): boolean {
    // Get the list of task of the route.
    const task = state.query.task as string;
    if (!task) {
      return false;
    }

    // Check if REMOTE_SENSING is part of it.
    return task.includes(TaskTypes.REMOTE_SENSING);
  },

  isGeoSpatial(state: Route, getters: any): boolean {
    return getters.getTrainingVariables.some(
      (v) => v.colType === GEOBOUNDS_TYPE || v.colType === GEOCOORDINATE_TYPE
    );
  },

  /* Check if the current task includes Timeseries. */
  isTimeseries(state: Route): boolean {
    // Get the list of task of the route.
    const task = state.query.task as string;
    if (!task) {
      return false;
    }

    // Check if TIME_SERIES is part of it.
    return task.includes(TaskTypes.TIME_SERIES);
  },

  getBandCombinationId(state: Route): BandID {
    const bandCombo = state.query.bandCombinationId;
    return _.isEmpty(bandCombo) ? BandID.NATURAL_COLORS : <BandID>bandCombo;
  },

  getModelTimeLimit(state: Route): number {
    const timeLimit = <string>state.query.modelTimeLimit;
    if (!timeLimit) {
      return null;
    }
    return parseInt(timeLimit, 10);
  },

  getModelLimit(state: Route): number {
    const limit = <string>state.query.modelLimit;
    if (!limit) {
      return null;
    }
    return parseInt(limit, 10);
  },

  getModelQuality(state: Route): ModelQuality {
    const qualityStr = <string>state.query.modelQuality;
    if (!qualityStr) {
      return null;
    }
    return $enum(ModelQuality).asValueOrDefault(
      qualityStr,
      ModelQuality.HIGHER_QUALITY
    );
  },

  getModelMetrics(state: Route): string[] {
    const metrics = <string>state.query.metrics;
    if (!metrics) {
      return null;
    }
    return metrics.split(",");
  },
  getRouteTimestampSplit(state: Route): number | null {
    const timestampSplit = <string>state.query.timestampSplit;
    if (!timestampSplit) {
      return null;
    }
    return parseInt(timestampSplit);
  },
  getRouteTrainTestSplit(state: Route): number {
    const trainTestSplit = <string>state.query.trainTestSplit;
    if (!trainTestSplit) {
      return null;
    }
    return parseFloat(trainTestSplit);
  },

  /* Check if the current page is SELECT_TARGET_ROUTE. */
  isPageSelectTarget(state: Route): Boolean {
    return state.path === SELECT_TARGET_ROUTE;
  },

  /* Check if the current page is SELECT_TRAINING_ROUTE. */
  isPageSelectTraining(state: Route): Boolean {
    return state.path === SELECT_TRAINING_ROUTE;
  },

  /* Get the active pane on the route */
  getRoutePane(state: Route): string {
    return state.query.pane as string;
  },

  /* Check if the current task includes a Binary classification. */
  isBinaryClassification(state: Route): boolean {
    // Get the list of task of the route.
    const task = state.query.task as string;
    if (!task) return false;

    // Check if BINARY and CLASSIFICATION are part of it.
    const isBinary = task.includes(TaskTypes.BINARY);
    const isClassification = task.includes(TaskTypes.CLASSIFICATION);

    return isBinary && isClassification;
  },

  /* Get the Binary Classification Positive Label */
  getPositiveLabel(state: Route): string {
    return (state.query?.positiveLabel as string) ?? null;
  },
};
