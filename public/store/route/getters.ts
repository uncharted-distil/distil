import {
  Variable,
  VariableSummary,
  Highlight,
  RowSelection,
  SummaryMode,
  DataMode,
  TaskTypes,
  BandID,
} from "../dataset/index";
import {
  JOIN_DATASETS_ROUTE,
  RESULTS_ROUTE,
  SELECT_TARGET_ROUTE,
  SELECT_TRAINING_ROUTE,
  JOINED_VARS_INSTANCE_PAGE,
  AVAILABLE_TARGET_VARS_INSTANCE_PAGE,
  AVAILABLE_TRAINING_VARS_INSTANCE_PAGE,
  TRAINING_VARS_INSTANCE_PAGE,
  RESULT_TRAINING_VARS_INSTANCE_PAGE,
  DATA_SIZE_DEFAULT,
  DATA_SIZE_REMOTE_SENSING_DEFAULT,
  AVAILABLE_TARGET_VARS_INSTANCE_SEARCH,
  JOINED_VARS_INSTANCE_SEARCH,
  AVAILABLE_TRAINING_VARS_INSTANCE_SEARCH,
  TRAINING_VARS_INSTANCE_SEARCH,
  RESULT_TRAINING_VARS_INSTANCE_SEARCH,
} from "../route/index";
import { ModelQuality } from "../requests/index";
import { decodeFilters, Filter, FilterParams } from "../../util/filters";
import { decodeHighlights } from "../../util/highlights";
import { decodeRowSelection } from "../../util/row";
import { Dictionary } from "../../util/dict";
import { buildLookup } from "../../util/lookup";
import { Route } from "vue-router";
import _ from "lodash";
import { $enum } from "ts-enum-util";
import { minimumRouteKey } from "../../util/data";
import { GEOBOUNDS_TYPE, GEOCOORDINATE_TYPE } from "../../util/types";

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

  getJoinDatasetsVariableSummaries(
    state: Route,
    getters: any
  ): VariableSummary[] {
    function hashSummary(datasetName: string, colName: string) {
      return `${datasetName}:${colName}`.toLowerCase();
    }

    const variables = getters.getJoinDatasetsVariables;
    const lookup = buildLookup(
      variables.map((v) => hashSummary(v.datasetName, v.colName))
    );
    const summaries = getters.getVariableSummaries;
    return summaries.filter(
      (summary) => lookup[hashSummary(summary.dataset, summary.key)]
    );
  },

  getJoinDatasetColumnA(state: Route, getters: any): string {
    return state.query.joinColumnA as string;
  },

  getJoinDatasetColumnB(state: Route, getters: any): string {
    return state.query.joinColumnB as string;
  },

  getBaseColumnSuggestions(state: Route, getters: any): string {
    return state.query.baseColumnSuggestions as string;
  },

  getJoinColumnSuggestions(state: Route, getters: any): string {
    return state.query.joinColumnSuggestions as string;
  },

  getJoinAccuracy(state: Route, getters: any): number {
    const accuracy = state.query.joinAccuracy;
    return accuracy ? _.toNumber(accuracy) : 1;
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
        const lookup = buildLookup(dataset.variables.map((v) => v.colName));
        const filtersForDataset = filters.filter((f) => {
          return lookup[f.key];
        });

        const filterParams = _.cloneDeep({
          filters: filtersForDataset,
          variables: dataset.variables.map((v) => v.colName),
        });
        res[datasetID] = filterParams;
      }
    });
    return res;
  },

  getRouteTrainingVariables(state: Route): string {
    return state.query.training ? (state.query.training as string) : null;
  },

  // Returns a boolean to say that the variables for this dataset has been ranked.
  getRouteIsTrainingVariablesRanked(state: Route): boolean {
    return state.query.varRanked && state.query.varRanked === "1"; // Use "1" for truth.
  },

  // Returns a boolean to say that the cluster for this dataset has been generated..
  getRouteIsClusterGenerated(state: Route): boolean {
    return state.query.clustering && state.query.clustering === "1"; // Use "1" for truth.
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

  getRouteTrainingVarsPage(state: Route): number {
    const pageVar = TRAINING_VARS_INSTANCE_PAGE;
    return state.query[pageVar] ? _.toNumber(state.query[pageVar]) : 1;
  },

  getRouteResultTrainingVarsPage(state: Route): number {
    const pageVar = RESULT_TRAINING_VARS_INSTANCE_PAGE;
    return state.query[pageVar] ? _.toNumber(state.query[pageVar]) : 1;
  },

  getAllRoutePages(state: Route, getters: any): Object {
    const pages = {};
    pages[JOIN_DATASETS_ROUTE] = [getters.getRouteJoinDatasetsVarsPage];
    pages[SELECT_TARGET_ROUTE] = [getters.getRouteAvailableTargetVarsPage];
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

  getAllSearchesByRoute(state: Route, getters: any): Object {
    const searches = {};
    searches[JOIN_DATASETS_ROUTE] = [getters.getRouteJoinDatasetsVarsSearch];
    searches[SELECT_TARGET_ROUTE] = [getters.getRouteAvailableTargetVarsSearch];
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
    return decodeFilters(state.query.filters as string);
  },

  getDecodedSolutionRequestFilterParams(
    state: Route,
    getters: any
  ): FilterParams {
    const filters = getters.getDecodedFilters;
    const size = getters.getRouteDataSize;
    const filterParams = _.cloneDeep({
      highlight: null,
      variables: [],
      filters,
      size,
    });
    // add training vars
    const training = getters.getDecodedTrainingVariableNames;
    filterParams.variables = filterParams.variables.concat(training);
    // add target vars
    const target = getters.getRouteTargetVariable as string;
    if (target) {
      filterParams.variables.push(target);
    }
    return filterParams;
  },

  getDecodedHighlight(state: Route): Highlight {
    return decodeHighlights(state.query.highlights as string);
  },

  getDecodedRowSelection(state: Route): RowSelection {
    return decodeRowSelection(state.query.row as string);
  },

  getTrainingVariables(state: Route, getters: any): Variable[] {
    const training = getters.getDecodedTrainingVariableNames;
    const lookup = buildLookup(training);
    const variables = getters.getVariables;
    return variables.filter(
      (variable) => lookup[variable.colName.toLowerCase()]
    );
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
        (summary) => target.toLowerCase() === summary.colName.toLowerCase()
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
    return variables.filter(
      (variable) => !lookup[variable.colName.toLowerCase()]
    );
  },

  getActiveSolutionIndex(state: Route, getters: any): number {
    const solutionId = getters.getRouteSolutionId;
    const solutions = getters.getSolutions;
    return _.findIndex(solutions, (solution: any) => {
      return solution.solutionId === solutionId;
    });
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

  getRouteTrainTestSplit(state: Route): number {
    const trainTestSplit = <string>state.query.trainTestSplit;
    if (!trainTestSplit) {
      return 0.9;
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
};
