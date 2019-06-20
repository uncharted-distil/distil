import { Dictionary } from '../../util/dict';

export const CATEGORICAL_SUMMARY = 'categorical';
export const NUMERICAL_SUMMARY = 'numerical';
export const TIMESERIES_SUMMMARY = 'timeseries';

export const D3M_INDEX_FIELD = 'd3mIndex';

export interface Highlight {
	context: string;
	dataset: string;
	key: string;
	value: any;
}

export interface Highlight {
	context: string;
	dataset: string;
	key: string;
	value: any;
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

export interface GroupingProperties {
	xCol: string;
	yCol: string;
	clusterCol: string;
}

export interface Grouping {
	dataset: string;
	idCol: string;
	type: string;
	hidden: string[];
	properties?: GroupingProperties;
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
	suggestedTypes: SuggestedType[];
	isColTypeChanged: boolean;
	isGrouping: boolean;
	grouping?: Grouping;
	isColTypeReviewed: boolean;
	min: number;
	max: number;
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
}

export interface JoinSuggestion {
	baseDataset: string;
	baseColumns: string[];
	joinColumns: string[];
}

export interface Extrema {
	min: number;
	max: number;
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
	key: string;
	dataset: string;
	type?: string;
	varType?: string;
	solutionId?: string;
	baseline: Histogram;
	filtered?: Histogram;
	err?: string;
	pending?: boolean;
}

export interface TimeseriesSummary {
	label: string;
	key: string;
	dataset: string;
	type?: string;
	varType?: string;
	err?: string;
	pending?: boolean;
}

export interface TableData {
	numRows: number;
	columns: TableColumn[];
	values: any[][];
}

export interface TableColumn {
	label: string;
	key: string;
	type: string;
	sortable?: boolean;
	variant?: string;
}

export interface TableRow {
	_key: number;
	_rowVariant: string;
	_cellVariants: Dictionary<string>;
}

export interface TimeseriesExtrema {
	x: Extrema;
	y: Extrema;
	sum?: number;
}

export enum DatasetPendingRequestType {
	VARIABLE_RANKING = 'VARIABLE_RANKING',
	GEOCODING = 'GEOCODING',
	JOIN_SUGGESTION = 'JOIN_SUGGESTION',
	JOIN_DATASET_IMPORT = 'JOIN_DATASET_IMPORT',
}

export enum DatasetPendingRequestStatus {
	PENDING = 'PENDING',
	RESOLVED = 'RESOLVED',
	ERROR = 'ERROR',
	REVIEWED = 'REVIEWED',
	ERROR_REVIEWED = 'ERROR_REVIEWED',
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

export type DatasetPendingRequest =
		VariableRankingPendingRequest
	| GeocodingPendingRequest
	| JoinSuggestionPendingRequest
	| JoinDatasetImportPendingRequest;

export interface DatasetState {
	datasets: Dataset[];
	filteredDatasets: Dataset[];
	variables: Variable[];
	variableSummaries: VariableSummary[];
	variableRankings: Dictionary<Dictionary<number>>;
	timeVariableSummaries: VariableSummary[];
	files: Dictionary<any>;
	timeseries: Dictionary<Dictionary<number[][]>>;
	timeseriesExtrema: Dictionary<TimeseriesExtrema>;
	joinTableData: Dictionary<TableData>;
	includedTableData: TableData;
	excludedTableData: TableData;
	pendingRequests: DatasetPendingRequest[];
}


export const state: DatasetState = {
	// datasets and filtered datasets
	datasets: [],
	filteredDatasets: [],

	// variable list for the active dataset
	variables: [],

	// variable summary data for the active dataset
	variableSummaries: [],

	// variable summary data for the variable used for time series analysis
	timeVariableSummaries: [],

	// variable rankings per dataset
	variableRankings: {},

	// linked files
	files: {},

	timeseries: {},
	timeseriesExtrema: {},

	// joined data table data
	joinTableData: {},

	// selected data entries for the active dataset
	includedTableData: null,

	// excluded data entries for the active dataset
	excludedTableData: null,

	// pending requests for the active dataset
	pendingRequests: [],
};
