import { Dictionary } from '../../util/dict';

export interface LoadedTimeSeries {
	url: string;
	timeseries: number[][];
	err: Error;
	timestamp: Number;
}

export interface TimeSeriesState {
	loadedTimeSeries: Dictionary<LoadedTimeSeries>;
}

export const state: TimeSeriesState = {
	loadedTimeSeries: {}
}
