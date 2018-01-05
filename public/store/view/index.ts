import { Location } from 'vue-router';
import { Dictionary } from '../../util/dict';

export const LAST_STATE: string = '__LAST_STATE__';

export interface ViewState {
	stack: Dictionary<Dictionary<Location>>;
}

export const state: ViewState = {
	// view route stack
	stack:  {} as any
}
