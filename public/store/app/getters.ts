import { AppState } from './index';

export const getters = {

	isAborted(state: AppState) {
		return state.isAborted;
	}
};
