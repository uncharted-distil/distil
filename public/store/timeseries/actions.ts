import _ from 'lodash';
import axios from 'axios';
import { mutations } from './module'
import { TimeSeriesState } from './index';
import { ActionContext } from 'vuex';
import { DistilState } from '../store';

export type TimeSeriesContext = ActionContext<TimeSeriesState, DistilState>;

export const actions = {

	fetchTimeSeries(context: TimeSeriesContext, args: { url: string }) {
		const TIME_SERIES = [
			'a.csv',
			'b.csv',
			'c.csv',
			'd.csv',
			'e.csv'
		];
		return axios.get(`timeseries/${TIME_SERIES[Math.floor(Math.random() * TIME_SERIES.length)]}`)
			.then(response => {
				const lines = response.data.split('\n');
				const timeseries = lines.slice(1, lines.length - 1).map(entry => {
					const split = entry.split(',');
					return {
						timestamp: split[0],
						count: _.toNumber(split[1])
					};
				});

				mutations.setTimeSeries(context, { url: args.url, timeseries: timeseries });
			})
			.catch(err => {
				console.error(err);
				mutations.setTimeSeries(context, { url: args.url, err: err });
			});

	}
}
