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

export function encodeHighlights(
  highlights: Highlight | Highlight[],
  deepUpdate: boolean
): string {
  if (_.isEmpty(highlights)) {
    return null;
  }
  const currentHighlights = deepUpdate
    ? []
    : routeGetters.getDecodedHighlights(store);
  if (Array.isArray(highlights)) {
    return btoa(JSON.stringify([...highlights, ...currentHighlights]));
  } else {
    return btoa(JSON.stringify([highlights, ...currentHighlights]));
  }
}

export function decodeHighlights(highlight: string): Highlight[] {
  if (_.isEmpty(highlight)) {
    return [];
  }
  return JSON.parse(atob(highlight)) as Highlight[];
}

export function createFiltersFromHighlights(
  highlights: Highlight[],
  mode: string
): Filter[] {
  if (!highlights || highlights.length < 1) {
    return [];
  }

  const filterHighlights = highlights.map((highlight) => {
    // inject metadata prefix for metadata vars
    const key = highlight.key;

    const variables = datasetGetters.getVariables(store);

    const variable = variables.find((v) => v.key === key);
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

    if (
      highlight.value.from !== undefined &&
      highlight.value.to !== undefined
    ) {
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
  });

  return filterHighlights;
}

export function addHighlightToFilterParams(
  filterParams: FilterParams,
  highlights: Highlight[],
  mode: string = INCLUDE_FILTER
): FilterParams {
  const params = _.cloneDeep(filterParams);
  const highlightFilters = createFiltersFromHighlights(highlights, mode);
  if (highlightFilters.length > 0) {
    params.highlights = highlightFilters;
  }
  return params;
}

export function updateHighlight(
  router: VueRouter,
  highlights: Highlight | Highlight[],
  deepUpdate?: boolean
) {
  const entry = overlayRouteEntry(routeGetters.getRoute(store), {
    highlights: encodeHighlights(highlights, deepUpdate),
    row: null, // clear row
  });
  router.push(entry).catch((err) => console.warn(err));
}

export function clearHighlight(router: VueRouter, key?: string) {
  if (!key) {
    // no key, clear everything
    const entry = overlayRouteEntry(routeGetters.getRoute(store), {
      highlights: null,
      row: null, // clear row
    });
    router.push(entry).catch((err) => console.warn(err));
  } else {
    // key, clear everything with that key
    const highlights = routeGetters.getRouteHighlight(store);
    const decodedHighlights = decodeHighlights(highlights).filter((h) => {
      return h.key && h.key !== key;
    });
    updateHighlight(router, decodedHighlights, true);
  }
}

export function highlightsExist() {
  const route = routeGetters.getRoute(store);
  const highlights = "highlights";
  return route.query[highlights] !== null;
}

export function hasHighlightInRoute(key: string): boolean {
  // retrieve the highlights from the route

  const highlights = routeGetters.getRouteHighlight(store);
  const decoded = decodeHighlights(highlights);
  return (
    decoded.filter((h) => {
      return h.key && h.key === key;
    }).length > 0
  );
}
