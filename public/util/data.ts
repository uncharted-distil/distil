import axios from "axios";
import sha1 from "crypto-js/sha1";
import _ from "lodash";
import Vue from "vue";
import {
  D3M_INDEX_FIELD,
  DatasetPendingRequestType,
  SummaryMode,
  TableColumn,
  TableData,
  TableRow,
  TimeseriesGrouping,
  Variable,
  VariableSummary,
} from "../store/dataset/index";
import {
  actions as datasetActions,
  getters as datasetGetters,
  mutations as datasetMutations,
} from "../store/dataset/module";
import { PredictionContext } from "../store/predictions/actions";
import {
  Predictions,
  PredictStatus,
  Solution,
  SolutionStatus,
} from "../store/requests/index";
import {
  getters as predictionsGetters,
  mutations as predictionsMutations,
} from "../store/predictions/module";
import { getters as requestGetters } from "../store/requests/module";
import { ResultsContext } from "../store/results/actions";
import {
  actions as resultsActions,
  getters as resultsGetters,
  mutations as resultsMutations,
} from "../store/results/module";
import { getters as routeGetters } from "../store/route/module";
import store from "../store/store";
import {
  CLUSTER_PREFIX,
  formatValue,
  hasComputedVarPrefix,
  IMAGE_TYPE,
  isGeoLocatedType,
  isIntegerType,
  isImageType,
  isLatitudeGroupType,
  isListType,
  isLongitudeGroupType,
  isTimeGroupType,
  isTimeType,
  isValueGroupType,
  LATITUDE_TYPE,
  LONGITUDE_TYPE,
  MULTIBAND_IMAGE_TYPE,
  TIMESERIES_TYPE,
} from "../util/types";
import { Dictionary } from "./dict";
import { FilterParams } from "./filters";
import { overlayRouteEntry } from "./routes";
import { Location } from "vue-router";
import {
  interpolateTurbo,
  interpolateViridis,
  interpolateInferno,
  interpolateMagma,
  interpolatePlasma,
} from "d3-scale-chromatic";
// Postfixes for special variable names
export const PREDICTED_SUFFIX = "_predicted";
export const ERROR_SUFFIX = "_error";

// constants for accessing variable summaries
export const VARIABLE_SUMMARY_BASE = "summary";
export const VARIABLE_SUMMARY_CONFIDENCE = "confidence";

export const NUM_PER_PAGE = 10;
export const NUM_PER_TARGET_PAGE = 9;
export const NUM_PER_DATA_EXPLORER_PAGE = 3;

export const DATAMART_PROVENANCE_NYU = "NYU";
export const DATAMART_PROVENANCE_ISI = "ISI";
export const ELASTIC_PROVENANCE = "elastic";
export const FILE_PROVENANCE = "file";

export const IMPORTANT_VARIABLE_RANKING_THRESHOLD = 0.5;

export const LOW_SHOT_LABEL_COLUMN_NAME = "LowShotLabel";
// LowShotLabels enum for labeling data in a binary classification
export enum LowShotLabels {
  positive = "positive",
  negative = "negative",
  unlabeled = "unlabeled",
}
// DatasetUpdate is an interface that contains the data to update existing data
export interface DatasetUpdate {
  index: string; // d3mIndex
  name: string; // colName
  value: string; // new value to replace old value
}

export interface TimeIntervals {
  value: number;
  text: string;
}

// ColorScaleNames is an enum that contains all the supported color scale names. Can be used to access COLOR_SCALES functions
export enum ColorScaleNames {
  viridis = "viridis",
  magma = "magma",
  inferno = "inferno",
  plasma = "plasma",
  turbo = "turbo",
}
// COLOR_SCALES contains the color scalefunctions that are js. This is for wrapping it in typescript.
export const COLOR_SCALES: Map<
  ColorScaleNames,
  (t: number) => string
> = new Map([
  [ColorScaleNames.viridis, interpolateViridis],
  [ColorScaleNames.magma, interpolateMagma],
  [ColorScaleNames.inferno, interpolateInferno],
  [ColorScaleNames.plasma, interpolatePlasma],
  [ColorScaleNames.turbo, interpolateTurbo],
]);
export interface ScoreInfo {
  d3mIndex: number;
  score: number;
}
// BinarySets contains the ranked data for the LowShotLabel binary classification
export interface RankedSet {
  data: ScoreInfo[];
}
export interface BinaryScoreResponse {
  progress: {
    ranked: string[][];
    colInfo: { d3mIndex: string; score: string };
  };
}
export function parseBinaryScoreResponse(res: BinaryScoreResponse): RankedSet {
  // check for properties
  if (!res.progress) {
    return null;
  }
  const result = res.progress;
  if (!result.ranked || !result.colInfo) {
    return null;
  }
  const d3mIndex = result.colInfo.d3mIndex;
  const scoreIndex = parseInt(result.colInfo.score);
  const rankedSet: RankedSet = { data: [] };
  result.ranked.forEach((p) => {
    rankedSet.data.push({
      d3mIndex: p[d3mIndex],
      score: parseFloat(p[scoreIndex]),
    });
  });
  return rankedSet;
}
export function getTimeseriesSummaryTopCategories(
  summary: VariableSummary
): string[] {
  return _.map(summary.baseline.categoryBuckets, (buckets, category) => {
    return {
      category: category,
      count: _.sumBy(buckets, (b) => b.count),
    };
  })
    .sort((a, b) => b.count - a.count)
    .map((c) => c.category);
}
export function getRandomInt(max: number): number {
  return Math.floor(Math.random() * Math.floor(max));
}
export function getTimeseriesGroupingsFromFields(
  variables: Variable[],
  fields: Dictionary<TableColumn>
): TimeseriesGrouping[] {
  // Check to see if any of the fields are the ID column of one of our variables
  const fieldKeys = _.map(fields, (_, key) => key);
  return variables
    .filter(
      (v) =>
        v.grouping &&
        v.grouping.idCol &&
        v.colType === TIMESERIES_TYPE &&
        _.includes(fieldKeys, v.grouping.idCol)
    )
    .map((v) => v.grouping as TimeseriesGrouping);
}

export function getComposedVariableKey(keys: string[]): string {
  return "__grouping_" + keys.join("_");
}

export function getTimeseriesAnalysisIntervals(
  timeVar: Variable,
  range: number
): TimeIntervals[] {
  const SECONDS_VALUE = 1;
  const MINUTES_VALUE = SECONDS_VALUE * 60;
  const HOURS_VALUE = MINUTES_VALUE * 60;
  const DAYS_VALUE = HOURS_VALUE * 24;
  const WEEKS_VALUE = DAYS_VALUE * 7;
  const MONTHS_VALUE = WEEKS_VALUE * 4;
  const YEARS_VALUE = MONTHS_VALUE * 12;
  const SECONDS_LABEL = "Seconds";
  const MINUTES_LABEL = "Minutes";
  const HOURS_LABEL = "Hours";
  const DAYS_LABEL = "Days";
  const WEEKS_LABEL = "Weeks";
  const MONTHS_LABEL = "Months";
  const YEARS_LABEL = "Years";

  if (isTimeType(timeVar.colType)) {
    if (range < DAYS_VALUE) {
      return [
        { value: SECONDS_VALUE, text: SECONDS_LABEL },
        { value: MINUTES_VALUE, text: MINUTES_LABEL },
        { value: HOURS_VALUE, text: HOURS_LABEL },
      ];
    } else if (range < 2 * WEEKS_VALUE) {
      return [
        { value: HOURS_VALUE, text: HOURS_LABEL },
        { value: DAYS_VALUE, text: DAYS_LABEL },
        { value: WEEKS_VALUE, text: WEEKS_LABEL },
      ];
    } else if (range < MONTHS_VALUE) {
      return [
        { value: HOURS_VALUE, text: HOURS_LABEL },
        { value: DAYS_VALUE, text: DAYS_LABEL },
        { value: WEEKS_VALUE, text: WEEKS_LABEL },
      ];
    } else if (range < 4 * MONTHS_VALUE) {
      return [
        { value: DAYS_VALUE, text: DAYS_LABEL },
        { value: WEEKS_VALUE, text: WEEKS_LABEL },
        { value: MONTHS_VALUE, text: MONTHS_LABEL },
      ];
    } else if (range < YEARS_VALUE) {
      return [
        { value: WEEKS_VALUE, text: WEEKS_LABEL },
        { value: MONTHS_VALUE, text: MONTHS_LABEL },
      ];
    } else {
      return [
        { value: MONTHS_VALUE, text: MONTHS_LABEL },
        { value: YEARS_VALUE, text: YEARS_LABEL },
      ];
    }
  }

  let small = 0;
  let med = 0;
  let large = 0;
  if (isIntegerType(timeVar.colType)) {
    small = Math.floor(range / 10);
    med = Math.floor(range / 20);
    large = Math.floor(range / 40);
  } else {
    small = range / 10.0;
    med = range / 20.0;
    large = range / 40.0;
  }
  return [
    { value: small, text: `${small}` },
    { value: med, text: `${med}` },
    { value: large, text: `${large}` },
  ];
}

export function fetchSummaryExemplars(
  datasetName: string,
  variableName: string,
  summary: VariableSummary
) {
  const variables = datasetGetters.getVariables(store);
  const variable = variables.find((v) => v.colName === variableName);

  const baselineExemplars = summary.baseline.exemplars;
  const filteredExemplars =
    summary.filtered && summary.filtered.exemplars
      ? summary.filtered.exemplars
      : null;
  const exemplars = filteredExemplars ? filteredExemplars : baselineExemplars;

  if (exemplars) {
    if (variable.grouping) {
      if (variable.grouping.type === TIMESERIES_TYPE) {
        // if there a linked exemplars, fetch those before resolving
        const solutionId = routeGetters.getRouteSolutionId(store);
        const grouping = variable.grouping as TimeseriesGrouping;
        const args = {
          dataset: datasetName,
          timeseriesColName: grouping.idCol,
          xColName: grouping.xCol,
          yColName: grouping.yCol,
          timeseriesIds: exemplars,
          solutionId: solutionId,
        };
        return () => {
          if (solutionId) {
            return resultsActions.fetchForecastedTimeseries(store, args);
          } else {
            return datasetActions.fetchTimeseries(store, args);
          }
        };
      }
    } else {
      // if there are linked files, fetch some of them before resolving
      return datasetActions.fetchFiles(store, {
        dataset: datasetName,
        variable: variableName,
        urls: exemplars.slice(0, 5),
      });
    }
  }

  return new Promise<void>((res) => res());
}

export function fetchResultExemplars(
  datasetName: string,
  variableName: string,
  key: string,
  solutionId: string,
  summary: VariableSummary
) {
  const variables = datasetGetters.getVariables(store);
  const variable = variables.find((v) => v.colName === variableName);

  const baselineExemplars = summary.baseline?.exemplars;
  const filteredExemplars = summary.filtered?.exemplars;
  const exemplars = filteredExemplars ? filteredExemplars : baselineExemplars;

  if (exemplars) {
    if (variable.grouping) {
      if (variable.grouping.type === TIMESERIES_TYPE) {
        const grouping = variable.grouping as TimeseriesGrouping;
        // if there a linked exemplars, fetch those before resolving
        return resultsActions.fetchForecastedTimeseries(store, {
          dataset: datasetName,
          timeseriesColName: grouping.idCol,
          xColName: grouping.xCol,
          yColName: grouping.yCol,
          timeseriesIds: exemplars,
          solutionId: solutionId,
        });
      }
    } else {
      // if there a linked files, fetch those before resolving
      return datasetActions.fetchFiles(store, {
        dataset: datasetName,
        variable: variableName,
        urls: exemplars,
      });
    }
  }

  return new Promise<void>((res) => res());
}

/*
  minimumRouteKey - Makes a unique key given route state to support
  saving to and retrieving from a variable summary dictionary cache
  using some of the route's query options (IE: not grabbing all
  options as that's too narrow in focus.) It SHA1 hashes a string
  of datasetId, solutionId, requestId, fittedSolutionId, highlight,
  filters, dataMode, varModes, active pane, and ranking as that's unique
  enough without being over specific and causing duplicate calls.
  The SHA1 hash of those fields is fast to calculate, maintains uniqueness,
  and keeps the store keys a consistent length, unlike base64.
*/
export function minimumRouteKey(): string {
  const routeKeys =
    JSON.stringify(routeGetters.getRouteDataset(store)) +
    JSON.stringify(routeGetters.getRouteSolutionId(store)) +
    JSON.stringify(routeGetters.getRouteProduceRequestId(store)) +
    JSON.stringify(routeGetters.getRouteFittedSolutionId(store)) +
    JSON.stringify(routeGetters.getRouteHighlight(store)) +
    JSON.stringify(routeGetters.getRouteFilters(store)) +
    JSON.stringify(routeGetters.getDataMode(store)) +
    JSON.stringify(routeGetters.getDecodedVarModes(store)) +
    "pane" +
    routeGetters.getRoutePane(store) +
    "ranked" +
    JSON.stringify(routeGetters.getRouteIsTrainingVariablesRanked(store));
  const sha1rk = sha1(routeKeys);
  return sha1rk;
}

export function updateSummaries(
  summary: VariableSummary,
  summaries: VariableSummary[]
) {
  const index = _.findIndex(summaries, (s) => {
    return s.dataset === summary.dataset && s.key === summary.key;
  });
  if (index >= 0) {
    // freezing the return to prevent slow, unnecessary deep reactivity.
    Vue.set(summaries, index, Object.freeze(summary));
  } else {
    summaries.push(Object.freeze(summary));
  }
}
export async function cloneDatasetUpdateRoute(): Promise<Location> {
  const dataset = routeGetters.getRouteDataset(store);
  // clone the current dataset
  const clonedInfo = await datasetActions.cloneDataset(store, {
    dataset,
  });
  // if null there was an error, if success is false there was a backend issue
  if (clonedInfo === null || !clonedInfo.success) {
    return null;
  }
  // update route to new cloned dataset name
  const entry = overlayRouteEntry(routeGetters.getRoute(store), {
    dataset: clonedInfo.clonedDatasetName,
  });
  return entry;
}
export function updateSummariesPerVariable(
  summary: VariableSummary,
  variableSummaryDictionary: Dictionary<Dictionary<VariableSummary>>
) {
  const routeKey = minimumRouteKey();
  const summaryKey = summary.key;
  // check for existing summaries for that variable, if not, instantiate
  if (!variableSummaryDictionary[summaryKey]) {
    Vue.set(variableSummaryDictionary, summaryKey, {});
  }
  // freezing the return to prevent slow, unnecessary deep reactivity.
  Vue.set(
    variableSummaryDictionary[summaryKey],
    routeKey,
    Object.freeze(summary)
  );
}
// removeTimeseries will not trigger a Vue update
export function removeTimeseries(
  args: {
    solutionId?: string;
    predictionsId?: string;
    dataset?: string;
  },
  items: TableRow[],
  uniqueTrail?: string
) {
  let fields = null;
  let mutator = null;
  const predDataset = routeGetters.getRoutePredictionsDataset(store);
  const solutionId = routeGetters.getRouteSolutionId(store);

  if (predDataset !== null) {
    // check if prediction view
    fields = predictionsGetters.getIncludedPredictionTableDataFields(store);
    mutator = predictionsMutations.removeTimeseries;
  } else if (solutionId !== null) {
    // check if result view
    fields = resultsGetters.getIncludedResultTableDataFields(store);
    mutator = resultsMutations.removeTimeseries;
  } else {
    // defaults to select view
    fields = datasetGetters.getIncludedTableDataFields(store);
    mutator = datasetMutations.removeTimeseries;
  }
  const variables = datasetGetters.getVariables(store);
  const timeseriesGroupings = getTimeseriesGroupingsFromFields(
    variables,
    fields
  );
  timeseriesGroupings.forEach((tsg) => {
    mutator(store, {
      ...args,
      ids: items.map((item) => {
        return (item[tsg.idCol].value as string) + (uniqueTrail ?? "");
      }),
    });
  });
}
export function removeSummary(
  summary: VariableSummary,
  summaries: VariableSummary[]
) {
  const index = _.findIndex(summaries, (s) => {
    return s.dataset === summary.dataset && s.key === summary.key;
  });
  if (index >= 0) {
    Vue.delete(summaries, index);
  }
}

export function filterVariablesByFeature(variables: Variable[]): Variable[] {
  // need to exclude the hidden variables
  const groupingVars = variables.filter(
    (v) => v.distilRole === "grouping" && v.grouping !== null
  );
  const hiddenFlat = [].concat.apply(
    [],
    groupingVars.map((v) =>
      [].concat(v.grouping.hidden).concat(v.grouping.subIds)
    )
  );
  const hidden = new Map(hiddenFlat.map((v) => [v, v]));

  // the groupings that hide variables are themselves variables to display
  const groupingDisplayed = new Map(groupingVars.map((v) => [v.colName, v]));

  return variables.filter(
    (v) =>
      (v.distilRole === "data" && !hidden.has(v.colName)) ||
      groupingDisplayed.has(v.colName)
  );
}

export function filterSummariesByDataset(
  summaries: VariableSummary[],
  dataset: string
): VariableSummary[] {
  return summaries.filter((summary) => {
    return summary.dataset === dataset && !hasComputedVarPrefix(summary.key);
  });
}

export function createEmptyTableData(): TableData {
  return {
    numRows: 0,
    columns: [],
    values: [],
    fittedSolutionId: null,
    produceRequestId: null,
  };
}

export function formatSlot(key: string, slotType: string): string {
  return `${slotType}(${key})`;
}

export function formatFieldsAsArray(
  fields: Dictionary<TableColumn>
): TableColumn[] {
  return _.map(fields, (field) => field);
}

export function createPendingSummary(
  key: string,
  label: string,
  description: string,
  dataset: string
): VariableSummary {
  return {
    key: key,
    label: label,
    description: description,
    dataset: dataset,
    pending: true,
    baseline: null,
    filtered: null,
  };
}

export function createErrorSummary(
  key: string,
  label: string,
  dataset: string,
  error: any
): VariableSummary {
  return {
    key: key,
    label: label,
    description: null,
    dataset: dataset,
    baseline: null,
    filtered: null,
    err: error.response ? error.response.data : error,
  };
}

export async function fetchSolutionResultSummary(
  context: ResultsContext,
  endpoint: string,
  solution: Solution,
  key: string,
  label: string,
  resultProperty: string,
  resultSummaries: VariableSummary[],
  updateFunction: (arg: ResultsContext, summary: VariableSummary) => void,
  filterParams: FilterParams,
  varMode: SummaryMode
): Promise<any> {
  const dataset = solution.dataset;
  const solutionId = solution.solutionId;
  const target = solution.feature;
  const resultId = solution.resultId;

  const exists = _.find(
    resultSummaries,
    (v) => v.dataset === dataset && v.key === key
  );
  if (!exists) {
    // add placeholder
    updateFunction(context, createPendingSummary(key, label, "", dataset));
  }

  // fetch the results for each solution
  if (solution.progress !== SolutionStatus.SOLUTION_COMPLETED) {
    // skip
    return;
  }
  // finish building endpoint
  const completeEndpoint = varMode
    ? `${endpoint}/${resultId}/${varMode}`
    : `${endpoint}/${resultId}`;

  // return promise
  try {
    const response = await axios.post(
      completeEndpoint,
      filterParams ? filterParams : {}
    );
    // save the histogram data
    const summary = response.data[resultProperty];
    await fetchResultExemplars(dataset, target, key, solutionId, summary);
    summary.solutionId = solutionId;
    summary.dataset = dataset;
    updateFunction(context, summary);
  } catch (error) {
    console.error(error);
    updateFunction(context, createErrorSummary(key, label, dataset, error));
  }
}

export async function fetchPredictionResultSummary(
  context: PredictionContext,
  endpoint: string,
  predictions: Predictions,
  key: string,
  label: string,
  resultSummaries: VariableSummary[],
  updateFunction: (arg: PredictionContext, summary: VariableSummary) => void,
  filterParams: FilterParams,
  varMode: SummaryMode
): Promise<any> {
  const dataset = predictions.dataset;
  const resultId = predictions.resultId;

  const exists = _.find(
    resultSummaries,
    (v) => v.dataset === dataset && v.key === key
  );
  if (!exists) {
    // add placeholder
    updateFunction(context, createPendingSummary(key, label, "", dataset));
  }

  // fetch the results for each solution
  if (predictions.progress !== PredictStatus.PREDICT_COMPLETED) {
    // skip
    return;
  }

  // return promise
  try {
    const response = await axios.post(
      `${endpoint}/${resultId}/${varMode}`,
      filterParams ? filterParams : {}
    );
    const summary = <VariableSummary>response.data.summary;
    summary.dataset = dataset;
    updateFunction(context, summary);
  } catch (error) {
    console.error(error);
    updateFunction(context, createErrorSummary(key, label, dataset, error));
  }
}

export function filterArrayByPage(
  pageIndex: number,
  pageSize: number,
  items: any[]
): any[] {
  if (items.length > pageSize) {
    const firstIndex = pageSize * (pageIndex - 1);
    const lastIndex = Math.min(firstIndex + pageSize, items.length);
    return items.slice(firstIndex, lastIndex);
  }
  return items;
}

export function searchVariables(
  variables: Variable[],
  searchQuery: string
): Variable[] {
  return variables.filter((v) => {
    return (
      searchQuery === undefined ||
      searchQuery === "" ||
      (v && v.colName.toLowerCase().includes(searchQuery.toLowerCase()))
    );
  });
}
export function sortVariablesByPCARanking(variables: Variable[]): Variable[] {
  variables.sort((a, b) => {
    return b.importance - a.importance;
  });
  return variables;
}
export function sortVariablesByImportance(variables: Variable[]): Variable[] {
  // prioritize FI over MI
  const datasetName = routeGetters.getRouteDataset(store);
  const solutionId = routeGetters.getRouteSolutionId(store);
  // set rankMap for MI
  let rankMap = datasetGetters.getVariableRankings(store)[datasetName];
  // check if FI
  if (solutionId !== null) {
    if (resultsGetters.getFeatureImportanceRanking(store)[solutionId]) {
      rankMap = resultsGetters.getFeatureImportanceRanking(store)[solutionId];
    }
  }
  // Fallback to PCA if none of the above is available
  if (!rankMap) {
    return variables.sort((a, b) => {
      return b.importance - a.importance;
    });
  }
  variables.sort((a, b) => {
    return rankMap[b.colName] - rankMap[a.colName];
  });
  return variables;
}

export function getVariableSummariesByState(
  pageIndex: number,
  pageSize: number,
  variables: Variable[],
  summaryDictionary: Dictionary<Dictionary<VariableSummary>>,
  isSorted = false
) {
  const routeKey = minimumRouteKey();
  const ranked =
    routeGetters.getRouteIsTrainingVariablesRanked(store) || isSorted;

  if (!(Object.keys(summaryDictionary).length > 0 && variables.length > 0)) {
    return [];
  }

  // remove any pattern cluster variables
  let sortedVariables = variables.filter((sv) => {
    return sv.colName.indexOf(CLUSTER_PREFIX) < 0;
  });

  if (ranked) {
    // prioritize FI over MI
    sortedVariables = sortVariablesByImportance(sortedVariables);
  }

  // select only the current variables on the page
  sortedVariables = filterArrayByPage(pageIndex, pageSize, sortedVariables);

  // map them back to the variable summary dictionary for the current route key
  const currentSummaries = sortedVariables.reduce((cs, vn) => {
    if (!summaryDictionary[vn.colName]) {
      const placeholder = createPendingSummary(
        vn.colName,
        vn.colDisplayName,
        vn.colDescription,
        vn.datasetName
      );
      cs.push(placeholder);
    } else {
      if (summaryDictionary[vn.colName][routeKey]) {
        cs.push(summaryDictionary[vn.colName][routeKey]);
      } else {
        const tempVariableSummaryKey = Object.keys(
          summaryDictionary[vn.colName]
        )[0];
        cs.push(summaryDictionary[vn.colName][tempVariableSummaryKey]);
      }
    }
    return cs;
  }, []);

  return currentSummaries;
}

export function getVariableImportance(v: Variable): number {
  const solutionID = routeGetters.getRouteSolutionId(store);
  const map = resultsGetters.getFeatureImportanceRanking(store)[solutionID];
  return map[v.colName];
}

export function getVariableRanking(v: Variable): number {
  const datasetName = routeGetters.getRouteDataset(store);
  const map = datasetGetters.getVariableRankings(store)[datasetName]; // get MI ranking map
  if (!map) {
    return v.importance; // if MI ranking does not exist default to PCA
  }
  return map[v.colName];
}

export function getSolutionFeatureImportance(
  v: Variable,
  solutionID: string
): number {
  const solutionRanks = resultsGetters.getFeatureImportanceRanking(store)[
    solutionID
  ];
  if (solutionRanks) {
    return solutionRanks[v.colName];
  }
  return null;
}

export function sortSolutionSummariesByImportance(
  summaries: VariableSummary[],
  solutionID: string
): VariableSummary[] {
  // create importance lookup map
  const importance: Dictionary<number> = resultsGetters.getFeatureImportanceRanking(
    store
  )[solutionID];
  if (!importance) {
    return null;
  }
  // sort by importance
  summaries.sort((a, b) => {
    if (!importance[b.key]) {
      return -1;
    }
    if (!importance[a.key]) {
      return 1;
    }
    return importance[b.key] - importance[a.key];
  });
  return summaries;
}

export function validateData(data: TableData) {
  return (
    !_.isEmpty(data) && !_.isEmpty(data.values) && !_.isEmpty(data.columns)
  );
}

export function getTableDataItems(data: TableData): TableRow[] {
  if (validateData(data)) {
    // convert fetched result data rows into table data rows
    const formattedTable = data.values.map((resultRow, rowIndex) => {
      const row = {} as TableRow;
      resultRow.forEach((colValue, colIndex) => {
        const colName = data.columns[colIndex].key;
        const colType = data.columns[colIndex].type;
        if (colName !== "d3mIndex") {
          row[colName] = {};
          row[colName].value = formatValue(colValue.value, colType);
          if (colValue.weight !== null && colValue.weight !== undefined) {
            row[colName].weight = colValue.weight;
          }
          if (colValue.confidence !== undefined) {
            const conKey = "confidence";
            row[conKey] = {};
            row[conKey].value = colValue.confidence;
          }
        } else {
          row[colName] = formatValue(colValue.value, colType);
        }
      });
      row._key = rowIndex;
      row._rowVariant = null;
      return Object.seal(row);
    });

    return formattedTable;
  }
  return !_.isEmpty(data) ? [] : null;
}

function isPredictedCol(arg: string): boolean {
  return arg.endsWith(":predicted");
}

function isErrorCol(arg: string): boolean {
  return arg.endsWith(":error");
}

export function getTableDataFields(data: TableData): Dictionary<TableColumn> {
  if (validateData(data)) {
    const result: Dictionary<TableColumn> = {};
    const variables = datasetGetters.getVariablesMap(store);

    for (const col of data.columns) {
      if (col.key === D3M_INDEX_FIELD) {
        continue;
      }

      // Error and predicted columns require unique handling.  They use a special key of the format
      // <solution_id>:<predicted|error> and are not available in the variables list.
      let variable: Variable = null;
      let description: string = null;
      let label: string = null;
      if (isPredictedCol(col.key)) {
        variable = requestGetters.getActiveSolutionTargetVariable(store)[0]; // always a single value
        label = variable.colDisplayName;
        description = `Model predicted value for ${variable.colName}`;

        result.confidence = {
          label: "Confidence",
          key: "confidence",
          type: "numeric",
          weight: null,
          headerTitle: `Prediction confidence ${variable.colName}`,
          sortable: true,
        };
      } else if (isErrorCol(col.key)) {
        variable = requestGetters.getActiveSolutionTargetVariable(store)[0];
        label = "Error";
        description = `Difference between actual and predicted value for ${variable.colName}`;
      } else {
        variable = variables[col.key];
        label = col.label;
        if (variable) {
          description = variable.colDescription;
        }
      }

      result[col.key] = {
        label: label,
        key: col.key,
        type: col.type,
        weight: col.weight,
        headerTitle: description ? label.concat(": ", description) : label,
        sortable: true,
      };
    }

    return result;
  }
  return {};
}

export function isDatamartProvenance(provenance: string): boolean {
  return (
    provenance === DATAMART_PROVENANCE_NYU ||
    provenance === DATAMART_PROVENANCE_ISI
  );
}

// Validates argument object based on input array of expected object fields
// if there's invalid members, it logs warning with the invalid members and
// returns false. Returns true otherwise.
export function validateArgs(args: object, expectedArgs: string[]) {
  const missingArgs = expectedArgs.reduce((missing, arg) => {
    if (args[arg] === undefined || args[arg] === null) missing.push(arg);
    return missing;
  }, []);
  if (missingArgs.length === 0) {
    return true;
  } else {
    console.warn(`${missingArgs} argument(s) are missing`);
    return false;
  }
}

// Computes the cell colour based on the
export function explainCellColor(
  weight: number,
  data: any,
  tableFields: TableColumn[],
  dataItems: TableRow[]
): string {
  if (!weight || !hasMultipleFeatures(tableFields)) {
    return "";
  }

  const absoluteWeight = Math.abs(
    weight /
      d3mRowWeightExtrema(tableFields, dataItems)[data.item[D3M_INDEX_FIELD]]
  );

  return `background: ${colorByWeight(absoluteWeight)}`;
}

export function colorByWeight(weight: number): string {
  const red = 255 - 128 * weight;
  const green = 255 - 64 * weight;
  return `rgba(${red}, ${green}, 255, .75)`;
}

function hasMultipleFeatures(tableFields: TableColumn[]): boolean {
  const featureNames = tableFields.reduce((uniqueNames, field) => {
    uniqueNames[field.label] = true;
    return uniqueNames;
  }, {});
  return Object.keys(featureNames).length > 2;
}

function d3mRowWeightExtrema(
  tableFields: TableColumn[],
  dataItems: TableRow[]
): Dictionary<number> {
  return dataItems.reduce((extremas, item) => {
    extremas[item[D3M_INDEX_FIELD]] = tableFields.reduce((rowMax, tableCol) => {
      if (item[tableCol.key].weight) {
        const currentWeight = Math.abs(item[tableCol.key].weight);
        return currentWeight > rowMax ? currentWeight : rowMax;
      } else {
        return rowMax;
      }
    }, 0);
    return extremas;
  }, {});
}

export function getImageFields(
  fields: Dictionary<TableColumn>
): { key: string; type: string }[] {
  // find basic image fields
  const imageFields = _.map(fields, (field, key) => {
    return {
      key: key,
      type: field.type,
    };
  }).filter((field) => field.type === IMAGE_TYPE);

  // find remote senings image fields
  const fieldKeys = _.map(fields, (_, key) => key);
  const MultiBandImageFields = datasetGetters
    .getVariables(store)
    .filter(
      (v) =>
        v.grouping &&
        v.grouping.idCol &&
        v.colType === MULTIBAND_IMAGE_TYPE &&
        _.includes(fieldKeys, v.grouping.idCol)
    )
    .map((v) => ({ key: v.grouping.idCol, type: v.colType }));

  // the two are probably mutually exclusive, but it doesn't hurt anything to allow for both
  return imageFields.concat(MultiBandImageFields);
}

export function getListFields(
  fields: Dictionary<TableColumn>
): { key: string; type: string }[] {
  return _.filter(fields, (f) => isListType(f.type)).map((f) => ({
    key: f.key,
    type: f.type,
  }));
}

export function shouldRunMi(dataset: string): boolean {
  // check if data exists
  if (datasetGetters.getVariableRankings(store)[dataset]) {
    return false;
  }
  // check previous requests
  const updates = datasetGetters
    .getPendingRequests(store)
    .filter((update) => update.dataset === dataset);
  // if none, ranking should be called
  if (!updates.length) {
    return true;
  }
  const size = updates.filter((u) => {
    return u.type === DatasetPendingRequestType.VARIABLE_RANKING;
  }).length;
  // if no previous variable ranking request
  if (size) {
    return false;
  }
  // default to true if all the above does not return
  return true;
}

export function hasTimeseriesFeatures(variables: Variable[]): boolean {
  const valueColumns = variables.filter((v) => isValueGroupType(v.colType));
  const timeColumns = variables.filter((v) => isTimeGroupType(v.colType));

  if (
    (valueColumns.length === 1 &&
      timeColumns.length === 1 &&
      valueColumns[0].colName !== timeColumns[0].colName) ||
    (valueColumns.length > 1 && timeColumns.length > 0) ||
    (valueColumns.length > 0 && timeColumns.length > 1)
  ) {
    return true;
  }

  return false;
}

export function hasGeoordinateFeatures(variables: Variable[]): boolean {
  const latColumns = variables.filter((v) => isLatitudeGroupType(v.colType));
  const lonColumns = variables.filter((v) => isLongitudeGroupType(v.colType));
  if (
    (latColumns.length === 1 &&
      lonColumns.length === 1 &&
      latColumns[0].colName !== lonColumns[0].colName) ||
    (latColumns.length > 1 && lonColumns.length > 0) ||
    (latColumns.length > 0 && lonColumns.length > 1)
  ) {
    return true;
  }

  return false;
}

export function hasGeoFeatures(variables: Variable[]): boolean {
  const hasLat = variables.some((v) => v.colType === LONGITUDE_TYPE);
  const hasLon = variables.some((v) => v.colType === LATITUDE_TYPE);
  const hasGeocoord = variables.some(
    (v) => v.grouping && isGeoLocatedType(v.grouping.type)
  );
  return (hasLat && hasLon) || hasGeocoord;
}

export function hasImageFeatures(variables: Variable[]): boolean {
  return variables.some((v) => isImageType(v.colType));
}
