import { Route } from 'vue-router';

export type Dictionary<T> = { [key: string]: T }

export interface Variable {
	name: string;
	type: string;
	suggestedTypes: string;
}

export interface Datasets {
	name: string;
	description: string;
	variables: Variable[];
}

export interface Extrema {
	min: number,
	max: number
}

export interface Bucket {
	key: string;
	count: number;
}

export interface VariableSummary {
	name: string;
	feature: string;
	buckets: Bucket[];
	extrema: Extrema;
	type?: string;
	err?: string;
	pending?: string;
}

export interface Data {
	name: string;
	columns: string[];
	types: string[];
	values: any[][];
}

export interface Score {
	metric: string;
	value: number;
}

export interface PipelineOutput {
	output: string,
	scores: Score[];
	resultId: string;
}

export interface PipelineInfo {
	requestId: string;
	name: string;
	id: string;
	feature: string;
	pipelineId: string;
	progress: string;
	pipeline?: PipelineOutput;
}

export interface PipelineRequestInfo {
	[pipelineId: string]: PipelineInfo;
}

export interface PipelineState {
	[requestId: string]: PipelineRequestInfo;
}

export interface Session {
	id: string;
	uuids: string[];
}

export interface Range {
	[name: string]: {
		from: number,
		to: number
	}
}

export interface DistilState {
	datasets: Datasets[];
	variables: Variable[];
	variableSummaries: VariableSummary[];
	resultsSummaries: VariableSummary[];
	resultData: Data;
	filteredData: Data;
	selectedData: Data;
	highlightedFeatureRanges: Range;
	highlightedFeatureValues: { [name: string]: any };
	runningPipelines: PipelineState;
	completedPipelines: PipelineState;
	wsConnection: WebSocket;
	pipelineSession: Session;
	route: Route;
}

// shared data model
export const state: DistilState = {
	// description of matched datasets
	datasets: [],

	// variable list for the active dataset
	variables: [],

	// variable summary data for the active dataset
	variableSummaries: [],

	// results summary data for the selected pipeline run
	resultsSummaries: [],

	// current set of pipeline results
	resultData: {} as any,

	// filtered data entries for the active dataset
	filteredData: {} as any,

	// selected data entries for the active dataset
	selectedData: {} as any,

	// highlighted features
	highlightedFeatureRanges: {} as any,

	highlightedFeatureValues: {},

	// running pipeline creation tasks grouped by parent create requestID
	runningPipelines: {} as any,

	// completed pipeline creation tasks grouped by parent create request ID
	completedPipelines: {} as any,

	// the underlying websocket connection
	wsConnection: {} as any,

	// the pipeline session id
	pipelineSession: {} as any,

	route: {} as any
};
