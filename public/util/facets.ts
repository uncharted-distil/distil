import _ from "lodash";
import moment from "moment";
import { spinnerHTML } from "../util/spinner";
import {
  formatValue,
  TIMESERIES_TYPE,
  CATEGORICAL_TYPE,
  ORDINAL_TYPE,
  BOOL_TYPE,
  ADDRESS_TYPE,
  CITY_TYPE,
  STATE_TYPE,
  COUNTRY_TYPE,
  EMAIL_TYPE,
  POSTAL_CODE_TYPE,
  PHONE_TYPE,
  URI_TYPE,
  DATE_TIME_TYPE,
  IMAGE_TYPE,
  DATE_TIME_LOWER_TYPE
} from "../util/types";
import { getTimeseriesSummaryTopCategories } from "../util/data";
import {
  VariableSummary,
  CATEGORICAL_SUMMARY,
  NUMERICAL_SUMMARY,
  TimeSeries
} from "../store/dataset/index";
import store from "../store/store";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as resultGetters } from "../store/results/module";
import { Forecast } from "../store/results";

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

export function createGroup(summary: VariableSummary): Group {
  if (summary.err) {
    // create error facet
    return createErrorFacet(summary);
  }
  if (summary.pending) {
    // create pending facet
    return createPendingFacet(summary);
  }
  // create facet
  return createSummaryFacet(summary);
}

// creates a facet to display a data fetch error
export function createErrorFacet(summary: VariableSummary): Group {
  return {
    dataset: summary.dataset,
    colName: summary.key,
    description: summary.description,
    label: summary.label,
    key: `${summary.dataset}:${summary.key}`,
    type: summary.varType,
    collapsible: false,
    collapsed: false,
    facets: [
      {
        placeholder: true,
        html: `<div>${summary.err}</div>`,
        filterable: false
      }
    ],
    summary: null
  };
}

// creates a place holder facet to dispay a spinner
export function createPendingFacet(summary: VariableSummary): Group {
  return {
    dataset: summary.dataset,
    colName: summary.key,
    label: summary.label,
    description: summary.description,
    key: `${summary.dataset}:${summary.key}`,
    type: summary.varType,
    collapsible: false,
    collapsed: false,
    facets: [
      {
        placeholder: true,
        html: spinnerHTML(),
        filterable: false
      }
    ],
    summary: null
  };
}

// creates categorical or numerical summary facets based on the input summary type
export function createSummaryFacet(summary: VariableSummary): Group {
  switch (summary.type) {
    case CATEGORICAL_SUMMARY:
      if (summary.varType === TIMESERIES_TYPE) {
        return createTimeseriesSummaryFacet(summary);
      } else {
        return createCategoricalSummaryFacet(summary);
      }
    case NUMERICAL_SUMMARY:
      return createNumericalSummaryFacet(summary);
  }
  console.warn("unrecognized summary type", summary.type);
  return null;
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

// creates a categorical facet
function createCategoricalSummaryFacet(summary: VariableSummary): Group {
  let total = 0;
  const facets = summary.baseline.buckets.map((b, index) => {
    const segments = [];
    const selected = {
      count: b.count
    };
    const countLabel = b.count.toString();

    const facet: CategoricalFacet = {
      icon: {
        class: getGroupIcon(summary)
      },
      value: b.key,
      countLabel: countLabel,
      count: b.count,
      selected: selected,
      segments: segments,
      filterable: false,
      file: summary.baseline.exemplars
        ? summary.baseline.exemplars[index]
        : null
    };
    total += b.count;
    return facet;
  });

  facets.sort((a, b) => {
    return b.count - a.count;
  });

  const chunkSize = getCategoricalChunkSize(summary.varType);
  const top = facets.slice(0, chunkSize);
  const remaining = facets.length > chunkSize ? facets.slice(chunkSize) : [];
  let remainingTotal = 0;
  remaining.forEach(facet => {
    remainingTotal += facet.count;
  });

  // Generate a facet group
  return {
    dataset: summary.dataset,
    colName: summary.key,
    label: summary.label,
    description: summary.description,
    key: `${summary.dataset}:${summary.key}`,
    type: summary.varType,
    collapsible: false,
    collapsed: false,
    facets: top,
    total: total,
    more: remaining.length,
    moreTotal: remainingTotal,
    all: facets,
    summary: summary
  };
}

function createTimeseriesSummaryFacet(summary: VariableSummary): Group {
  const group = createCategoricalSummaryFacet(summary);

  let timeseries = null as TimeSeries;
  let forecasts = null as Forecast;
  if (summary.solutionId) {
    timeseries = resultGetters.getPredictedTimeseries(store)[
      summary.solutionId
    ];
    forecasts = resultGetters.getPredictedForecasts(store)[summary.solutionId];
  } else {
    timeseries = datasetGetters.getTimeseries(store)[group.dataset];
  }

  group.all.forEach((facet: CategoricalFacet) => {
    if (summary.solutionId) {
      facet.multipleTimeseries = [
        timeseries.timeseriesData[facet.file],
        forecasts.forecastData[facet.file]
      ];
      facet.colors = ["#000", "#00c6e1"];
    } else {
      facet.timeseries = timeseries.timeseriesData[facet.file];
    }
  });

  return group;
}

function getHistogramSlices(summary: VariableSummary) {
  const buckets = summary.baseline.buckets;
  const extrema = summary.baseline.extrema;
  const slices = new Array(buckets.length);
  for (let i = 0; i < buckets.length; i++) {
    const bucket = buckets[i];
    let from: any, to: any;
    if (summary.varType === DATE_TIME_TYPE) {
      from = bucket.key;
      to = i < buckets.length - 1 ? buckets[i + 1].key : extrema.max;
      from = moment
        .unix(_.toNumber(from))
        .utc()
        .format("YYYY/MM/DD");
      to = moment
        .unix(_.toNumber(to))
        .utc()
        .format("YYYY/MM/DD");
    } else {
      from = _.toNumber(bucket.key);
      to =
        i < buckets.length - 1 ? _.toNumber(buckets[i + 1].key) : extrema.max;
    }
    slices[i] = {
      label: `${formatValue(from, summary.varType)}`,
      toLabel: `${formatValue(to, summary.varType)}`,
      count: bucket.count
    };
  }
  return slices;
}

function createNumericalSummaryFacet(summary: VariableSummary): Group {
  const slices = getHistogramSlices(summary);
  return {
    dataset: summary.dataset,
    colName: summary.key,
    label: summary.label,
    description: summary.description,
    key: `${summary.dataset}:${summary.key}`,
    type: summary.varType,
    collapsible: false,
    collapsed: false,
    facets: [
      {
        histogram: {
          slices: slices
        },
        filterable: false,
        selection: {} as any
      }
    ],
    summary: summary
  };
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
      summary.varType === DATE_TIME_TYPE ? DATE_TIME_LOWER_TYPE : summary.type
  };
}

export function getTimeseriesFacetValue(
  summary: VariableSummary,
  type: string
): { from: number; to: number } {
  return {
    from: _.toNumber(
      _.minBy(summary.baseline.buckets, b => _.toNumber(b.key)).key
    ),
    to: _.toNumber(
      _.maxBy(summary.baseline.buckets, b => _.toNumber(b.key)).key
    )
  };
}
