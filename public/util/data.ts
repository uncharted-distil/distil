import _ from "lodash";
import axios from "axios";
import Vue from "vue";
import {
  Variable,
  VariableSummary,
  TimeseriesSummary,
  TableData,
  TableRow,
  TableColumn,
  Grouping,
  D3M_INDEX_FIELD
} from "../store/dataset/index";
import { Solution, SOLUTION_COMPLETED } from "../store/solutions/index";
import { Dictionary } from "./dict";
import { FilterParams } from "./filters";
import store from "../store/store";
import { actions as resultsActions } from "../store/results/module";
import { ResultsContext } from "../store/results/actions";
import {
  getters as datasetGetters,
  actions as datasetActions
} from "../store/dataset/module";
import { getters as solutionGetters } from "../store/solutions/module";
import {
  formatValue,
  hasComputedVarPrefix,
  isIntegerType,
  isTimeType,
  IMAGE_TYPE,
  TIMESERIES_TYPE
} from "../util/types";

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
  return _.map(fields, (field, key) => key)
    .filter(key => {
      const v = variables.find(v => v.colName === key);
      return v && v.grouping && v.grouping.type === TIMESERIES_TYPE;
    })
    .map(key => {
      const v = variables.find(v => v.colName === key);
      return v.grouping;
    });
}

export function getComposedVariableKey(keys: string[]): string {
  return keys.join("_");
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
    values: []
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
  filterParams: FilterParams
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

  // return promise
  return axios
    .post(`${endpoint}/${resultId}`, filterParams ? filterParams : {})
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

export function filterUnsupportedTargets(
  variables: VariableSummary[]
): VariableSummary[] {
  return variables.filter(variableSummary => {
    return (
      !(variableSummary.varType && variableSummary.varType === IMAGE_TYPE) &&
      !hasComputedVarPrefix(variableSummary.key)
    );
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
        label = variable.colDisplayName;
        description = `Difference between actual and predicted value for ${variable.colName}`;
      } else {
        variable = variables[col.key];
        if (variable) {
          label = col.label;
          description = variable.colDescription;
        }
      }

      if (col.type === TIMESERIES_TYPE) {
        if (isPredictedCol(col.key)) {
          // do not display predicted col for timeseries
          continue;
        }

        if (variable && variable.grouping) {
          label = variable.grouping.properties.yCol;
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

// temp dummy data for geocoordinate to circumvent need to add to includedSets
export const DUMMY_GEODATA = [
  {
    latitude: "35.814300",
    longitude: "36.320600",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7224",
    _key: 0,
    _rowVariant: null
  },
  {
    latitude: "33.528400",
    longitude: "36.352400",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7225",
    _key: 1,
    _rowVariant: null
  },
  {
    latitude: "32.686200",
    longitude: "36.223500",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7227",
    _key: 2,
    _rowVariant: null
  },
  {
    latitude: "34.800800",
    longitude: "36.711600",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7228",
    _key: 3,
    _rowVariant: null
  },
  {
    latitude: "34.800800",
    longitude: "36.711600",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7229",
    _key: 4,
    _rowVariant: null
  },
  {
    latitude: "35.524400",
    longitude: "36.482300",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7230",
    _key: 5,
    _rowVariant: null
  },
  {
    latitude: "33.558300",
    longitude: "36.440900",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7231",
    _key: 6,
    _rowVariant: null
  },
  {
    latitude: "34.800100",
    longitude: "36.626500",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7234",
    _key: 7,
    _rowVariant: null
  },
  {
    latitude: "35.398400",
    longitude: "36.619700",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7235",
    _key: 8,
    _rowVariant: null
  },
  {
    latitude: "33.576900",
    longitude: "36.447700",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7239",
    _key: 9,
    _rowVariant: null
  },
  {
    latitude: "33.522400",
    longitude: "36.458400",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7243",
    _key: 10,
    _rowVariant: null
  },
  {
    latitude: "33.522400",
    longitude: "36.458400",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7244",
    _key: 11,
    _rowVariant: null
  },
  {
    latitude: "32.551300",
    longitude: "36.186900",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7245",
    _key: 12,
    _rowVariant: null
  },
  {
    latitude: "33.510600",
    longitude: "36.485500",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7246",
    _key: 13,
    _rowVariant: null
  },
  {
    latitude: "33.550300",
    longitude: "36.400500",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7250",
    _key: 14,
    _rowVariant: null
  },
  {
    latitude: "35.358500",
    longitude: "36.653000",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7252",
    _key: 15,
    _rowVariant: null
  },
  {
    latitude: "35.373600",
    longitude: "36.601800",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7255",
    _key: 16,
    _rowVariant: null
  },
  {
    latitude: "33.512600",
    longitude: "36.372100",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7257",
    _key: 17,
    _rowVariant: null
  },
  {
    latitude: "34.803700",
    longitude: "36.638900",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7258",
    _key: 18,
    _rowVariant: null
  },
  {
    latitude: "33.524900",
    longitude: "36.481400",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7261",
    _key: 19,
    _rowVariant: null
  },
  {
    latitude: "33.547200",
    longitude: "36.460000",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7262",
    _key: 20,
    _rowVariant: null
  },
  {
    latitude: "33.527000",
    longitude: "36.420200",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7263",
    _key: 21,
    _rowVariant: null
  },
  {
    latitude: "33.516500",
    longitude: "36.489700",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7265",
    _key: 22,
    _rowVariant: null
  },
  {
    latitude: "35.346900",
    longitude: "36.536600",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7268",
    _key: 23,
    _rowVariant: null
  },
  {
    latitude: "34.939100",
    longitude: "36.622400",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7269",
    _key: 24,
    _rowVariant: null
  },
  {
    latitude: "33.564300",
    longitude: "36.371200",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7270",
    _key: 25,
    _rowVariant: null
  },
  {
    latitude: "33.564300",
    longitude: "36.371200",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7271",
    _key: 26,
    _rowVariant: null
  },
  {
    latitude: "33.564300",
    longitude: "36.371200",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7272",
    _key: 27,
    _rowVariant: null
  },
  {
    latitude: "33.514900",
    longitude: "36.466500",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7274",
    _key: 28,
    _rowVariant: null
  },
  {
    latitude: "34.822000",
    longitude: "36.696600",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7276",
    _key: 29,
    _rowVariant: null
  },
  {
    latitude: "33.513400",
    longitude: "36.350000",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7277",
    _key: 30,
    _rowVariant: null
  },
  {
    latitude: "33.570600",
    longitude: "36.404600",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7280",
    _key: 31,
    _rowVariant: null
  },
  {
    latitude: "32.611100",
    longitude: "36.118000",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7285",
    _key: 32,
    _rowVariant: null
  },
  {
    latitude: "32.624100",
    longitude: "36.104900",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7287",
    _key: 33,
    _rowVariant: null
  },
  {
    latitude: "32.754400",
    longitude: "36.131100",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7291",
    _key: 34,
    _rowVariant: null
  },
  {
    latitude: "33.537900",
    longitude: "36.400500",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7292",
    _key: 35,
    _rowVariant: null
  },
  {
    latitude: "33.538800",
    longitude: "36.365300",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7296",
    _key: 36,
    _rowVariant: null
  },
  {
    latitude: "35.353300",
    longitude: "36.558200",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7297",
    _key: 37,
    _rowVariant: null
  },
  {
    latitude: "35.340600",
    longitude: "36.577700",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7299",
    _key: 38,
    _rowVariant: null
  },
  {
    latitude: "32.633400",
    longitude: "36.161600",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7300",
    _key: 39,
    _rowVariant: null
  },
  {
    latitude: "34.874700",
    longitude: "36.523800",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7301",
    _key: 40,
    _rowVariant: null
  },
  {
    latitude: "35.321300",
    longitude: "36.620400",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7302",
    _key: 41,
    _rowVariant: null
  },
  {
    latitude: "35.467800",
    longitude: "36.536700",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7305",
    _key: 42,
    _rowVariant: null
  },
  {
    latitude: "35.331900",
    longitude: "40.146100",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7347",
    _key: 43,
    _rowVariant: null
  },
  {
    latitude: "35.346900",
    longitude: "36.536600",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7349",
    _key: 44,
    _rowVariant: null
  },
  {
    latitude: "35.814300",
    longitude: "36.320600",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7350",
    _key: 45,
    _rowVariant: null
  },
  {
    latitude: "33.528400",
    longitude: "36.352400",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7352",
    _key: 46,
    _rowVariant: null
  },
  {
    latitude: "32.686200",
    longitude: "36.223500",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7354",
    _key: 47,
    _rowVariant: null
  },
  {
    latitude: "34.842700",
    longitude: "36.726700",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7360",
    _key: 48,
    _rowVariant: null
  },
  {
    latitude: "33.558300",
    longitude: "36.440900",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7366",
    _key: 49,
    _rowVariant: null
  },
  {
    latitude: "33.576900",
    longitude: "36.447700",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7374",
    _key: 50,
    _rowVariant: null
  },
  {
    latitude: "35.546400",
    longitude: "36.800600",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7377",
    _key: 51,
    _rowVariant: null
  },
  {
    latitude: "33.522400",
    longitude: "36.458400",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7380",
    _key: 52,
    _rowVariant: null
  },
  {
    latitude: "35.373900",
    longitude: "36.689300",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7381",
    _key: 53,
    _rowVariant: null
  },
  {
    latitude: "33.550300",
    longitude: "36.400500",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7383",
    _key: 54,
    _rowVariant: null
  },
  {
    latitude: "33.550300",
    longitude: "36.400500",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7384",
    _key: 55,
    _rowVariant: null
  },
  {
    latitude: "35.358500",
    longitude: "36.653000",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7392",
    _key: 56,
    _rowVariant: null
  },
  {
    latitude: "35.298400",
    longitude: "40.289400",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7394",
    _key: 57,
    _rowVariant: null
  },
  {
    latitude: "33.512600",
    longitude: "36.372100",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7406",
    _key: 58,
    _rowVariant: null
  },
  {
    latitude: "33.506300",
    longitude: "36.385700",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7408",
    _key: 59,
    _rowVariant: null
  },
  {
    latitude: "33.226400",
    longitude: "35.831900",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7409",
    _key: 60,
    _rowVariant: null
  },
  {
    latitude: "33.547200",
    longitude: "36.460000",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7416",
    _key: 61,
    _rowVariant: null
  },
  {
    latitude: "33.564300",
    longitude: "36.371200",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7420",
    _key: 62,
    _rowVariant: null
  },
  {
    latitude: "33.564300",
    longitude: "36.371200",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7421",
    _key: 63,
    _rowVariant: null
  },
  {
    latitude: "33.564300",
    longitude: "36.371200",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7422",
    _key: 64,
    _rowVariant: null
  },
  {
    latitude: "34.822000",
    longitude: "36.696600",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7423",
    _key: 65,
    _rowVariant: null
  },
  {
    latitude: "33.514700",
    longitude: "36.399200",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7425",
    _key: 66,
    _rowVariant: null
  },
  {
    latitude: "33.570600",
    longitude: "36.404600",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7428",
    _key: 67,
    _rowVariant: null
  },
  {
    latitude: "35.340200",
    longitude: "37.003000",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7429",
    _key: 68,
    _rowVariant: null
  },
  {
    latitude: "32.624100",
    longitude: "36.104900",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7430",
    _key: 69,
    _rowVariant: null
  },
  {
    latitude: "32.624100",
    longitude: "36.104900",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7431",
    _key: 70,
    _rowVariant: null
  },
  {
    latitude: "34.823200",
    longitude: "36.675700",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7438",
    _key: 71,
    _rowVariant: null
  },
  {
    latitude: "34.924100",
    longitude: "36.731200",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7444",
    _key: 72,
    _rowVariant: null
  },
  {
    latitude: "33.538800",
    longitude: "36.365300",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7445",
    _key: 73,
    _rowVariant: null
  },
  {
    latitude: "35.340600",
    longitude: "36.577700",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7448",
    _key: 74,
    _rowVariant: null
  },
  {
    latitude: "32.633400",
    longitude: "36.161600",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7450",
    _key: 75,
    _rowVariant: null
  },
  {
    latitude: "35.321300",
    longitude: "36.620400",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7452",
    _key: 76,
    _rowVariant: null
  },
  {
    latitude: "34.895100",
    longitude: "36.495300",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7453",
    _key: 77,
    _rowVariant: null
  },
  {
    latitude: "35.011300",
    longitude: "37.051000",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7482",
    _key: 78,
    _rowVariant: null
  },
  {
    latitude: "35.814300",
    longitude: "36.320600",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7483",
    _key: 79,
    _rowVariant: null
  },
  {
    latitude: "33.528400",
    longitude: "36.352400",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7485",
    _key: 80,
    _rowVariant: null
  },
  {
    latitude: "34.870100",
    longitude: "36.653000",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7489",
    _key: 81,
    _rowVariant: null
  },
  {
    latitude: "34.870100",
    longitude: "36.653000",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7490",
    _key: 82,
    _rowVariant: null
  },
  {
    latitude: "35.257000",
    longitude: "36.382600",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7492",
    _key: 83,
    _rowVariant: null
  },
  {
    latitude: "33.665300",
    longitude: "36.329000",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7494",
    _key: 84,
    _rowVariant: null
  },
  {
    latitude: "33.518100",
    longitude: "36.384100",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7499",
    _key: 85,
    _rowVariant: null
  },
  {
    latitude: "34.838400",
    longitude: "36.571700",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7500",
    _key: 86,
    _rowVariant: null
  },
  {
    latitude: "34.919600",
    longitude: "36.944900",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7501",
    _key: 87,
    _rowVariant: null
  },
  {
    latitude: "33.576900",
    longitude: "36.447700",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7502",
    _key: 88,
    _rowVariant: null
  },
  {
    latitude: "33.522400",
    longitude: "36.458400",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7507",
    _key: 89,
    _rowVariant: null
  },
  {
    latitude: "32.888200",
    longitude: "36.041000",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7508",
    _key: 90,
    _rowVariant: null
  },
  {
    latitude: "35.373900",
    longitude: "36.689300",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7512",
    _key: 91,
    _rowVariant: null
  },
  {
    latitude: "33.550300",
    longitude: "36.400500",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7513",
    _key: 92,
    _rowVariant: null
  },
  {
    latitude: "33.545600",
    longitude: "36.385500",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7517",
    _key: 93,
    _rowVariant: null
  },
  {
    latitude: "35.414400",
    longitude: "36.389000",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7518",
    _key: 94,
    _rowVariant: null
  },
  {
    latitude: "35.358500",
    longitude: "36.653000",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7524",
    _key: 95,
    _rowVariant: null
  },
  {
    latitude: "35.373600",
    longitude: "36.601800",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7525",
    _key: 96,
    _rowVariant: null
  },
  {
    latitude: "33.150500",
    longitude: "36.038800",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7527",
    _key: 97,
    _rowVariant: null
  },
  {
    latitude: "33.512600",
    longitude: "36.372100",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7529",
    _key: 98,
    _rowVariant: null
  },
  {
    latitude: "35.814300",
    longitude: "36.320600",
    Main_Actor: "Military Forces of Syria (2000-)",
    d3mIndex: "7530",
    _key: 99,
    _rowVariant: null
  }
];
