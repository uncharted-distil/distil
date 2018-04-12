import { AppState } from './index';

export const mutations = {

	setAborted(state: AppState) {
		state.isAborted = true;
	}
};
