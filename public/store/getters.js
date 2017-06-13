import _ from 'lodash';

export function getVariables(state) {
	return (id) => {
		const dataset = state.datasets.find(d => d.name === id);
		if (dataset) {
			return dataset.variables;
		}
		return null;
	};
}

export function getDataset(state) {
	return (id) => state.datasets.find(d => d.name === id);
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
