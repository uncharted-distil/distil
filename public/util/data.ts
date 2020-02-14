import _ from "lodash";
import axios from "axios";
import Vue from "vue";
import {
  Variable,
  VariableSummary,
  TableData,
  TableRow,
  TableColumn,
  Grouping,
  D3M_INDEX_FIELD,
  SummaryMode
} from "../store/dataset/index";
import { Solution, SOLUTION_COMPLETED } from "../store/solutions/index";
import { Dictionary } from "./dict";
import { FilterParams } from "./filters";
import store from "../store/store";
import { actions as resultsActions } from "../store/results/module";
import { ResultsContext } from "../store/results/actions";
import { PredictionContext } from "../store/predictions/actions";
import {
  getters as datasetGetters,
  actions as datasetActions
} from "../store/dataset/module";
import { getters as solutionGetters } from "../store/solutions/module";
import { formatValue, isIntegerType, isTimeType } from "../util/types";

// Postfixes for special variable names
export const PREDICTED_SUFFIX = "_predicted";
export const ERROR_SUFFIX = "_error";

export const NUM_PER_PAGE = 10;

export const DATAMART_PROVENANCE_NYU = "NYU";
export const DATAMART_PROVENANCE_ISI = "ISI";
export const ELASTIC_PROVENANCE = "elastic";
export const FILE_PROVENANCE = "file";

export const IMPORTANT_VARIABLE_RANKING_THRESHOLD = 0.5;

export function getTimeseriesSummaryTopCategories(
  summary: VariableSummary
): string[] {
  return _.map(summary.baseline.categoryBuckets, (buckets, category) => {
    return {
      category: category,
      count: _.sumBy(buckets, b => b.count)
    };
  })
    .sort((a, b) => b.count - a.count)
    .map(c => c.category);
}

export function getTimeseriesGroupingsFromFields(
  variables: Variable[],
  fields: Dictionary<TableColumn>
): Grouping[] {
  // Check to see if any of the fields are the ID column of one of our variables
  const fieldKeys = _.map(fields, (_, key) => key);
  return variables
    .filter(
      v =>
        v.grouping &&
        v.grouping.idCol &&
        _.includes(fieldKeys, v.grouping.idCol)
    )
    .map(v => v.grouping);
}

export function getComposedVariableKey(keys: string[]): string {
  return "__grouping_" + keys.join("_");
}

export function getTimeseriesAnalysisIntervals(
  timeVar: Variable,
  range: number
): any[] {
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
        { value: HOURS_VALUE, text: HOURS_LABEL }
      ];
    } else if (range < 2 * WEEKS_VALUE) {
      return [
        { value: HOURS_VALUE, text: HOURS_LABEL },
        { value: DAYS_VALUE, text: DAYS_LABEL },
        { value: WEEKS_VALUE, text: WEEKS_LABEL }
      ];
    } else if (range < MONTHS_VALUE) {
      return [
        { value: HOURS_VALUE, text: HOURS_LABEL },
        { value: DAYS_VALUE, text: DAYS_LABEL },
        { value: WEEKS_VALUE, text: WEEKS_LABEL }
      ];
    } else if (range < 4 * MONTHS_VALUE) {
      return [
        { value: DAYS_VALUE, text: DAYS_LABEL },
        { value: WEEKS_VALUE, text: WEEKS_LABEL },
        { value: MONTHS_VALUE, text: MONTHS_LABEL }
      ];
    } else if (range < YEARS_VALUE) {
      return [
        { value: WEEKS_VALUE, text: WEEKS_LABEL },
        { value: MONTHS_VALUE, text: MONTHS_LABEL }
      ];
    } else {
      return [
        { value: MONTHS_VALUE, text: MONTHS_LABEL },
        { value: YEARS_VALUE, text: YEARS_LABEL }
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
    { value: large, text: `${large}` }
  ];
}

export function fetchSummaryExemplars(
  datasetName: string,
  variableName: string,
  summary: VariableSummary
) {
  const variables = datasetGetters.getVariables(store);
  const variable = variables.find(v => v.colName === variableName);

  const baselineExemplars = summary.baseline.exemplars;
  const filteredExemplars =
    summary.filtered && summary.filtered.exemplars
      ? summary.filtered.exemplars
      : null;
  const exemplars = filteredExemplars ? filteredExemplars : baselineExemplars;

  if (exemplars) {
    if (variable.grouping) {
      if (variable.grouping.type === "timeseries") {
        // if there a linked exemplars, fetch those before resolving
        return Promise.all(
          exemplars.map(exemplar => {
            return datasetActions.fetchTimeseries(store, {
              dataset: datasetName,
              timeseriesColName: variable.grouping.idCol,
              xColName: variable.grouping.properties.xCol,
              yColName: variable.grouping.properties.yCol,
              timeseriesID: exemplar
            });
          })
        );
      }
    } else {
      // if there a linked files, fetch those before resolving
      return datasetActions.fetchFiles(store, {
        dataset: datasetName,
        variable: variableName,
        urls: exemplars
      });
    }
  }

  return new Promise(res => res());
}

export function fetchResultExemplars(
  datasetName: string,
  variableName: string,
  key: string,
  solutionId: string,
  summary: VariableSummary
) {
  const variables = datasetGetters.getVariables(store);
  const variable = variables.find(v => v.colName === variableName);

  const baselineExemplars = summary.baseline.exemplars;
  const filteredExemplars =
    summary.filtered && summary.filtered.exemplars
      ? summary.filtered.exemplars
      : null;
  const exemplars = filteredExemplars ? filteredExemplars : baselineExemplars;

  if (exemplars) {
    if (variable.grouping) {
      if (variable.grouping.type === "timeseries") {
        // if there a linked exemplars, fetch those before resolving
        return Promise.all(
          exemplars.map(exemplar => {
            return resultsActions.fetchForecastedTimeseries(store, {
              dataset: datasetName,
              timeseriesColName: variable.grouping.idCol,
              xColName: variable.grouping.properties.xCol,
              yColName: variable.grouping.properties.yCol,
              timeseriesID: exemplar,
              solutionId: solutionId
            });
          })
        );
      }
    } else {
      // if there a linked files, fetch those before resolving
      return datasetActions.fetchFiles(store, {
        dataset: datasetName,
        variable: variableName,
        urls: exemplars
      });
    }
  }

  return new Promise(res => res());
}

export function updateSummaries(
  summary: VariableSummary,
  summaries: VariableSummary[]
) {
  const index = _.findIndex(summaries, s => {
    return s.dataset === summary.dataset && s.key === summary.key;
  });
  if (index >= 0) {
    Vue.set(summaries, index, summary);
  } else {
    summaries.push(summary);
  }
}

export function removeSummary(
  summary: VariableSummary,
  summaries: VariableSummary[]
) {
  const index = _.findIndex(summaries, s => {
    return s.dataset === summary.dataset && s.key === summary.key;
  });
  if (index >= 0) {
    Vue.delete(summaries, index);
  }
}

export function filterSummariesByDataset(
  summaries: VariableSummary[],
  dataset: string
): VariableSummary[] {
  return summaries.filter(summary => {
    return summary.dataset === dataset;
  });
}

export function createEmptyTableData(): TableData {
  return {
    numRows: 0,
    columns: [],
    values: [],
    fittedSolutionId: null,
    produceRequestId: null
  };
}

export function formatSlot(key: string, slotType: string): string {
  return `${slotType}(${key})`;
}

export function formatFieldsAsArray(
  fields: Dictionary<TableColumn>
): TableColumn[] {
  return _.map(fields, field => field);
}

export function createPendingSummary(
  key: string,
  label: string,
  description: string,
  dataset: string,
  solutionId?: string
): VariableSummary {
  return {
    key: key,
    label: label,
    description: description,
    dataset: dataset,
    pending: true,
    baseline: null,
    filtered: null,
    solutionId: solutionId
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
    err: error.response ? error.response.data : error
  };
}

export function fetchSolutionResultSummary(
  context: ResultsContext,
  endpoint: string,
  solution: Solution,
  target: string,
  key: string,
  label: string,
  resultSummaries: VariableSummary[],
  updateFunction: (arg: ResultsContext, summary: VariableSummary) => void,
  filterParams: FilterParams,
  varMode: SummaryMode
): Promise<any> {
  const dataset = solution.dataset;
  const solutionId = solution.solutionId;
  const resultId = solution.resultId;

  const exists = _.find(
    resultSummaries,
    v => v.dataset === dataset && v.key === key
  );
  if (!exists) {
    // add placeholder
    updateFunction(
      context,
      createPendingSummary(key, label, dataset, solutionId)
    );
  }

  // fetch the results for each solution
  if (solution.progress !== SOLUTION_COMPLETED) {
    // skip
    return;
  }
  // finish building endpoint
  const completeEndpoint = varMode
    ? `${endpoint}/${resultId}/${varMode}`
    : `${endpoint}/${resultId}`;

  // return promise
  return axios
    .post(completeEndpoint, filterParams ? filterParams : {})
    .then(response => {
      // save the histogram data
      const summary = response.data.summary;
      return fetchResultExemplars(
        dataset,
        target,
        key,
        solutionId,
        summary
      ).then(() => {
        summary.solutionId = solutionId;
        summary.dataset = dataset;
        updateFunction(context, summary);
      });
    })
    .catch(error => {
      console.error(error);
      updateFunction(context, createErrorSummary(key, label, dataset, error));
    });
}

export function fetchPredictionResultSummary(
  context: PredictionContext,
  endpoint: string,
  solution: Solution,
  target: string,
  key: string,
  label: string,
  resultSummaries: VariableSummary[],
  updateFunction: (arg: PredictionContext, summary: VariableSummary) => void,
  filterParams: FilterParams,
  varMode: SummaryMode
): Promise<any> {
  const dataset = solution.dataset;
  const solutionId = solution.solutionId;

  const exists = _.find(
    resultSummaries,
    v => v.dataset === dataset && v.key === key
  );
  if (!exists) {
    // add placeholder
    updateFunction(
      context,
      createPendingSummary(key, label, dataset, solutionId)
    );
  }

  // fetch the results for each solution
  if (solution.progress !== SOLUTION_COMPLETED) {
    // skip
    return;
  }

  // return promise
  return axios
    .post(
      `${endpoint}/${solution.resultId}/${varMode}`,
      filterParams ? filterParams : {}
    )
    .then(response => {
      // save the histogram data
      const summary = response.data.summary;
      return fetchResultExemplars(
        dataset,
        target,
        key,
        solutionId,
        summary
      ).then(() => {
        summary.solutionId = solutionId;
        summary.dataset = dataset;
        updateFunction(context, summary);
      });
    })
    .catch(error => {
      console.error(error);
      updateFunction(context, createErrorSummary(key, label, dataset, error));
    });
}

export function filterVariablesByPage(
  pageIndex: number,
  numPerPage: number,
  variables: VariableSummary[]
): VariableSummary[] {
  if (variables.length > numPerPage) {
    const firstIndex = numPerPage * (pageIndex - 1);
    const lastIndex = Math.min(firstIndex + numPerPage, variables.length);
    return variables.slice(firstIndex, lastIndex);
  }
  return variables;
}

export function getVariableImportance(v: Variable): number {
  return v.ranking !== undefined ? v.ranking : v.importance;
}

export function getVariableRanking(v: Variable): number {
  return v.ranking !== undefined ? v.ranking : 0;
}

export function sortVariablesByImportance(variables: Variable[]): Variable[] {
  variables.sort((a, b) => {
    return getVariableImportance(b) - getVariableImportance(a);
  });
  return variables;
}

export function sortSummariesByImportance(
  summaries: VariableSummary[],
  variables: Variable[]
): VariableSummary[] {
  // create importance lookup map
  const importance: Dictionary<number> = {};
  variables.forEach(variable => {
    importance[variable.colName] = getVariableImportance(variable);
  });
  // sort by importance
  summaries.sort((a, b) => {
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
    return data.values.map((resultRow, rowIndex) => {
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
        } else {
          row[colName] = formatValue(colValue.value, colType);
        }
      });
      row._key = rowIndex;
      return row;
    });
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
        variable = solutionGetters.getActiveSolutionTargetVariable(store)[0]; // always a single value
        label = variable.colDisplayName;
        description = `Model predicted value for ${variable.colName}`;
      } else if (isErrorCol(col.key)) {
        variable = solutionGetters.getActiveSolutionTargetVariable(store)[0];
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
        headerTitle: description ? label.concat(": ", description) : label,
        sortable: true
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
