import { Dictionary } from "../../util/dict";

export const CATEGORICAL_SUMMARY = "categorical";
export const NUMERICAL_SUMMARY = "numerical";

export const D3M_INDEX_FIELD = "d3mIndex";

export const JOIN_DATASET_MAX_SIZE = 100000;

export interface Highlight {
  context: string;
  dataset: string;
  key: string;
  value: any;
  include?: string;
}

export interface Column {
  key: string;
  value: any;
}

export interface Row {
  index: number;
  d3mIndex: number;
  cols: Column[];
  included: boolean;
}

export interface RowSelection {
  context: string;
  d3mIndices: number[];
}

export interface SuggestedType {
  probability: number;
  provenance: string;
  type: string;
}

export function isClusteredGrouping(
  grouping: Grouping
): grouping is ClusteredGrouping {
  return (grouping as ClusteredGrouping).clusterCol !== undefined;
}

export interface Grouping {
  dataset: string;
  idCol: string;
  subIds: string[];
  type: string;
  hidden: string[];
}

export interface ClusteredGrouping extends Grouping {
  clusterCol: string;
}

export interface TimeseriesGrouping extends ClusteredGrouping {
  xCol: string;
  yCol: string;
}

export interface GeoCoordinateGrouping extends Grouping {
  xCol: string;
  yCol: string;
}

export interface MultiBandImageGrouping extends ClusteredGrouping {
  imageCol: string;
  bandCol: string;
}

export interface Variable {
  datasetName: string;
  colDisplayName: string;
  colName: string;
  colType: string;
  importance: number;
  ranking?: number;
  novelty: number;
  colOriginalType: string;
  colDescription: string;
  suggestedTypes: SuggestedType[];
  isColTypeChanged: boolean;
  grouping: Grouping;
  isColTypeReviewed: boolean;
  min: number;
  max: number;
  role: string[];
  distilRole: string;
}

export interface Dataset {
  id: string;
  name: string;
  description: string;
  folder: string;
  summary: string;
  summaryML: string;
  variables: Variable[];
  numBytes: number;
  numRows: number;
  provenance: string;
  source: string;
  joinSuggestion?: JoinSuggestion[];
  joinScore?: number;
  storageName?: string;
  clone?: boolean;
  type: string;
}

export interface JoinSuggestion {
  baseDataset: string;
  baseColumns: string[];
  joinDataset: string;
  joinColumns: string[];
  joinScore: number;
  datasetOrigin?: DatasetOrigin;
  index: number;
}

export interface DatasetOrigin {
  searchResult: string;
  provenance: string;
}

export interface Extrema {
  min: number;
  max: number;
  mean?: number;
}

export interface Bucket {
  key: string;
  count: number;
  buckets?: Bucket[];
}

export interface Histogram {
  buckets?: Bucket[];
  categoryBuckets?: Dictionary<Bucket[]>;
  extrema: Extrema;
  exemplars?: string[];
  stddev?: number;
  mean?: number;
}

export interface VariableSummary {
  label: string;
  description: string;
  key: string;
  dataset: string;
  type?: string;
  varType?: string;
  baseline: Histogram;
  filtered?: Histogram;
  timeline?: Histogram;
  timelineBaseline?: Histogram;
  timelineType?: string;
  selected?: Histogram;
  err?: string;
  weighted?: boolean;
  pending?: boolean;
  solutionId?: string;
}

// Flags the display mode for a variable summary.  Generally Default is correct,
// but in the case of something like a timeseries summary, we can display a sample
// of the series set, or use cluster info to sample.
export enum SummaryMode {
  Default = "default",
  Cluster = "cluster",
  Timeseries = "timeseries",
  MultiBandImage = "multiband_image",
}

// Flags the display mode for filtering.  Generally Default is correct,
// but in the case where a cluster has been applied, then filtering needs to be
// done on the cluster column.
export enum DataMode {
  Default = "default",
  Cluster = "cluster",
}

export interface TableValue {
  value: any;
  weight: number;
  confidence: number;
}

export interface TableData {
  numRows: number;
  numRowsFiltered?: number;
  columns: TableColumn[];
  values: TableValue[][];
  fittedSolutionId: string;
  produceRequestId: string;
}

export interface TableColumn {
  label: string;
  key: string;
  type: string;
  weight: number;
  headerTitle: string;
  sortable?: boolean;
  variant?: string;
}

export interface TableRow {
  _key: number;
  _rowVariant: string;
  _cellVariants: Dictionary<string>;
  coordinates: any;
  d3mIndex?: number;
  isExcluded?: boolean;
}

export interface TimeseriesExtrema {
  x: Extrema;
  y: Extrema;
  sum?: number;
}

// task string definitions - should mirror those defined in the MIT/LL d3m problem schema
export enum TaskTypes {
  CLASSIFICATION = "classification",
  REGRESSION = "regression",
  CLUSTERING = "clustering",
  LINK_PREDICTION = "linkPrediction",
  VERTEX_NOMINATION = "vertexNomination",
  VERTEX_CLASSIFICATION = "vertexClassification",
  COMMUNITY_DETECTION = "communityDetection",
  GRAPH_MATCHING = "graphMatching",
  FORECASTING = "forecasting",
  COLLABORATIVE_FILTERING = "collaborativeFiltering",
  OBJECT_DETECTION = "objectDetection",
  SEMISUPERVISED = "semiSupervised",
  BINARY = "binary",
  MULTICLASS = "multiClass",
  MULTILABEL = "multilabel",
  UNIVARIATE = "univariate",
  MULTIVARIATE = "multivariate",
  OVERLAPPING = "overlapping",
  NONOVERLAPPING = "nonOverlapping",
  TABULAR = "tabular",
  RELATIONAL = "relational",
  IMAGE = "image",
  AUDIO = "audio",
  VIDEO = "video",
  SPEECH = "speech",
  TEXT = "text",
  GRAPH = "graph",
  MULTIGRAPH = "multigraph",
  TIME_SERIES = "timeseries",
  GROUPED = "grouped",
  GEOSPATIAL = "geospatial",
  REMOTE_SENSING = "remoteSensing",
  LUPI = "lupi",
}

export enum BandID {
  NATURAL_COLORS = "natural_colors",
  FALSE_COLOR_INFRARED = "false_color_infrared",
  FALSE_COLOR_URBAN = "false_color_urban",
  AGRICULTURE = "agriculture",
  ATMOSPHERIC_PENETRATION = "atmospheric_penetration",
  HEALTHY_VEGETATION = "healthy_vegetation",
  LAND_WATER = "land_water",
  ATMOSPHERIC_REMOVAL = "atmospheric_removal",
  SHORTWAVE_INFRARED = "shortwave_infrared",
  VEGETATION_ANALYSIS = "vegetation_analysis",
}

export interface Task {
  task: TaskTypes[];
}

export enum DatasetPendingRequestType {
  VARIABLE_RANKING = "VARIABLE_RANKING",
  GEOCODING = "GEOCODING",
  JOIN_SUGGESTION = "JOIN_SUGGESTION",
  JOIN_DATASET_IMPORT = "JOIN_DATASET_IMPORT",
  CLUSTERING = "CLUSTERING",
}

export enum DatasetPendingRequestStatus {
  PENDING = "PENDING",
  RESOLVED = "RESOLVED",
  ERROR = "ERROR",
  REVIEWED = "REVIEWED",
  ERROR_REVIEWED = "ERROR_REVIEWED",
}
export interface ClonedInfo {
  success: boolean;
  clonedDatasetName: string;
}
export interface VariableRankingPendingRequest {
  id: string;
  status: DatasetPendingRequestStatus;
  type: DatasetPendingRequestType.VARIABLE_RANKING;
  dataset: string;
  target: string;
  rankings: Dictionary<number>;
}

export interface GeocodingPendingRequest {
  id: string;
  status: DatasetPendingRequestStatus;
  type: DatasetPendingRequestType.GEOCODING;
  dataset: string;
  field: string;
}

export interface JoinSuggestionPendingRequest {
  id: string;
  status: DatasetPendingRequestStatus;
  type: DatasetPendingRequestType.JOIN_SUGGESTION;
  dataset: string;
  suggestions: Dataset[];
}

export interface JoinDatasetImportPendingRequest {
  id: string;
  status: DatasetPendingRequestStatus;
  type: DatasetPendingRequestType.JOIN_DATASET_IMPORT;
  dataset: string;
}

export interface ClusteringPendingRequest {
  id: string;
  status: DatasetPendingRequestStatus;
  type: DatasetPendingRequestType.CLUSTERING;
  dataset: string;
}

export type DatasetPendingRequest =
  | VariableRankingPendingRequest
  | GeocodingPendingRequest
  | JoinSuggestionPendingRequest
  | JoinDatasetImportPendingRequest
  | ClusteringPendingRequest;

export interface TimeSeriesValue {
  value: number;
  time: number;
  confidenceLow: number;
  confidenceHigh: number;
}

export interface TimeSeries {
  timeseriesData: Dictionary<TimeSeriesValue[]>;
  isDateTime: Dictionary<boolean>;
  info: Dictionary<Extrema>;
}

export interface DatasetState {
  datasets: Dataset[];
  filteredDatasets: Dataset[];
  variables: Variable[];
  variableRankings: Dictionary<Dictionary<number>>;
  files: Dictionary<any>;
  timeseries: Dictionary<TimeSeries>;
  timeseriesExtrema: Dictionary<TimeseriesExtrema>;
  joinTableData: Dictionary<TableData>;
  includedSet: WorkingSet;
  excludedSet: WorkingSet;
  highlightedIncludeSet: TableData;
  highlightedExcludeSet: TableData;
  areaOfInterestIncludeInner: TableData;
  areaOfInterestIncludeOuter: TableData;
  areaOfInterestExcludeInner: TableData;
  areaOfInterestExcludeOuter: TableData;
  pendingRequests: DatasetPendingRequest[];
  task: Task;
  bands: BandCombination[];
  metrics: Metric[];
}

export interface WorkingSet {
  tableData: TableData;
  variableSummariesByKey: Dictionary<Dictionary<VariableSummary>>;
  rowSelectionData?: Row[];
}

export interface BandCombination {
  id: BandID;
  displayName: string;
}

export interface BandCombinations {
  combinations: BandCombination[];
}

export interface Metric {
  id: BandID;
  displayName: string;
  description: string;
}

export interface Metrics {
  metrics: Metric[];
}

export interface MetricDropdownItem {
  value: {
    id: string;
    description: string;
  };
  text: string;
}

export const state: DatasetState = {
  // datasets and filtered datasets
  datasets: [],
  filteredDatasets: [],

  // variable list and rankings for the active dataset
  variables: [],
  variableRankings: {},

  // working set of data
  includedSet: {
    tableData: null,
    variableSummariesByKey: {},
    rowSelectionData: [],
  },
  excludedSet: {
    tableData: null,
    variableSummariesByKey: {},
    rowSelectionData: [],
  },
  // highlight set
  highlightedIncludeSet: null,
  highlightedExcludeSet: null,
  // tiles surrounding tile that was clicked
  // include area of interest
  areaOfInterestIncludeInner: null,
  areaOfInterestIncludeOuter: null,
  // exclude area of interest
  areaOfInterestExcludeInner: null,
  areaOfInterestExcludeOuter: null,
  // linked files / representation data
  files: {},
  timeseries: {},
  timeseriesExtrema: {},

  // joined data table data
  joinTableData: {},

  // pending requests for the active dataset
  pendingRequests: [],

  // task information
  task: {
    task: [TaskTypes.CLASSIFICATION, TaskTypes.MULTICLASS],
  },

  // bands
  bands: [],

  // metrics
  metrics: [],
};
