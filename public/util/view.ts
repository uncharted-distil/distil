import { Location } from 'vue-router';
import { Dictionary } from './dict';
import { LAST_STATE } from '../store/view/index';

export function popViewStack(stack: Dictionary<Dictionary<Location>>, view: string, dataset: string): Location {
	if (!stack[view]) {
		console.log('no previous state for view', view);
		return null;
	}
	if (!dataset) {
		console.log('no dataset available for view', view, 'pull last state');
		return stack[view][LAST_STATE];
	}
	if (!stack[view][dataset]) {
		console.log('no previous state for view', view, 'and dataset', dataset);
		return null;
	}
	return stack[view][dataset];
}
