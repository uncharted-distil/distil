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
import {
  CATEGORICAL_SUMMARY,
  NUMERICAL_SUMMARY,
  RowSelection,
  VariableSummary,
} from "../store/dataset/index";
import { getters as datasetGetters } from "../store/dataset/module";
import store from "../store/store";
import { getTimeseriesSummaryTopCategories } from "../util/data";
import {
  ADDRESS_TYPE,
  BOOL_TYPE,
  CATEGORICAL_TYPE,
  CITY_TYPE,
  COUNTRY_TYPE,
  DATE_TIME_LOWER_TYPE,
  DATE_TIME_TYPE,
  DISTIL_ROLES,
  EMAIL_TYPE,
  GEOCOORDINATE_TYPE,
  IMAGE_TYPE,
  ORDINAL_TYPE,
  PHONE_TYPE,
  POSTAL_CODE_TYPE,
  STATE_TYPE,
  TIMESERIES_TYPE,
  URI_TYPE,
} from "../util/types";
import { ColorScaleNames, COLOR_SCALES, DISCRETE_COLOR_MAPS } from "./color";

export const CATEGORICAL_CHUNK_SIZE = 5;
export const IMAGE_CHUNK_SIZE = 5;

export const MID_RANGE_HIGHLIGHT = "bell";
export const TOP_RANGE_HIGHLIGHT = "top";
export const BOTTOM_RANGE_HIGHLIGHT = "bottom";
export const DEFAULT_HIGHLIGHT_PERCENTILE = 0.75;

export const FACET_COLOR_SELECT = { color: "#ff0067", colorHover: "#ffaaaa" };
export const FACET_COLOR_EXCLUDE = { color: "#000000", colorHover: "#333333" };
export const FACET_COLOR_FILTERED = { color: "#999999", colorHover: "#bbbbbb" };
export const FACET_COLOR_ERROR = { color: "#e05353", colorHover: "#e0aaaa" };

export interface PlaceHolderFacet {
  placeholder: boolean;
  html: string;
  filterable: boolean;
}

export interface Segment {
  color: string;
  count: number;
}
export interface FacetColor {
  color: string;
  colorHover: string;
}
export interface SelectedSegments {
  selected: number;
  segments: Segment[];
}

export interface CategoricalFacet {
  icon: { class: string };
  selected: { count: number } | SelectedSegments;
  value: string;
  count: number;
  countLabel: string;
  filterable: boolean;
  segments: Segment[];
  timeseries?: number[][];
  multipleTimeseries?: number[][][];
  colors?: string[];
  file: string;
}

export interface Slice {
  label: string;
  toLabel: string;
  count: number;
}

export interface Selection {
  range: {
    to: string;
    from: string;
  };
}

export interface NumericalFacet {
  histogram: { slices: Slice[] };
  filterable: boolean;
  selection: Selection;
}

export interface Group {
  dataset: string;
  label: string;
  description: string;
  key: string;
  groupKey: string;
  type: string;
  collapsible: boolean;
  collapsed: boolean;
  facets: (PlaceHolderFacet | CategoricalFacet | NumericalFacet)[];
  more?: number;
  moreTotal?: number;
  total?: number;
  less?: number;
  all?: (PlaceHolderFacet | CategoricalFacet | NumericalFacet)[];
  isImportant?: boolean;
  summary: VariableSummary;
}

export function getGroupIcon(summary: VariableSummary): string {
  if (summary.distilRole === DISTIL_ROLES.Augmented) return "fa fa-code-fork";
  switch (summary.varType) {
    case CATEGORICAL_TYPE:
    case ORDINAL_TYPE:
    case BOOL_TYPE:
      return "fa fa-info";

    case ADDRESS_TYPE:
    case CITY_TYPE:
    case STATE_TYPE:
    case COUNTRY_TYPE:
      return "fa fa-globe";

    case EMAIL_TYPE:
    case POSTAL_CODE_TYPE:
      return "fa fa-envelope";

    case PHONE_TYPE:
      return "fa fa-phone";

    case URI_TYPE:
    case "keyword":
      return "fa fa-book";

    case DATE_TIME_TYPE:
      return "fa fa-calendar";

    default:
      return "fa fa-info";
  }
}

export function getCategoricalChunkSize(type: string): number {
  if (type === IMAGE_TYPE) {
    return IMAGE_CHUNK_SIZE;
  }
  return CATEGORICAL_CHUNK_SIZE;
}

export function isCategoricalFacet(
  facet: PlaceHolderFacet | CategoricalFacet | NumericalFacet
): facet is CategoricalFacet {
  return (<CategoricalFacet>facet).value !== undefined;
}

export function isNumericalFacet(
  facet: PlaceHolderFacet | CategoricalFacet | NumericalFacet
): facet is NumericalFacet {
  return (<NumericalFacet>facet).histogram !== undefined;
}

export function isPlaceHolderFacet(
  facet: PlaceHolderFacet | CategoricalFacet | NumericalFacet
): facet is PlaceHolderFacet {
  return (<PlaceHolderFacet>facet).placeholder !== undefined;
}

export function getCategoricalFacetValue(summary: VariableSummary): string {
  return summary.baseline.categoryBuckets
    ? getTimeseriesSummaryTopCategories(summary)[0]
    : summary.baseline.buckets[0].key;
}

export function getNumericalFacetValue(
  summary: VariableSummary,
  type: string
): { from: number; to: number; type: string } {
  // facet library is incapable of selecting a range that isnt exactly
  // on a bin boundary, so we need to iterate through and find it
  // manually.
  const extrema = summary.baseline.extrema;

  let from = extrema.min;
  let to = extrema.max;
  if (
    summary.baseline.mean !== undefined &&
    summary.baseline.stddev !== undefined
  ) {
    switch (type) {
      case TOP_RANGE_HIGHLIGHT:
        from =
          summary.baseline.mean +
          summary.baseline.stddev * DEFAULT_HIGHLIGHT_PERCENTILE;
        break;

      case BOTTOM_RANGE_HIGHLIGHT:
        to =
          summary.baseline.mean -
          summary.baseline.stddev * DEFAULT_HIGHLIGHT_PERCENTILE;
        break;

      case MID_RANGE_HIGHLIGHT:
        from =
          summary.baseline.mean -
          summary.baseline.stddev * DEFAULT_HIGHLIGHT_PERCENTILE;
        to =
          summary.baseline.mean +
          summary.baseline.stddev * DEFAULT_HIGHLIGHT_PERCENTILE;
        break;
    }
  } else {
    const range = extrema.max - extrema.min;
    const mid = (extrema.max + extrema.min) / 2;
    switch (type) {
      case TOP_RANGE_HIGHLIGHT:
        from = extrema.min + range * DEFAULT_HIGHLIGHT_PERCENTILE;
        break;

      case BOTTOM_RANGE_HIGHLIGHT:
        to = extrema.max - range * DEFAULT_HIGHLIGHT_PERCENTILE;
        break;

      case MID_RANGE_HIGHLIGHT:
        from = mid - range * DEFAULT_HIGHLIGHT_PERCENTILE;
        to = mid + range * DEFAULT_HIGHLIGHT_PERCENTILE;
        break;
    }
  }
  const buckets = summary.baseline.buckets;
  // case case set to full range
  let fromSlice = _.toNumber(buckets[0].key);
  let toSlice = _.toNumber(buckets[buckets.length - 1].key);
  // try to narrow into percentile
  for (let i = 0; i < buckets.length; i++) {
    const slice = _.toNumber(buckets[i].key);
    if (from <= slice) {
      fromSlice = slice;
      break;
    }
  }
  for (let i = buckets.length - 1; i >= 0; i--) {
    const slice = _.toNumber(buckets[i].key);
    if (to >= slice) {
      toSlice = slice;
      break;
    }
  }
  return {
    from: fromSlice,
    to: toSlice,
    type:
      summary.varType === DATE_TIME_TYPE ? DATE_TIME_LOWER_TYPE : summary.type,
  };
}

export function getTimeseriesFacetValue(
  summary: VariableSummary,
  type: string
): { from: number; to: number } {
  return {
    from: _.toNumber(
      _.minBy(summary.baseline.buckets, (b) => _.toNumber(b.key)).key
    ),
    to: _.toNumber(
      _.maxBy(summary.baseline.buckets, (b) => _.toNumber(b.key)).key
    ),
  };
}

export function hasSummary(summary: VariableSummary) {
  return !!summary;
}

export function hasBaseline(summary: VariableSummary) {
  return (
    hasSummary(summary) &&
    !!summary.baseline &&
    !!summary.baseline.buckets &&
    summary.baseline.buckets.length > 0
  );
}

export function hasFiltered(summary: VariableSummary) {
  return (
    hasSummary(summary) &&
    !!summary.filtered &&
    !!summary.filtered.buckets &&
    summary.filtered.buckets.length > 0
  );
}

// applyColor generates the string to change the facet dynamic css variables
export function applyColor(
  colors: FacetColor[] | null,
  startIndex?: number
): string {
  let result = "";
  const start = startIndex ?? 0;
  colors.forEach((c, i) => {
    const j = i + start;
    if (!!c?.color && !!c?.colorHover) {
      result += `--facet-bars-${j}-normal: ${c.color};
      --facet-bars-${j}-normal-contrast: ${c.colorHover};
      --facet-bars-${j}-normal-contrast-hover: ${c.color};
      --facet-bars-${j}-selected: ${c.color};
      --facet-bars-${j}-selected-contrast: ${c.colorHover};
      --facet-bars-${j}-selected-contrast-hover: ${c.color};
      --facet-bars-${j}-unselected: ${c.colorHover};
      --facet-bars-${j}-unselected-contrast: ${c.colorHover};
      --facet-bars-${j}-unselected-contrast-hover: ${c.color};
      --facet-bars-${j}-muted: ${c.color};
      --facet-bars-${j}-muted-contrast: ${c.color};
      --facet-bars-${j}-muted-contrast-hover: ${c.colorHover};
      --facet-terms-bar-${j}-normal: ${c.color};
      --facet-terms-bar-${j}-normal-contrast: ${c.colorHover};
      --facet-terms-bar-${j}-normal-contrast-hover: ${c.color};
      --facet-terms-bar-${j}-selected: ${c.color};
      --facet-terms-bar-${j}-selected-contrast: ${c.colorHover};
      --facet-terms-bar-${j}-selected-contrast-hover: ${c.color};
      --facet-terms-bar-${j}-unselected: ${c.colorHover};
      --facet-terms-bar-${j}-unselected-contrast: ${c.colorHover};
      --facet-terms-bar-${j}-unselected-contrast-hover: ${c.color};
      --facet-terms-bar-${j}-muted: ${c.color};
      --facet-terms-bar-${j}-muted-contrast: ${c.color};
      --facet-terms-bar-${j}-muted-contrast-hover: ${c.colorHover};`;
    }
  });
  return result;
}
export function generateFacetDiscreteStyle(
  facetId: string,
  partId: string,
  variableSummary: VariableSummary,
  colorScaleName: ColorScaleNames
): string {
  let result = "";
  const colorScale = DISCRETE_COLOR_MAPS.get(colorScaleName);
  const end = colorScale.length - 1;
  variableSummary.baseline.buckets.forEach((_, idx) => {
    const idSelector = idx.toString().split("").join(" ");
    const colorIdx = Math.min(idx, end);
    result += `#${facetId} #\\3${idSelector}::part(${partId}){background-color:${colorScale[colorIdx]}}`;
  });
  return result;
}
export function generateFacetLinearStyle(
  facetId: string,
  partId: string,
  variableSummary: VariableSummary,
  colorScaleName: ColorScaleNames
): string {
  let result = "";
  const colorScale = COLOR_SCALES.get(colorScaleName);
  const end = variableSummary.baseline.buckets.length - 1;
  variableSummary.baseline.buckets.forEach((_, idx) => {
    const idSelector = idx.toString().split("").join(" ");
    result += `#${facetId} #\\3${idSelector}::part(${partId}){background-color:${colorScale(
      idx / end
    )}}`;
  });
  return result;
}

export function getSubSelectionValues(
  summary: VariableSummary,
  rowSelection: RowSelection,
  max: number,
  include: boolean
): number[][] {
  if (!summary.baseline?.buckets) {
    return [];
  }
  const hasFilterBuckets = hasFiltered(summary);
  if (!hasFilterBuckets && !rowSelection) {
    return summary.baseline?.buckets?.map((b) => [null, b.count / max]);
  }
  const isNumeric = summary.type === NUMERICAL_SUMMARY;
  const rowLabels = getRowSelectionLabels(summary, include);
  let subSelectionValues = null;

  if (hasFilterBuckets) {
    const filteredKeys = summary.filtered.buckets.reduce((acc, b) => {
      acc[b.key] = b.count;
      return acc;
    }, {});
    const variableKeys = summary.baseline.buckets.reduce((acc, b) => {
      acc[b.key] = b.count;
      return acc;
    }, {});
    subSelectionValues = summary.baseline.buckets.map((b) => {
      const hasRowLabels = rowLabelMatches(rowLabels, b.key, isNumeric);
      const bucketCount = hasRowLabels
        ? filteredKeys[b.key]
          ? filteredKeys[b.key]
          : variableKeys[b.key]
          ? variableKeys[b.key]
          : 0
        : filteredKeys[b.key]
        ? filteredKeys[b.key]
        : 0;
      return hasRowLabels
        ? [null, bucketCount / max, null]
        : include
        ? [null, null, bucketCount / max]
        : [bucketCount / max, null, null];
    });
  } else {
    subSelectionValues = summary.baseline.buckets.map((b) =>
      rowLabelMatches(rowLabels, b.key, isNumeric)
        ? [null, b.count / max, null]
        : [null, null, b.count / max]
    );
  }
  return subSelectionValues;
}

export function rowLabelMatches(
  rowLabels: string[],
  bucketKey: string,
  isNumeric: boolean
): boolean {
  if (isNumeric) {
    const numBk = _.toNumber(bucketKey);
    return rowLabels.reduce((hasRl: boolean, rl: string) => {
      if (_.toNumber(rl) === numBk) {
        hasRl = true;
      }
      return hasRl;
    }, false);
  } else {
    return rowLabels.reduce((hasRl: boolean, rl: string) => {
      if (rl === bucketKey) {
        hasRl = true;
      }
      return hasRl;
    }, false);
  }
}

export function getRowSelectionLabels(
  summary: VariableSummary,
  include: boolean
): string[] {
  if (!summary.baseline?.buckets) {
    return [];
  }
  const selectedRows = include
    ? datasetGetters.getIncludedSelectedRowData(store)
    : datasetGetters.getExcludedSelectedRowData(store);
  if (selectedRows.length === 0) return [];
  let rowKeys = [];
  let rowLabels = [];

  selectedRows.forEach((row) =>
    row.cols.forEach((col) => {
      if (col.key === summary.label || col.key === summary.key) {
        rowKeys.push(col.value.value);
      }
    })
  );
  // if date time parse into numeric value
  if (summary.varType === DATE_TIME_TYPE) {
    rowKeys = rowKeys.map((key) => {
      return (Date.parse(key) / 1000).toString(); // convert to seconds
    });
  }
  if (summary.type === NUMERICAL_SUMMARY) {
    const bucketFloors = summary.baseline.buckets.map((b) => _.toNumber(b.key));
    rowKeys = rowKeys.map((rk) => _.toNumber(rk));
    rowLabels = rowKeys.map((rk) => {
      return `${bucketFloors.filter((bf) => rk >= bf).pop()}`;
    });
  } else {
    rowLabels = summary.baseline.buckets.reduce((acc, b) => {
      if (rowKeys.indexOf(b.key) > -1) {
        acc.push(b.key);
      }
      return acc;
    }, []);
  }
  return rowLabels;
}

export function getFacetByType(type: string): string {
  switch (type) {
    case CATEGORICAL_SUMMARY:
      return "facet-categorical";
    case NUMERICAL_SUMMARY:
      return "facet-numerical";
    case GEOCOORDINATE_TYPE:
      return "geocoordinate-facet";
    case TIMESERIES_TYPE:
      return "facet-timeseries";
    default:
      return null;
  }
}

export function viewMoreData(
  moreNumToDisplay: number,
  facetMoreCount: number,
  baseNumToDisplay: number,
  facetValueCount: number
): number {
  return facetMoreCount >= baseNumToDisplay
    ? moreNumToDisplay + baseNumToDisplay
    : moreNumToDisplay + (facetValueCount % baseNumToDisplay);
}

export function viewLessData(
  moreNumToDisplay: number,
  facetMoreCount: number,
  baseNumToDisplay: number,
  facetValueCount: number
): number {
  return facetMoreCount === 0 && facetValueCount % baseNumToDisplay !== 0
    ? moreNumToDisplay - (facetValueCount % baseNumToDisplay)
    : moreNumToDisplay - baseNumToDisplay;
}

export function facetTypeChangeState(
  dataset: string,
  key: string,
  enabledTypeChanges: string[]
): boolean {
  const typeKey = `${dataset}:${key}`;
  return enabledTypeChanges
    ? Boolean(enabledTypeChanges.find((e) => e === typeKey))
    : false;
}
