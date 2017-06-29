import _ from 'lodash';

export function getRouteTerms(state) {
	return () => state.route.query.terms;
}

export function getRouteDataset(state) {
	return () => state.route.query.dataset;
}

export function getRouteFilter(state) {
	return (varName) => {
		return state.route.query[varName] !== undefined ? `${varName}=${state.route.query[varName]}` : null;
	};
}

export function getRouteFilters(state) {
	return () => {
		const filters = [];
		_.forIn(state.route.query, (value, key) => {
			if (key !== 'dataset' && key !== 'terms') {
				filters.push(`${key}=${value}`);
			}
		});
		return filters;
	};
}

export function getVariables(state) {
	return () => state.variables;
}

export function getDatasets(state) {
	return (ids) => {
		if (_.isUndefined(ids)) {
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

export function getFilteredDataItems(state) {
	return () => {
		const data = state.filteredData;
		if (!_.isEmpty(data)) {
			return _.map(data.values, d => {
				const rowObj = {};
				for (const [idx, varMeta] of data.metadata.entries()) {
					rowObj[varMeta.name] = d[idx];
				}
				return rowObj;
			});
		} else {
			return [];
		}
	};
}

export function getFilteredDataFields(state) {
	return () => {
		const data = state.filteredData;
		if (!_.isEmpty(data)) {
			const result = {};
			for (let varMeta of data.metadata) {
				result[varMeta.name] = {
					label: varMeta.name
				};
			}
			return result;
		} else {
			return {};
		}
	};
}
