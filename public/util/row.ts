import {
  RowSelection,
  Row,
  D3M_INDEX_FIELD,
  state,
} from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import {
  getters as dataGetters,
  actions as dataActions,
} from "../store/dataset/module";
import { getters as resultsGetters } from "../store/results/module";
import { overlayRouteEntry } from "../util/routes";
import { Filter, ROW_FILTER } from "../util/filters";
import { SELECT_TRAINING_ROUTE, RESULTS_ROUTE } from "../store/route/index";
import _ from "lodash";
import store from "../store/store";
import VueRouter from "vue-router";

export function encodeRowSelection(row: RowSelection): string {
  if (_.isEmpty(row)) {
    return null;
  }
  return btoa(JSON.stringify(row));
}

export function decodeRowSelection(row: string): RowSelection {
  if (_.isEmpty(row)) {
    return null;
  }
  return JSON.parse(atob(row)) as RowSelection;
}

export function createFilterFromRowSelection(
  selection: RowSelection,
  mode: string
): Filter {
  if (!selection || selection.d3mIndices.length === 0) {
    return null;
  }
  return {
    type: ROW_FILTER,
    mode: mode,
    d3mIndices: selection.d3mIndices.map((num) => _.toString(num)),
  };
}

export function getNumIncludedRows(selection: RowSelection): number {
  if (!selection || selection.d3mIndices.length === 0) {
    return 0;
  }
  const includedData = dataGetters.getIncludedTableDataItems(store);
  if (!includedData) {
    return 0;
  }
  const d3mIndices = {};
  selection.d3mIndices.forEach((index) => {
    d3mIndices[index] = true;
  });
  return includedData.filter((data) => d3mIndices[data[D3M_INDEX_FIELD]])
    .length;
}

export function getNumExcludedRows(selection: RowSelection): number {
  if (!selection || selection.d3mIndices.length === 0) {
    return 0;
  }
  const excludedData = dataGetters.getExcludedTableDataItems(store);
  const d3mIndices = {};
  selection.d3mIndices.forEach((index) => {
    d3mIndices[index] = true;
  });
  return excludedData.filter((data) => d3mIndices[data[D3M_INDEX_FIELD]])
    .length;
}

export function isRowSelected(
  selection: RowSelection,
  d3mIndex: number
): boolean {
  if (!selection || selection.d3mIndices.length === 0) {
    return false;
  }
  for (let i = 0; i < selection.d3mIndices.length; i++) {
    if (selection.d3mIndices[i] === d3mIndex) {
      return true;
    }
  }
  return false;
}

export function updateTableRowSelection(
  items: any,
  selection: RowSelection,
  context: string
) {
  if (!items) {
    return null;
  }

  // clear selections
  _.forEach(items, (row) => {
    row._rowVariant = null;
  });

  // skip highlighting when the context is the originating table
  if (!selection) {
    return items;
  }

  if (selection.context !== context) {
    return items;
  }
  // add selections
  const d3mIndices = {};
  selection.d3mIndices.forEach((index) => {
    d3mIndices[index] = true;
  });
  items.forEach((item: any) => {
    if (d3mIndices[item[D3M_INDEX_FIELD]]) {
      item._rowVariant = "selected-row";
    }
  });
  return items;
}

export function getSelectedRows(): Row[] {
  const selection = routeGetters.getDecodedRowSelection(store);
  const include = routeGetters.getRouteInclude(store);
  if (!selection || selection.d3mIndices.length === 0) {
    return [];
  }

  const path = routeGetters.getRoutePath(store);

  let tableData = [];

  if (path === SELECT_TRAINING_ROUTE) {
    tableData = include
      ? dataGetters.getIncludedTableDataItems(store)
      : dataGetters.getExcludedTableDataItems(store);
  } else if (path === RESULTS_ROUTE) {
    tableData = include
      ? resultsGetters.getIncludedResultTableDataItems(store)
      : resultsGetters.getExcludedResultTableDataItems(store);
  }

  if (!tableData) {
    return [];
  }

  const d3mIndices = {};
  selection.d3mIndices.forEach((index) => {
    d3mIndices[index] = true;
  });

  const rows = [];
  tableData.forEach((data, index) => {
    if (d3mIndices[data[D3M_INDEX_FIELD]]) {
      rows.push({
        index: index,
        included: include,
        cols: _.map(data, (value, key) => {
          return {
            key: key,
            value: value,
          };
        }),
      });
    }
  });
  return rows;
}

export function addRowSelection(
  router: VueRouter,
  context: string,
  selection: RowSelection,
  d3mIndex: number
) {
  if (!selection || selection.context !== context) {
    selection = {
      context: context,
      d3mIndices: [],
    };
  }
  selection.d3mIndices.push(d3mIndex);
  const entry = overlayRouteEntry(routeGetters.getRoute(store), {
    row: encodeRowSelection(selection),
  });
  router.push(entry).catch((err) => console.warn(err));
  dataActions.updateRowSelectionData(store);
}

export function removeRowSelection(
  router: VueRouter,
  context: string,
  selection: RowSelection,
  d3mIndex: number
) {
  if (!selection) {
    return;
  }
  _.remove(selection.d3mIndices, (r) => {
    return r === d3mIndex;
  });
  if (selection.d3mIndices.length === 0) {
    selection = null;
  }
  const entry = overlayRouteEntry(routeGetters.getRoute(store), {
    row: encodeRowSelection(selection),
  });
  router.push(entry).catch((err) => console.warn(err));
  dataActions.updateRowSelectionData(store);
}

export function clearRowSelection(router: VueRouter) {
  const entry = overlayRouteEntry(routeGetters.getRoute(store), {
    row: null,
  });
  router.push(entry).catch((err) => console.warn(err));
  dataActions.updateRowSelectionData(store);
}
