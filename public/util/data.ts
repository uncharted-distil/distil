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

import area from "@turf/area";
import { polygon } from "@turf/helpers";
import axios from "axios";
import sha1 from "crypto-js/sha1";
import _ from "lodash";
import localStorage from "store";
import Vue from "vue";
import VueRouter, { Location } from "vue-router";
import router from "../router/router";
import {
  D3M_INDEX_FIELD,
  DatasetPendingRequestType,
  GeoBoundsGrouping,
  SummaryMode,
  TableColumn,
  TableData,
  TableRow,
  TaskTypes,
  TimeseriesGrouping,
  Variable,
  VariableSummary,
  VariableSummaryKey,
  VariableSummaryResp,
} from "../store/dataset/index";
import {
  actions as datasetActions,
  getters as datasetGetters,
  mutations as datasetMutations,
} from "../store/dataset/module";
import { PredictionContext } from "../store/predictions/actions";
import {
  getters as predictionsGetters,
  mutations as predictionsMutations,
} from "../store/predictions/module";
import {
  Predictions,
  PredictStatus,
  Solution,
  SolutionStatus,
} from "../store/requests/index";
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
  DISTIL_ROLES,
  Field,
  formatValue,
  GEOBOUNDS_TYPE,
  hasComputedVarPrefix,
  IMAGE_TYPE,
  isGeoLocatedType,
  isImageType,
  isIntegerType,
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
import { Group } from "./facets";
import { FilterParams, FilterSetsParams, removeFiltersByName } from "./filters";
import { overlayRouteEntry, varModesToString } from "./routes";
import { findBestMatch } from "string-similarity";

// Postfixes for special variable names
export const PREDICTED_SUFFIX = "_predicted";
export const ERROR_SUFFIX = "_error";

// constants for accessing variable summaries
export const VARIABLE_SUMMARY_BASE = "summary";
export const VARIABLE_SUMMARY_CONFIDENCE = "confidence";
export const VARIABLE_SUMMARY_RANKING = "rank";

export const NUM_PER_PAGE = 10;
export const NUM_PER_TARGET_PAGE = 9;

export const DATAMART_PROVENANCE_NYU = "NYU";
export const DATAMART_PROVENANCE_ISI = "ISI";
export const ELASTIC_PROVENANCE = "elastic";
export const FILE_PROVENANCE = "file";

export const IMPORTANT_VARIABLE_RANKING_THRESHOLD = 0.5;
export const LOW_SHOT_SCORE_COLUMN_PREFIX = "__query_";
export const LOW_SHOT_RANK_COLUMN_PREFIX = "__rank_";
export enum UnitTypes {
  Time,
  Distance,
  None,
  Disabled,
}
export interface AccuracyData {
  joinPair: JoinPair<string>;
  absolute: boolean;
  accuracy: number;
  unitType: UnitTypes;
  unit: string;
}
export interface JoinPair<T> {
  first: T;
  second: T;
}
// LowShotLabels enum for labeling data in a binary classification
export enum LowShotLabels {
  positive = "positive",
  negative = "negative",
  unlabeled = "unlabeled",
}

// DatasetUpdate is an interface that contains the data to update existing data
export interface DatasetUpdate {
  index: string; // d3mIndex
  name: string; // storageName
  value: string; // new value to replace old value
}

export interface TimeIntervals {
  value: number;
  text: string;
}

// update datasets in local storage
export function addRecentDataset(dataset: string) {
  const datasets = localStorage.get("recent-datasets") || [];
  if (datasets.indexOf(dataset) === -1) {
    datasets.unshift(dataset);
    localStorage.set("recent-datasets", datasets);
  }
}
const findBestRating = (
  mainString: string,
  targetStrings: string[]
): number => {
  return findBestMatch(mainString, targetStrings)?.bestMatch.rating ?? 0;
};

// Find which labels is most suited to be the positive one
export function findAPositiveLabel(labels: string[]): string {
  // List of positives and negatives words that could be used in labels
  const positives = ["true", "positive", "aff", "1", "yes", "good", "high"];
  const negatives = ["false", "negative", "not", "0", "no", "bad", "low"];
  const ratings = labels.map((label) => {
    return {
      positive: findBestRating(label, positives),
      negative: findBestRating(label, negatives),
    };
  });

  // Default to the first label
  let positiveLabel = labels[0];

  // Select the second label, if the first label...
  if (
    // has a lower or identical positive rating and
    ratings[0].positive <= ratings[1].positive &&
    // has a higher negative rating
    ratings[0].negative > ratings[1].negative
  ) {
    positiveLabel = labels[1];
  }

  return positiveLabel;
}
// include the highlight
export function getAllDataItems(includedActive: boolean): TableRow[] {
  const tableData = includedActive
    ? datasetGetters.getBaselineIncludeTableDataItems(store)
    : datasetGetters.getBaselineExcludeTableDataItems(store);
  const highlighted = tableData
    ? tableData.map((h) => {
        return { ...h, isExcluded: true }; // adding isExcluded for the geoplot to color it gray
      })
    : [];
  return includedActive
    ? [...highlighted, ...datasetGetters.getIncludedTableDataItems(store)]
    : [...highlighted, ...datasetGetters.getExcludedTableDataItems(store)];
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
  return getTimeseriesVariablesFromFields(variables, fields).map(
    (v) => v.grouping as TimeseriesGrouping
  );
}

export function getTimeseriesVariablesFromFields(
  variables: Variable[],
  fields: Dictionary<TableColumn>
): Variable[] {
  // Check to see if any of the fields are the ID column of one of our variables
  const fieldKeys = _.map(fields, (_, key) => key);
  return variables.filter(
    (v) =>
      v.grouping &&
      v.grouping.idCol &&
      v.colType === TIMESERIES_TYPE &&
      _.includes(fieldKeys, v.key)
  );
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

export async function fetchSummaryExemplars(
  datasetName: string,
  variableName: string,
  summary: VariableSummary
) {
  const variables = datasetGetters.getVariables(store);
  const variable = variables.find((v) => v.key === variableName);

  const baselineExemplars = summary.baseline?.exemplars;
  const filteredExemplars = summary.filtered?.exemplars;
  const exemplars = filteredExemplars ? filteredExemplars : baselineExemplars;

  if (exemplars) {
    if (variable.grouping) {
      if (variable.grouping.type === TIMESERIES_TYPE) {
        // if there a linked exemplars, fetch those before resolving
        const solutionId = routeGetters.getRouteSolutionId(store);
        const grouping = variable.grouping as TimeseriesGrouping;
        const args = {
          dataset: datasetName,
          variableKey: variable.key,
          xColName: grouping.xCol,
          yColName: grouping.yCol,
          timeseriesIds: exemplars,
          solutionId: solutionId,
        };
        if (solutionId) {
          return await resultsActions.fetchForecastedTimeseries(store, args);
        } else {
          return await datasetActions.fetchTimeseries(store, args);
        }
      }
    } else {
      // if there are linked files, fetch some of them before resolving
      return await datasetActions.fetchFiles(store, {
        dataset: datasetName,
        variable: variableName,
        urls: exemplars.slice(0, 5),
      });
    }
  }
}

export async function fetchResultExemplars(
  datasetName: string,
  variableName: string,
  key: string,
  solutionId: string,
  summary: VariableSummary
) {
  const variables = datasetGetters.getVariables(store);
  const variable = variables.find((v) => v.key === variableName);

  const baselineExemplars = summary?.baseline?.exemplars;
  const filteredExemplars = summary?.filtered?.exemplars;
  const exemplars = filteredExemplars ? filteredExemplars : baselineExemplars;

  if (exemplars) {
    if (variable.grouping) {
      if (variable.grouping.type === TIMESERIES_TYPE) {
        const grouping = variable.grouping as TimeseriesGrouping;
        // if there a linked exemplars, fetch those before resolving
        return await resultsActions.fetchForecastedTimeseries(store, {
          dataset: datasetName,
          variableKey: variable.key,
          xColName: grouping.xCol,
          yColName: grouping.yCol,
          timeseriesIds: exemplars,
          solutionId: solutionId,
        });
      }
    } else {
      // if there a linked files, fetch those before resolving
      return await datasetActions.fetchFiles(store, {
        dataset: datasetName,
        variable: variableName,
        urls: exemplars,
      });
    }
  }
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
    const solutionIdAvailable =
      !s.solutionId || s.solutionId === summary.solutionId;
    return (
      s.dataset === summary.dataset &&
      s.key === summary.key &&
      solutionIdAvailable
    );
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
export function bulkUpdateSummaries(
  summaries: VariableSummary[],
  variableSummaryDictionary: Dictionary<Dictionary<VariableSummary>>
): Dictionary<Dictionary<VariableSummary>> {
  const routeKey = minimumRouteKey();
  const clone = _.cloneDeep(variableSummaryDictionary);
  summaries.forEach((summary) => {
    const summaryKey = summary.key;
    const dataset = summary.dataset;
    const compositeKey = VariableSummaryKey(summaryKey, dataset);
    if (!clone[compositeKey]) {
      clone[compositeKey] = {};
    }
    clone[compositeKey][routeKey] = Object.freeze(summary);
  });
  return clone;
}
export function updateSummariesPerVariable(
  summary: VariableSummary,
  variableSummaryDictionary: Dictionary<Dictionary<VariableSummary>>
) {
  const routeKey = minimumRouteKey();
  const summaryKey = summary.key;
  const dataset = summary.dataset;
  const compositeKey = VariableSummaryKey(summaryKey, dataset);
  // check for existing summaries for that variable, if not, instantiate
  if (!variableSummaryDictionary[compositeKey]) {
    Vue.set(variableSummaryDictionary, compositeKey, {});
  }
  // freezing the return to prevent slow, unnecessary deep reactivity.
  Vue.set(
    variableSummaryDictionary[compositeKey],
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
        return (item[tsg.idCol]?.value as string) + (uniqueTrail ?? "");
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
    (v) => v.distilRole === DISTIL_ROLES.Grouping && v.grouping !== null
  );
  const hiddenFlat = [].concat.apply(
    [],
    groupingVars.map((v) =>
      [].concat(v.grouping.hidden).concat(v.grouping.subIds)
    )
  );
  const hidden = new Map(hiddenFlat.map((v) => [v, v]));

  // the groupings that hide variables are themselves variables to display
  const groupingDisplayed = new Map(groupingVars.map((v) => [v.key, v]));

  return variables.filter(
    (v) =>
      (v.distilRole === "data" && !hidden.has(v.key)) ||
      groupingDisplayed.has(v.key)
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

export function sameData(old: TableRow[], cur: TableRow[]): boolean {
  if (old === null || cur === null) {
    return false;
  }
  if (old.length !== cur.length) {
    return false;
  }
  const oldNumOfProps = old[0] ? Object.keys(old[0]).length : 0;
  const curNumOfProps = cur[0] ? Object.keys(cur[0]).length : 0;
  if (oldNumOfProps != curNumOfProps) {
    return false;
  }
  for (let i = 0; i < old.length; ++i) {
    if (old[i][D3M_INDEX_FIELD] !== cur[i][D3M_INDEX_FIELD]) {
      return false;
    }
  }
  return true;
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
  filterParams: FilterParams | FilterSetsParams,
  varMode: SummaryMode,
  handleMutations: boolean
): Promise<void | VariableSummaryResp<ResultsContext>> {
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
      filterParams
        ? filterParams
        : { highlights: { invert: false }, filters: { invert: false } }
    );
    // save the histogram data if this is summary data
    const summary = response.data[resultProperty] as VariableSummary;
    if (!summary) {
      return;
    }
    await fetchResultExemplars(
      dataset,
      target,
      resultProperty,
      solutionId,
      summary
    );
    summary.solutionId = solutionId;
    summary.dataset = dataset;
    if (handleMutations) {
      updateFunction(context, summary);
      return;
    }
    return { context, summary };
  } catch (error) {
    console.error(error);
    if (handleMutations) {
      updateFunction(context, createErrorSummary(key, label, dataset, error));
      return;
    }
    return { context, summary: createErrorSummary(key, label, dataset, error) };
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
  filterParams: FilterParams | FilterSetsParams,
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
    const summary = response.data.summary as VariableSummary;
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
      (v && v.key.toLowerCase().includes(searchQuery.toLowerCase()))
    );
  });
}
export function totalAreaCoverage(
  data: TableRow[],
  variables: Variable[]
): number {
  if (!data || !data.length || !variables || !variables.length) {
    return 0;
  }
  const coordinateColumns = variables
    .filter((v) => v.colType === GEOBOUNDS_TYPE)
    .map((v) => (v.grouping as GeoBoundsGrouping).coordinatesCol);
  if (!coordinateColumns.length) {
    return 0;
  }
  const coordinateColumn = coordinateColumns[0];
  const coordinates = data[0][coordinateColumn]?.value;
  if (!coordinates || coordinates.some((x) => x === undefined)) {
    return 0;
  }
  const quad = polygon([
    [
      [coordinates[7], coordinates[6]],
      [coordinates[1], coordinates[0]],
      [coordinates[3], coordinates[2]],
      [coordinates[5], coordinates[4]],
      [coordinates[7], coordinates[6]],
    ],
  ]);
  // meters to km *limits to 2 decimal place*
  return Math.round(area(quad) * data.length * 0.0001 + Number.EPSILON) / 100;
}
export function topVariablesNames(variables: Variable[], max = 5): string[] {
  return sortVariablesByPCARanking(filterVariablesByFeature(variables))
    .slice(0, max)
    .map((variable) => variable.colDisplayName);
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
  const isImageOrGeoType = (v: Variable) => {
    return isImageType(v.colType) || isGeoLocatedType(v.colType);
  };
  const isNotImageOrGeoType = (v: Variable) => {
    return !isImageType(v.colType) && !isGeoLocatedType(v.colType);
  };
  // Fallback to PCA if none of the above is available
  if (!rankMap) {
    variables.sort((a, b) => {
      return b.importance - a.importance;
    });
    return [
      ...variables.filter(isImageOrGeoType),
      ...variables.filter(isNotImageOrGeoType),
    ];
  }
  variables.sort((a, b) => {
    return rankMap[b.key] - rankMap[a.key];
  });
  // give image and coordinate data priority
  return [
    ...variables.filter(isImageOrGeoType),
    ...variables.filter(isNotImageOrGeoType),
  ];
}
export function multibandURLtoInfo(url: string) {
  let split = url.split("_");
  const id = split[0];
  split = split[1].split("T");
  const time = split[1];
  const date = split[0];
  return { id, time, date };
}
// remove variable from training
export async function removeVariableFromTraining(
  group: Group,
  router: VueRouter
) {
  const dataset = routeGetters.getRouteDataset(store);
  const targetName = routeGetters.getRouteTargetVariable(store);
  const isCategorical: boolean = group.type === "categorical";
  const isTimeseries = routeGetters.isTimeseries(store);
  // get an updated view of the training data list
  const training = routeGetters.getDecodedTrainingVariableNames(store);
  training.splice(training.indexOf(group.key), 1);

  // update task based on the current training data
  const taskResponse = await datasetActions.fetchTask(store, {
    dataset,
    targetName,
    variableNames: training,
  });

  // update route with training data
  const entry = overlayRouteEntry(routeGetters.getRoute(store), {
    training: training.join(","),
    task: taskResponse.data.task.join(","),
  });

  if (isTimeseries && isCategorical) {
    // Fetch the information of the timeseries grouping
    const currentGrouping = datasetGetters
      .getGroupings(store)
      .find((v) => v.key === targetName)?.grouping;

    // Simply duplicate its grouping information and remove the series ID
    const grouping = JSON.parse(JSON.stringify(currentGrouping));
    grouping.subIds = grouping.subIds.filter((subId) => subId !== group.key);
    grouping.idCol = getComposedVariableKey(grouping.subIds);

    // Request to update the timeseries grouping without this series ID
    await datasetActions.updateGrouping(store, {
      variable: targetName,
      grouping,
    });
  }

  router.push(entry).catch((err) => console.warn(err));
  removeFiltersByName(router, group.key);
}
// add variable to training data
export async function addVariableToTraining(group: Group, router: VueRouter) {
  const dataset = routeGetters.getRouteDataset(store);
  const targetName = routeGetters.getRouteTargetVariable(store);
  const isTimeseries = routeGetters.isTimeseries(store);
  const isCategorical: boolean = group.type === "categorical";
  // get an updated view of the training data list
  const training = routeGetters
    .getDecodedTrainingVariableNames(store)
    .concat([group.key]);

  // update task based on the current training data
  const taskResponse = await datasetActions.fetchTask(store, {
    dataset,
    targetName,
    variableNames: training,
  });
  const task = taskResponse.data.task.join(",");
  // update route with training data
  let entry = overlayRouteEntry(routeGetters.getRoute(store), {
    training: training.join(","),
    task: task,
  });

  if (isTimeseries && isCategorical) {
    // Fetch the information of the timeseries grouping
    const currentGrouping = datasetGetters
      .getGroupings(store)
      .find((v) => v.key === targetName)?.grouping;

    // Simply duplicate its grouping information and add the new variable
    const grouping = JSON.parse(JSON.stringify(currentGrouping));
    grouping.subIds.push(group.key);
    grouping.idCol = getComposedVariableKey(grouping.subIds);

    // Request to update the timeserie grouping
    await datasetActions.updateGrouping(store, {
      variable: targetName,
      grouping,
    });
  }
  if (task.includes(TaskTypes.REMOTE_SENSING)) {
    const available = routeGetters.getAvailableVariables(store);
    const varModesMap = routeGetters.getDecodedVarModes(store);
    training.forEach((v) => {
      varModesMap.set(v, SummaryMode.MultiBandImage);
    });

    available.forEach((v) => {
      varModesMap.set(v.key, SummaryMode.MultiBandImage);
    });

    varModesMap.set(
      routeGetters.getRouteTargetVariable(store),
      SummaryMode.MultiBandImage
    );
    const varModesStr = varModesToString(varModesMap);
    entry = overlayRouteEntry(routeGetters.getRoute(store), {
      training: training.join(","),
      task: task,
      varModes: varModesStr,
    });
  }
  router.push(entry).catch((err) => console.warn(err));
}
export function getAllVariablesSummaries(
  variables: Variable[],
  summaryDictionary: Dictionary<Dictionary<VariableSummary>>,
  dataset?: string
): VariableSummary[] {
  return getVariableSummariesByState(
    0,
    variables.length,
    variables,
    summaryDictionary,
    false,
    dataset
  );
}
export function getVariableSummariesByState(
  pageIndex: number,
  pageSize: number,
  variables: Variable[],
  summaryDictionary: Dictionary<Dictionary<VariableSummary>>,
  isSorted = false,
  dataset = ""
): VariableSummary[] {
  const routeKey = minimumRouteKey();
  const ranked =
    routeGetters.getRouteIsTrainingVariablesRanked(store) || isSorted;
  const ds = dataset.length ? dataset : routeGetters.getRouteDataset(store);
  if (!(Object.keys(summaryDictionary).length > 0 && variables.length > 0)) {
    return [];
  }

  // remove any pattern cluster variables
  let sortedVariables = variables.filter((sv) => {
    return sv.key.indexOf(CLUSTER_PREFIX) < 0;
  });

  if (ranked) {
    // prioritize FI over MI
    sortedVariables = sortVariablesByImportance(sortedVariables);
  }

  // select only the current variables on the page
  sortedVariables = filterArrayByPage(pageIndex, pageSize, sortedVariables);

  // map them back to the variable summary dictionary for the current route key
  const currentSummaries = sortedVariables.reduce((cs, vn) => {
    if (!summaryDictionary[vn.key + ds]) {
      const placeholder = createPendingSummary(
        vn.key,
        vn.colDisplayName,
        vn.colDescription,
        vn.datasetName
      );
      cs.push(placeholder);
    } else {
      if (summaryDictionary[vn.key + ds][routeKey]) {
        cs.push(summaryDictionary[vn.key + ds][routeKey]);
      } else {
        const tempVariableSummaryKey = Object.keys(
          summaryDictionary[vn.key + ds]
        )[0];
        cs.push(summaryDictionary[vn.key + ds][tempVariableSummaryKey]);
      }
    }
    return cs;
  }, []);

  return currentSummaries;
}

export function getVariableImportance(v: Variable): number {
  const solutionID = routeGetters.getRouteSolutionId(store);
  const map = resultsGetters.getFeatureImportanceRanking(store)[solutionID];
  return map[v.key];
}

export function getVariableRanking(v: Variable): number {
  const datasetName = routeGetters.getRouteDataset(store);
  const map = datasetGetters.getVariableRankings(store)[datasetName]; // get MI ranking map
  if (!map) {
    return v.importance; // if MI ranking does not exist default to PCA
  }
  return map[v.key];
}

export function getSolutionFeatureImportance(
  v: Variable,
  solutionID: string
): number {
  const solutionRanks = resultsGetters.getFeatureImportanceRanking(store)[
    solutionID
  ];
  if (solutionRanks) {
    return solutionRanks[v.key];
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
export function clearAreaOfInterest() {
  // select view store
  datasetMutations.clearAreaOfInterestIncludeInner(store);
  datasetMutations.clearAreaOfInterestIncludeOuter(store);
  datasetMutations.clearAreaOfInterestExcludeInner(store);
  datasetMutations.clearAreaOfInterestExcludeOuter(store);
  // result view store
  resultsMutations.clearAreaOfInterestInner(store);
  resultsMutations.clearAreaOfInterestOuter(store);
  // prediction view store
  predictionsMutations.clearAreaOfInterestInner(store);
  predictionsMutations.clearAreaOfInterestOuter(store);
}
export function updateTableDataItems(
  data: TableData,
  newVals: Map<number, unknown>
) {
  const colTypeMap = new Map(
    data?.columns.map((val, idx) => {
      return [val.key, idx];
    })
  );
  const d3mIdx = colTypeMap.get(D3M_INDEX_FIELD);
  if (d3mIdx === undefined) {
    console.error("Error updating table data items");
    return;
  }
  data.values.forEach((resultRow) => {
    const rowD3mIdx = resultRow[d3mIdx].value;
    if (newVals.has(rowD3mIdx.toString())) {
      const val = newVals.get(rowD3mIdx.toString());
      Object.keys(val).forEach((key) => {
        const idx = colTypeMap.get(key);
        Vue.set(resultRow, idx, { value: val[key] });
      });
    }
  });
}
export function addOrderBy(orderByName: string) {
  if (routeGetters.getOrderBy(store) == orderByName) {
    return;
  }
  const entry = overlayRouteEntry(routeGetters.getRoute(store), {
    orderBy: orderByName,
  });
  router.push(entry).catch((err) => console.warn(err));
}

export function fetchLowShotScores() {
  const highlights = routeGetters.getDecodedHighlights(store);
  const filterParams = _.cloneDeep(
    routeGetters.getDecodedSolutionRequestFilterParams(store)
  );
  const lowShotScore = "__query_LowShotLabel";
  const dataset = routeGetters.getRouteDataset(store);
  const dataMode = routeGetters.getDataMode(store);
  datasetActions.fetchIncludedTableData(store, {
    dataset,
    filterParams,
    highlights,
    dataMode,
    orderBy: lowShotScore,
  });
}

export function getTableDataItems(data: TableData): TableRow[] {
  if (validateData(data)) {
    // convert fetched result data rows into table data rows
    const formattedTable = data.values.map((resultRow, rowIndex) => {
      const row = {} as TableRow;
      resultRow.forEach((colValue, colIndex) => {
        const key = data.columns[colIndex].key;
        const colType = data.columns[colIndex].type;
        if (key !== "d3mIndex") {
          row[key] = {};
          row[key].value = formatValue(colValue.value, colType);
          if (colValue.weight !== null && colValue.weight !== undefined) {
            row[key].weight = colValue.weight;
          }
          if (colValue.confidence !== undefined) {
            const conKey = "confidence";
            row[conKey] = {};
            row[conKey].value = colValue.confidence;
          }
          if (colValue.rank !== undefined) {
            const conKey = "rank";
            row[conKey] = {};
            row[conKey].value = colValue.rank;
          }
        } else {
          row[key] = formatValue(colValue.value, colType);
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

    data.columns.forEach((col, ind) => {
      if (col.key !== D3M_INDEX_FIELD) {
        // Error and predicted columns require unique handling.  They use a special key of the format
        // <solution_id>:<predicted|error> and are not available in the variables list.
        let variable: Variable = null;
        let description: string = null;
        let label: string = null;
        if (isPredictedCol(col.key)) {
          label = routeGetters.getRouteTargetVariable(store);
          description = `Model predicted value for ${label}`;

          // if we actually have defined confidence values, then let's add confidence to the table
          if (data.values[0][ind]?.confidence !== null) {
            result.confidence = {
              label: "Confidence",
              key: "confidence",
              type: "numeric",
              weight: null,
              headerTitle: `Prediction confidence ${label}`,
              sortable: true,
            };
          }
        } else if (isErrorCol(col.key)) {
          variable = requestGetters.getActiveSolutionTargetVariable(store);
          label = "Error";
          description = `Difference between actual and predicted value for ${variable.key}`;
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
    });

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

  // find remote sensing image fields
  const fieldKeys = _.map(fields, (_, key) => key);
  const multiBandImageFields = datasetGetters
    .getVariables(store)
    .filter(
      (v) =>
        v.grouping &&
        v.grouping.idCol &&
        v.colType === MULTIBAND_IMAGE_TYPE &&
        _.includes(fieldKeys, v.key)
    )
    .map((v) => ({ key: v.key, type: v.colType }));

  // the two are probably mutually exclusive, but it doesn't hurt anything to allow for both
  return imageFields.concat(multiBandImageFields);
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
      valueColumns[0].key !== timeColumns[0].key) ||
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
      latColumns[0].key !== lonColumns[0].key) ||
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

export function downloadFile(
  fileContent: string,
  fileName: string,
  extension: string,
  type = "text/csv"
) {
  const data = new Blob([fileContent], { type });
  const url = URL.createObjectURL(data);
  const link = document.createElement("a");
  link.setAttribute("href", url);
  link.setAttribute("download", fileName + extension);
  link.style.display = "none";
  document.body.appendChild(link); // Required for FF

  link.click(); // This will download the data file named "my_data.csv".
  document.body.removeChild(link);
  return;
}
export function debounceFetchImagePack(args: {
  items: TableRow[];
  imageFields: Field[];
  dataset?: string;
  uniqueTrail?: string;
  debounceKey: number;
  timeout?: number;
}) {
  const timeout = args.timeout ?? 1000;
  clearTimeout(args.debounceKey);
  args.debounceKey = setTimeout(() => {
    bulkRemoveImages(args);
    FetchImagePack(args);
  }, timeout);
}
export function bulkRemoveImages(args: {
  imageFields: Field[];
  items: TableRow[];
  uniqueTrail?: string;
}) {
  if (!args.imageFields.length || !args.items.length) {
    return;
  }
  const imageKey = args.imageFields[0].key;
  if (!imageKey || !args.items[0][imageKey]) {
    return;
  }
  let imageUrlBuilder = (item: TableRow) => {
    return `${item[imageKey].value}/${args.uniqueTrail}`;
  };
  if (!args.uniqueTrail) {
    imageUrlBuilder = (item: TableRow) => {
      return `${item[imageKey].value}`;
    };
  }
  datasetMutations.bulkRemoveFiles(store, {
    urls: args.items.map(imageUrlBuilder),
  });
}
export function FetchImagePack(args: {
  items: TableRow[];
  imageFields: Field[];
  dataset?: string;
  uniqueTrail?: string;
}) {
  const band = routeGetters.getBandCombinationId(store);
  if (!args.imageFields.length || !args.items.length) {
    return;
  }
  const key = args.imageFields[0].key;
  const type = args.imageFields[0].type;
  if (!args.items[0][key]) {
    return;
  }
  let dataset = args.dataset ?? routeGetters.getRouteDataset(store);
  datasetActions.fetchImagePack(store, {
    multiBandImagePackRequest: {
      imageIds: args.items.map((item) => {
        return item[key].value as string;
      }),
      dataset,
      band: type === MULTIBAND_IMAGE_TYPE ? band : "",
      colorScale: routeGetters.getImageLayerScale(store),
    },
    uniqueTrail: args.uniqueTrail,
  });
}
