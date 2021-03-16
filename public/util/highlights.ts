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

import { Highlight } from "../store/dataset/index";
import {
  Filter,
  FilterParams,
  FilterObject,
  CATEGORICAL_FILTER,
  CLUSTER_FILTER,
  VECTOR_FILTER,
  INCLUDE_FILTER,
  TEXT_FILTER,
} from "../util/filters";
import { getters as routeGetters } from "../store/route/module";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as resultGetters } from "../store/results/module";
import { overlayRouteEntry } from "../util/routes";
import { getTypeFromKey, getIDFromKey } from "./summaries";
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

export const UPDATE_ALL = "updateAll";
export const UPDATE_FOR_KEY = "updateForKey";

export function encodeHighlights(
  highlights: Highlight | Highlight[],
  deepUpdate: string
): string {
  if (_.isEmpty(highlights)) {
    return null;
  }

  // normalize the highlights to an array
  const incomingHighlights = Array.isArray(highlights)
    ? highlights
    : [highlights];

  // get the current highlights
  const currentHighlights = routeGetters.getDecodedHighlights(store);

  // if updating all highlights, ignored the current highlights
  // if updating for a given variable (key) filter highlights from that key out of the current highlights
  // if updating normally, use the current highlights
  const workingHighlights =
    deepUpdate === UPDATE_ALL
      ? []
      : deepUpdate === UPDATE_FOR_KEY
      ? currentHighlights.filter((wh) => incomingHighlights[0].key !== wh.key)
      : currentHighlights;

  // combine the incoming and working highlight sets
  const allHighlights = [...incomingHighlights, ...workingHighlights];

  // then filter unique highlights, no duplicates
  const uniqueHighlights = allHighlights.reduce((acc, h) => {
    if (!acc.find((uh) => uh.value === h.value)) {
      acc.push(h);
    }
    return acc;
  }, [] as Highlight[]);

  return btoa(JSON.stringify(uniqueHighlights));
}

export function decodeHighlights(highlight: string): Highlight[] {
  if (_.isEmpty(highlight)) {
    return [];
  }
  return JSON.parse(atob(highlight)) as Highlight[];
}
// applies the supplied invert if invert is not present on the filter or highlight object (returns clone)
export function setInvert(
  filterParams: FilterParams,
  invert: boolean
): FilterParams {
  const fp = cloneFilters(filterParams);
  if (!fp.highlights) {
    fp.highlights = { list: [], invert } as FilterObject;
  }
  if (!fp.filters) {
    fp.filters = { list: [], invert } as FilterObject;
  }
  fp.highlights.invert = fp.highlights.invert ?? invert;
  fp.filters.invert = fp.filters.invert ?? invert;
  return fp;
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

    let type = getVarType(key);
    const displayName = variable?.colDisplayName;
    if (!type) {
      type = resultSummaryHighlight(highlight);
    }
    if (type === IMAGE_TYPE) {
      return {
        key: key,
        type: CLUSTER_FILTER,
        mode: highlight.include ?? mode,
        categories: [highlight.value],
        displayName: displayName,
      };
    }

    if (type === TEXT_TYPE) {
      return {
        key: key,
        type: TEXT_FILTER,
        mode: highlight.include ?? mode,
        categories: [highlight.value],
        displayName: displayName,
      };
    }

    if (_.isString(highlight.value)) {
      return {
        key: key,
        type: CATEGORICAL_FILTER,
        mode: highlight.include ?? mode,
        categories: [highlight.value],
        displayName: displayName,
      };
    }
    if (Array.isArray(highlight.value)) {
      return {
        key: key,
        type: CATEGORICAL_FILTER,
        mode: highlight.include ?? mode,
        categories: highlight.value,
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
          mode: highlight.include ?? mode,
          min: highlight.value.from,
          max: highlight.value.to,
          displayName: displayName,
        };
      }

      return {
        key: key,
        type: highlight.value.type,
        mode: highlight.include ?? mode,
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
        type: variable.colType,
        mode: highlight.include ?? mode,
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
export function resultSummaryHighlight(highlight: Highlight) {
  const key = getTypeFromKey(highlight.key);
  const solutionID = getIDFromKey(highlight.key);
  switch (key) {
    case "predicted":
      const predictedSummaries = resultGetters.getPredictedSummaries(store);
      const predictedSummary = predictedSummaries.find((sum) => {
        return sum.solutionId === solutionID;
      });
      if (!predictedSummary) {
        return null;
      }
      return predictedSummary.type;
    case "correctness":
      const correctnessSummaries = resultGetters.getPredictedSummaries(store);
      const correctnessSummary = correctnessSummaries.find((sum) => {
        return sum.solutionId === solutionID;
      });
      if (!correctnessSummary) {
        return null;
      }
      return correctnessSummary.type;
    case "residual":
      const residualSummaries = resultGetters.getPredictedSummaries(store);
      const residualSummary = residualSummaries.find((sum) => {
        return sum.solutionId === solutionID;
      });
      if (!residualSummary) {
        return null;
      }
      return residualSummary.type;
    case "rank":
      const rankSummaries = resultGetters.getPredictedSummaries(store);
      const rankSummary = rankSummaries.find((sum) => {
        return sum.solutionId === solutionID;
      });
      if (!rankSummary) {
        return null;
      }
      return rankSummary.type;
    case "confidence":
      const confidenceSummaries = resultGetters.getPredictedSummaries(store);
      const confidenceSummary = confidenceSummaries.find((sum) => {
        return sum.solutionId === solutionID;
      });
      if (!confidenceSummary) {
        return null;
      }
      return confidenceSummary.type;
    default:
      return null;
  }
}
export function cloneFilters(filterParams: FilterParams): FilterParams {
  return _.cloneDeep(filterParams);
}
export function setHighlightModes(
  filterParams: FilterParams,
  mode: string
): FilterParams {
  filterParams.highlights.list.forEach((highlight) => {
    highlight.mode = mode;
  });
  return filterParams;
}
export function setFilterModes(
  filterParams: FilterParams,
  mode: string
): FilterParams {
  filterParams.filters.list.forEach((filter) => {
    filter.mode = mode;
  });
  return filterParams;
}
export function addHighlightToFilterParams(
  filterParams: FilterParams,
  highlights: Highlight[],
  mode: string = INCLUDE_FILTER
): FilterParams {
  const params = _.cloneDeep(filterParams);
  const highlightFilters = createFiltersFromHighlights(highlights, mode);
  if (highlightFilters.length > 0) {
    params.highlights.list = highlightFilters;
  }
  return params;
}

export function updateHighlight(
  router: VueRouter,
  highlights: Highlight | Highlight[],
  deepUpdate?: string
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
    updateHighlight(router, decodedHighlights, UPDATE_ALL);
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
