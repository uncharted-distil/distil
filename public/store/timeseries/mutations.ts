import _ from 'lodash';
import Vue from 'vue';
import { TimeSeriesState } from './index';

export const mutations = {

	setTimeSeries(state: TimeSeriesState, args: { url: string, timeseries?: number[][], err?: Error }) {
		if (args.timeseries) {
			Vue.set(state.loadedTimeSeries, args.url, {
				url: args.url,
				timeseries: args.timeseries,
				err: null,
				timestamp: Date.now()
			});
		} else {
			Vue.set(state.loadedTimeSeries, args.url, {
				url: args.url,
				timeseries: null,
				err: args.err,
				timestamp: Date.now()
			});
		}

		// LRU
		const MAX_TIME_SERIES = 100;
		let entries = _.values(state.loadedTimeSeries);
		if (entries.length > MAX_TIME_SERIES) {
			// take n latest
			entries = entries.sort((a: any, b: any) => {
				return b.timestamp - a.timestamp;
			}).slice(0, MAX_TIME_SERIES);
			// remove all others
			state.loadedTimeSeries = {};
			entries.forEach((entry: any) => {
				Vue.set(state.loadedTimeSeries, entry.url, entry);
			});
		}

	}
}
