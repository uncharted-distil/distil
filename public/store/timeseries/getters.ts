import { TimeSeriesState } from './index';
import { Dictionary } from '../../util/dict';

export const getters = {

	getTimeSeries(state: TimeSeriesState): Dictionary<any> {
		return state.loadedTimeSeries;
	}
}
