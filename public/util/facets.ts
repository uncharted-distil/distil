import _ from "lodash";
import {
  CATEGORICAL_SUMMARY,
  NUMERICAL_SUMMARY,
  RowSelection,
  VariableSummary,
} from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
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

export const CATEGORICAL_CHUNK_SIZE = 5;
export const IMAGE_CHUNK_SIZE = 5;

export const MID_RANGE_HIGHLIGHT = "bell";
export const TOP_RANGE_HIGHLIGHT = "top";
export const BOTTOM_RANGE_HIGHLIGHT = "bottom";
export const DEFAULT_HIGHLIGHT_PERCENTILE = 0.75;

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
  colName: string;
  label: string;
  description: string;
  key: string;
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
export function applyColor(colors: FacetColor[]): string {
  let result = "";
  colors.forEach((c, i) => {
    if (!!c?.color && !!c?.colorHover) {
      result += `--facet-bars-${i}-normal: ${c.color};
      --facet-bars-${i}-normal-contrast: ${c.colorHover};
      --facet-bars-${i}-normal-contrast-hover: ${c.color};
      --facet-bars-${i}-selected: ${c.color};
      --facet-bars-${i}-selected-contrast: ${c.colorHover};
      --facet-bars-${i}-selected-contrast-hover: ${c.color};
      --facet-bars-${i}-unselected: ${c.colorHover};
      --facet-bars-${i}-unselected-contrast: ${c.colorHover};
      --facet-bars-${i}-unselected-contrast-hover: ${c.color};
      --facet-bars-${i}-muted: ${c.color};
      --facet-bars-${i}-muted-contrast: ${c.color};
      --facet-bars-${i}-muted-contrast-hover: ${c.colorHover};
      --facet-terms-bar-${i}-normal: ${c.color};
      --facet-terms-bar-${i}-normal-contrast: ${c.colorHover};
      --facet-terms-bar-${i}-normal-contrast-hover: ${c.color};
      --facet-terms-bar-${i}-selected: ${c.color};
      --facet-terms-bar-${i}-selected-contrast: ${c.colorHover};
      --facet-terms-bar-${i}-selected-contrast-hover: ${c.color};
      --facet-terms-bar-${i}-unselected: ${c.colorHover};
      --facet-terms-bar-${i}-unselected-contrast: ${c.colorHover};
      --facet-terms-bar-${i}-unselected-contrast-hover: ${c.color};
      --facet-terms-bar-${i}-muted: ${c.color};
      --facet-terms-bar-${i}-muted-contrast: ${c.color};
      --facet-terms-bar-${i}-muted-contrast-hover: ${c.colorHover};`;
    }
  });
  return result;
}
export function getSubSelectionValues(
  summary: VariableSummary,
  rowSelection: RowSelection,
  max: number
): number[][] {
  if (!summary.baseline?.buckets) {
    return [];
  }
  const include = routeGetters.getRouteInclude(store);
  const hasFilterBuckets = hasFiltered(summary);
  if (!hasFilterBuckets && !rowSelection) {
    return summary.baseline?.buckets?.map((b) => [null, b.count / max]);
  }
  const isNumeric = summary.type === NUMERICAL_SUMMARY;
  const rowLabels = getRowSelectionLabels(summary);
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

export function getRowSelectionLabels(summary: VariableSummary): string[] {
  if (!summary.baseline?.buckets) {
    return [];
  }
  const include = routeGetters.getRouteInclude(store);
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
