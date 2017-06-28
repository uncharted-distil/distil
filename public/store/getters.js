import _ from 'lodash';

export function getSearchTerms(state) {
	return () => state.searchTerms;
}

export function getVariables(state) {
	return () => state.variables;
}

export function getVariablesMap(state) {
	return () => {
		const map = new Map();
		state.variables.forEach(variable => {
			map.set(variable.name, variable.type);
		});
		return map;
	};
}

export function getDataset(state) {
	return (id) => state.datasets.find(d => d.name === id);
}

export function getActiveDataset(state) {
	return () => state.activeDataset;
}

export function getDatasets(state) {
	return (ids) => {
		if (_.isUndefined) {
			return state.datasets;
		}
		return _.intersectionWith(state.datasets, ids, (l, r) => l.name === r);
	};
}

export function getVariableSummaries(state) {
	return () => state.variableSummaries;
}

export function getFilteredData(state) {
	return () => state.filteredData;
}
