import _ from 'lodash';
import * as index from './index';

export function setDatasets(state, datasets) {
	state.datasets = datasets;
}

export function addDataset(state, dataset) {
	state.datasets.push(dataset);
}

export function removeDataset(state, id) {
	return !_.isUndefined(_.remove(state.datasets, elem => elem.name === id));
}

export function setActiveDataset(state, id) {
	state.activeDataset = id;
}

export function setVariableSummaries(state, summaries) {
	state.variableSummaries = summaries;
}

export function updateVariableSummaries(state, args) {
	state.variableSummaries.splice(args.index, 1);
	state.variableSummaries.splice(args.index, 0, args.histogram);
}

export function setFilteredData(state, filteredData) {
	state.filteredData = filteredData;
}

// toggles variable enabled/disable state
export function setVarEnabled(state, args) {
	const varState = state.filterState[args.name];
	if (varState) {
		 varState.enabled = args.enabled;
	}
}

// update the range/category toggle for a variable
export function setVarFilterRange(state, args) {
	const varState = state.filterState[args.name];	
	if (varState) {
		if (varState.type === index.NUMERICAL_SUMMARY_TYPE) {
			varState.min = args.min;
			varState.max = args.max;
		} else if (varState.type === index.CATEGORICAL_SUMMARY_TYPE) {
			varState.categories = args.categories;
		} else {
			console.error(`Unhandle category type ${varState.type}`);
		}
	}
}

// add/replace the  filter state for a variable
export function updateVarFilterState(state, args) {
	state.filterState[args.name] = args.filterState;
}

// replace the entire filter state 
export function setFilterState(state, args) {
	state.filterState = args;
}
