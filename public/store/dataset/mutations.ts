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
import Vue from "vue";
import {
  isDatamartProvenance,
  updateSummariesPerVariable,
  updateTableDataItems,
} from "../../util/data";
import { Dictionary } from "../../util/dict";
import { getSelectedRows } from "../../util/row";
import {
  GEOCOORDINATE_TYPE,
  LATITUDE_TYPE,
  LONGITUDE_TYPE,
} from "../../util/types";
import {
  BandCombination,
  Dataset,
  DatasetPendingRequest,
  DatasetState,
  defaultState,
  Metric,
  TableData,
  Task,
  TimeSeriesValue,
  Variable,
  VariableSummary,
} from "./index";

function sortDatasets(a: Dataset, b: Dataset) {
  if (
    isDatamartProvenance(a.provenance) &&
    !isDatamartProvenance(b.provenance)
  ) {
    return 1;
  }
  if (
    isDatamartProvenance(b.provenance) &&
    !isDatamartProvenance(a.provenance)
  ) {
    return -1;
  }
  const aID = a.id.toUpperCase();
  const bID = b.id.toUpperCase();
  if (aID < bID) {
    return -1;
  }
  if (aID > bID) {
    return 1;
  }

  return 0;
}

export interface TimeSeriesUpdate {
  variableKey: string;
  seriesID: string;
  uniqueTrail: string;
  timeseries: TimeSeriesValue[];
  isDateTime: boolean;
  min: number;
  max: number;
  mean: number;
}

export interface TimeSeriesForecastUpdate extends TimeSeriesUpdate {
  forecast: TimeSeriesValue[];
  forecastTestRange: number[];
}

export const mutations = {
  setDataset(state: DatasetState, dataset: Dataset) {
    const index = _.findIndex(state.datasets, (d) => {
      return d.id === dataset.id;
    });
    if (index === -1) {
      state.datasets.push(dataset);
    } else {
      Vue.set(state.datasets, index, dataset);
    }
    state.datasets.sort(sortDatasets);
  },

  setDatasets(state: DatasetState, datasets: Dataset[]) {
    if (!datasets) {
      return;
    }
    // individually add datasets if they do not exist
    const lookup = {};
    state.datasets.forEach((d, index) => {
      lookup[d.id] = index;
    });
    datasets.forEach((d) => {
      const index = lookup[d.id];
      if (index !== undefined) {
        // update if it already exists
        Vue.set(state.datasets, index, d);
      } else {
        // push if not
        state.datasets.push(d);
      }
    });
    state.datasets.sort(sortDatasets);

    // replace all filtered datasets
    state.filteredDatasets = datasets;
    state.filteredDatasets.sort(sortDatasets);
  },

  setVariables(state: DatasetState, variables: Variable[]) {
    const oldVariables = new Map<string, Variable>();
    // the dataset to update
    const dataset = variables.length ? variables[0].datasetName : "";
    state.variables.forEach((variable) => {
      const { datasetName, key } = variable;
      oldVariables.set(`${datasetName}:${key}`, variable);
    });

    const newVariables = variables.map((variable) => {
      const { datasetName, key } = variable;
      const variableKey = `${datasetName}:${key}`;
      const oldVariable = oldVariables.get(variableKey);

      if (oldVariable) {
        // keep previous column type reviewed state
        variable.isColTypeReviewed = oldVariable.isColTypeReviewed;
        // keep previous variable rankings
        variable.ranking = oldVariable.ranking;
      }
      return variable;
    });
    // there are instances where we have variables from multiple datasets
    // the join view it is important to keep those variables
    const otherDatasetVariables = Array.from(oldVariables.values()).filter(
      (v) => v.datasetName !== dataset
    );
    state.variables = newVariables.concat(otherDatasetVariables);
  },

  updateVariableType(
    state: DatasetState,
    args: {
      dataset: string;
      field: string;
      type: string;
      variables: Variable[];
    }
  ) {
    // TODO: fix this, this is hacky and error prone manually changing the
    // type across the app state.
    // Ideally we have it only in one state, or instead refresh all the
    // relevant store data.

    // geocoordinate temporary logic
    if (args.type === GEOCOORDINATE_TYPE) {
      Vue.set(state, "isGeocoordinateFacet", [LONGITUDE_TYPE, LATITUDE_TYPE]);
    }

    // update dataset variables & active variables
    const dataset = state.datasets.find((d) => d.name === args.dataset);
    if (dataset) {
      const target = dataset.variables.findIndex((v) => v.key === args.field);
      const stateTarget = state.variables.findIndex(
        (v) => v.key === args.field && v.datasetName === args.dataset
      );
      if (target > -1) {
        Vue.set(dataset.variables[target], "colType", args.type);
        Vue.set(state.variables[stateTarget], "colType", args.type);
        Vue.set(
          state.variables[stateTarget],
          "values",
          args.variables[target].values
        );
      }
    } else {
      const datasetExists = state.variables.some((v) => {
        return v.datasetName === args.dataset;
      });
      if (datasetExists) {
        const idx = state.variables.findIndex((v) => {
          return v.datasetName === args.dataset && v.colName === args.field;
        });
        if (idx > -1) {
          Vue.set(state.variables[idx], "colType", args.type);
        }
      }
    }

    // update table data
    if (state.includedSet.tableData) {
      const col = state.includedSet.tableData.columns.find(
        (c) => c.key === args.field
      );
      if (col) {
        col.type = args.type;
      }
    }
    if (state.excludedSet.tableData) {
      const col = state.excludedSet.tableData.columns.find(
        (c) => c.key === args.field
      );
      if (col) {
        col.type = args.type;
      }
    }

    const joined = state.joinTableData[args.dataset];
    if (joined) {
      const col = joined.columns.find((c) => c.key === args.field);
      if (col) {
        col.type = args.type;
      }
    }
  },

  reviewVariableType(state: DatasetState, update) {
    const index = _.findIndex(state.variables, (v) => {
      return v.key === update.field;
    });
    state.variables[index].isColTypeReviewed = update.isColTypeReviewed;
  },
  // setIncludedVariableSummary used to clear an individual summary
  setIncludedVariableSummary(
    state: DatasetState,
    args: {
      key: string;
      summary: VariableSummary;
    }
  ) {
    Vue.set(state.includedSet.variableSummariesByKey, args.key, args.summary);
  },
  updateIncludedVariableSummaries(
    state: DatasetState,
    summary: VariableSummary
  ) {
    updateSummariesPerVariable(
      summary,
      state.includedSet.variableSummariesByKey
    );
  },

  updateExcludedVariableSummaries(
    state: DatasetState,
    summary: VariableSummary
  ) {
    updateSummariesPerVariable(
      summary,
      state.excludedSet.variableSummariesByKey
    );
  },
  clearVariableSummary(state: DatasetState, key: string) {
    Vue.delete(state.includedSet, key);
    Vue.delete(state.excludedSet, key);
  },
  clearVariableSummaries(state: DatasetState) {
    state.includedSet.variableSummariesByKey = {};
    state.excludedSet.variableSummariesByKey = {};
  },

  setVariableRankings(
    state: DatasetState,
    args: { dataset: string; rankings: Dictionary<number> }
  ) {
    Vue.set(state.variableRankings, args.dataset, args.rankings);
  },

  updateVariableRankings(state: DatasetState, rankings: Dictionary<number>) {
    // add rank property if ranking data returned, otherwise don't include it
    if (!_.isEmpty(rankings)) {
      state.variables.forEach((v) => {
        let rank = 0;
        if (rankings[v.key]) {
          rank = rankings[v.key];
        }
        Vue.set(v, "ranking", rank);
      });
    } else {
      state.variables.forEach((v) => Vue.delete(v, "ranking"));
    }
  },

  updatePendingRequests(
    state: DatasetState,
    pendingRequest: DatasetPendingRequest
  ) {
    const sameIdIndex = state.pendingRequests.findIndex(
      (item) => pendingRequest.id === item.id
    );
    const sameTypeIndex = state.pendingRequests.findIndex(
      (item) => pendingRequest.type === item.type
    );
    if (sameIdIndex >= 0) {
      Vue.set(state.pendingRequests, sameIdIndex, pendingRequest);
      // only keep latest single request object for each type in the pendingRequests list
    } else if (sameTypeIndex >= 0) {
      Vue.set(state.pendingRequests, sameTypeIndex, pendingRequest);
    } else {
      state.pendingRequests.push(pendingRequest);
    }
  },

  removePendingRequest(state: DatasetState, id: string) {
    state.pendingRequests = state.pendingRequests.filter(
      (item) => item.id !== id
    );
  },

  updateFile(state: DatasetState, args: { url: string; file: any }) {
    Vue.set(state.files, args.url, Object.freeze(args.file));
  },
  bulkUpdateFiles(state: DatasetState, args: { urls: string[]; files: any[] }) {
    const clone = _.cloneDeep(state.files);
    for (let i = 0; i < args.urls.length; ++i) {
      clone[args.urls[i]] = args.files[i];
    }
    Vue.set(state, "files", clone);
  },
  bulkRemoveFiles(state: DatasetState, args: { urls: string[] }) {
    const clone = _.cloneDeep(state.files);
    args.urls.forEach((url) => {
      delete clone[url];
    });
    Vue.set(state, "files", clone);
  },
  removeFile(state: DatasetState, url: string) {
    Vue.delete(state.files, url);
  },

  bulkUpdateTimeseries(
    state: DatasetState,
    args: {
      dataset: string;
      uniqueTrail?: string;
      updates: TimeSeriesUpdate[];
    }
  ) {
    args.updates.forEach((update) => {
      mutations.updateTimeseries(state, {
        dataset: args.dataset,
        uniqueTrail: args.uniqueTrail,
        update: update,
      });
    });
  },

  updateTimeseries(
    state: DatasetState,
    args: {
      dataset: string;
      uniqueTrail?: string;
      update: TimeSeriesUpdate;
    }
  ) {
    const timeseriesKey =
      args.update.variableKey +
      (args.update.seriesID ?? "") +
      (args.uniqueTrail ?? "");

    if (!state.timeseries[args.dataset]) {
      Vue.set(state.timeseries, args.dataset, {});
    }

    if (!state.timeseries[args.dataset].timeseriesData) {
      Vue.set(state.timeseries[args.dataset], "timeseriesData", {});
    }

    // freezing the return to prevent slow, unnecessary deep reactivity.
    Vue.set(
      state.timeseries[args.dataset].timeseriesData,
      timeseriesKey,
      Object.freeze(args.update.timeseries)
    );

    if (!state.timeseries[args.dataset].isDateTime) {
      Vue.set(state.timeseries[args.dataset], "isDateTime", {});
    }
    Vue.set(
      state.timeseries[args.dataset].isDateTime,
      timeseriesKey,
      args.update.isDateTime
    );

    // Set the min/max/mean for each timeseries data
    if (!state.timeseries[args.dataset].info) {
      Vue.set(state.timeseries[args.dataset], "info", {});
    }
    Vue.set(state.timeseries[args.dataset].info, timeseriesKey, {
      min: args.update.min as number,
      max: args.update.max as number,
      mean: args.update.mean as number,
    });

    const validTimeseries = args.update.timeseries.filter((t) => !_.isNil(t));
    const minX = _.minBy(validTimeseries, (d) => d.time).time;
    const maxX = _.maxBy(validTimeseries, (d) => d.time).time;
    const minY = _.minBy(validTimeseries, (d) => d.value).value;
    const maxY = _.maxBy(validTimeseries, (d) => d.value).value;

    if (!state.timeseriesExtrema[args.dataset]) {
      Vue.set(state.timeseriesExtrema, args.dataset, {
        x: {
          min: minX,
          max: maxX,
        },
        y: {
          min: minY,
          max: maxY,
        },
      });
      return;
    }

    const x = state.timeseriesExtrema[args.dataset].x;
    const y = state.timeseriesExtrema[args.dataset].y;
    Vue.set(x, "min", Math.min(x.min, minX));
    Vue.set(x, "max", Math.max(x.max, maxX));
    Vue.set(y, "min", Math.min(y.min, minY));
    Vue.set(y, "max", Math.max(y.max, maxY));
  },
  removeTimeseries(
    state: DatasetState,
    args: { dataset: string; ids: string[] }
  ) {
    args.ids.forEach((id) => {
      delete state.timeseries[args.dataset].timeseriesData[id];
      delete state.timeseries[args.dataset].isDateTime[id];
      delete state.timeseries[args.dataset].info[id];
    });
  },
  setJoinDatasetsTableData(
    state: DatasetState,
    args: { dataset: string; data: TableData }
  ) {
    Vue.set(state.joinTableData, args.dataset, args.data);
  },

  clearJoinDatasetsTableData(state: DatasetState) {
    state.joinTableData = {};
  },
  setBaselineIncludeTableData(state: DatasetState, tableData: TableData) {
    state.baselineIncludeSet = Object.freeze(tableData);
  },
  setBaselineExcludeTableData(state: DatasetState, tableData: TableData) {
    state.baselineExcludeSet = Object.freeze(tableData);
  },
  updateAreaOfInterestIncludeInner(
    state: DatasetState,
    tableData: Map<number, unknown>
  ) {
    updateTableDataItems(state.areaOfInterestIncludeInner, tableData);
  },
  setAreaOfInterestIncludeInner(state: DatasetState, tableData: TableData) {
    state.areaOfInterestIncludeInner = Object.freeze(tableData);
  },
  setAreaOfInterestIncludeOuter(state: DatasetState, tableData: TableData) {
    state.areaOfInterestIncludeOuter = Object.freeze(tableData);
  },
  setAreaOfInterestExcludeInner(state: DatasetState, tableData: TableData) {
    state.areaOfInterestExcludeInner = Object.freeze(tableData);
  },
  setAreaOfInterestExcludeOuter(state: DatasetState, tableData: TableData) {
    state.areaOfInterestExcludeOuter = Object.freeze(tableData);
  },
  clearAreaOfInterestIncludeInner(state: DatasetState) {
    state.areaOfInterestIncludeInner = null;
  },
  clearAreaOfInterestIncludeOuter(state: DatasetState) {
    state.areaOfInterestIncludeOuter = null;
  },
  clearAreaOfInterestExcludeInner(state: DatasetState) {
    state.areaOfInterestExcludeInner = null;
  },
  clearAreaOfInterestExcludeOuter(state: DatasetState) {
    state.areaOfInterestExcludeOuter = null;
  },

  // sets the current selected data into the store
  setIncludedTableData(state: DatasetState, tableData: TableData) {
    // freezing the return to prevent slow, unnecessary deep reactivity.
    state.includedSet.tableData = Object.freeze(tableData);
    // add selected row data to state
    state.includedSet.rowSelectionData = getSelectedRows(true);
  },

  // sets the current excluded data into the store
  setExcludedTableData(state: DatasetState, tableData: TableData) {
    // freezing the return to prevent slow, unnecessary deep reactivity.
    state.excludedSet.tableData = Object.freeze(tableData);
    // add selected row data to state
    state.excludedSet.rowSelectionData = getSelectedRows(false);
  },

  updateRowSelectionData(state) {
    state.includedSet.rowSelectionData = Object.freeze(getSelectedRows(true));
    state.excludedSet.rowSelectionData = Object.freeze(getSelectedRows(false));
  },

  updateTask(state: DatasetState, task: Task) {
    state.task = task;
  },

  updateBands(state: DatasetState, bands: BandCombination[]) {
    state.bands = bands;
  },

  updateMetrics(state: DatasetState, metrics: Metric[]) {
    state.metrics = metrics;
  },
  resetState(state: DatasetState) {
    Object.assign(state, defaultState());
  },
};
