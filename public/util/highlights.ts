import { Highlight } from "../store/dataset/index";
import {
  Filter,
  FilterParams,
  CATEGORICAL_FILTER,
  GEOBOUNDS_FILTER,
  CLUSTER_FILTER,
  VECTOR_FILTER,
  INCLUDE_FILTER,
  TEXT_FILTER,
} from "../util/filters";
import { getters as routeGetters } from "../store/route/module";
import { getters as datasetGetters } from "../store/dataset/module";
import { overlayRouteEntry } from "../util/routes";
import {
  TIMESERIES_TYPE,
  IMAGE_TYPE,
  TEXT_TYPE,
  getVarType,
  isCollectionType,
} from "../util/types";
import _ from "lodash";
import store from "../store/store";
import VueRouter from "vue-router";

export function encodeHighlights(highlight: Highlight): string {
  if (_.isEmpty(highlight)) {
    return null;
  }
  return btoa(JSON.stringify(highlight));
}

export function decodeHighlights(highlight: string): Highlight {
  if (_.isEmpty(highlight)) {
    return null;
  }
  return JSON.parse(atob(highlight)) as Highlight;
}

export function createFilterFromHighlight(
  highlight: Highlight,
  mode: string
): Filter {
  if (!highlight || highlight.value === null || highlight.value === undefined) {
    return null;
  }

  // inject metadata prefix for metadata vars
  const key = highlight.key;

  const variables = datasetGetters.getVariables(store);

  const variable = variables.find((v) => v.colName === key);
  let grouping = null;
  if (variable && variable.grouping) {
    grouping = variable.grouping;
  }

  const type = getVarType(key);
  const displayName = variable?.colDisplayName;
  if (type === IMAGE_TYPE) {
    return {
      key: key,
      type: CLUSTER_FILTER,
      mode: mode,
      categories: [highlight.value],
      displayName: displayName,
    };
  }

  if (type === TEXT_TYPE) {
    return {
      key: key,
      type: TEXT_FILTER,
      mode: mode,
      categories: [highlight.value],
      displayName: displayName,
    };
  }

  if (_.isString(highlight.value)) {
    return {
      key: key,
      type: CATEGORICAL_FILTER,
      mode: mode,
      categories: [highlight.value],
      displayName: displayName,
    };
  }

  if (highlight.value.from !== undefined && highlight.value.to !== undefined) {
    // TODO: we currently have no support for filter timeseries data by
    // ranges and handle it in the client.
    if (grouping && grouping.type === TIMESERIES_TYPE) {
      return null;
    }

    if (isCollectionType(type)) {
      return {
        key: key,
        type: VECTOR_FILTER,
        nestedType: highlight.value.type,
        mode: mode,
        min: highlight.value.from,
        max: highlight.value.to,
        displayName: displayName,
      };
    }

    return {
      key: key,
      type: highlight.value.type,
      mode: mode,
      min: highlight.value.from,
      max: highlight.value.to,
      displayName: displayName,
    };
  }
  if (
    highlight.value.minX !== undefined &&
    highlight.value.maxX !== undefined &&
    highlight.value.minY !== undefined &&
    highlight.value.maxY !== undefined
  ) {
    return {
      key: key,
      type: GEOBOUNDS_FILTER,
      mode: mode,
      minX: highlight.value.minX,
      maxX: highlight.value.maxX,
      minY: highlight.value.minY,
      maxY: highlight.value.maxY,
      displayName: displayName,
    };
  }
  return null;
}

export function addHighlightToFilterParams(
  filterParams: FilterParams,
  highlight: Highlight,
  mode: string = INCLUDE_FILTER
): FilterParams {
  if (highlight && highlight.include) {
    // added the potential for tweaking the mode outside of the immediate store code
    mode = highlight.include;
  }
  const params = _.cloneDeep(filterParams);
  const highlightFilter = createFilterFromHighlight(highlight, mode);
  if (highlightFilter) {
    params.highlight = highlightFilter;
  }
  return params;
}

export function updateHighlight(router: VueRouter, highlight: Highlight) {
  const entry = overlayRouteEntry(routeGetters.getRoute(store), {
    highlights: encodeHighlights(highlight),
    row: null, // clear row
  });
  router.push(entry).catch((err) => console.warn(err));
}

export function clearHighlight(router: VueRouter) {
  const entry = overlayRouteEntry(routeGetters.getRoute(store), {
    highlights: null,
    row: null, // clear row
  });
  router.push(entry).catch((err) => console.warn(err));
}

export function highlightsExist(router: VueRouter) {
  const route = routeGetters.getRoute(store);
  const highlights = "highlights";
  return route.query[highlights] !== null;
}
