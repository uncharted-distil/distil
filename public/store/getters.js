import _ from 'lodash';
import Connection from '../util/ws';

export function getRoute(state) {
	return () => state.route;
}

export function getRoutePath(state) {
	return () => {
		return state.route.path;
	};
}

export function getRouteTerms(state) {
	return () => state.route.query.terms;
}

export function getRouteDataset(state) {
	return () => state.route.query.dataset;
}

export function getRouteFilter(state) {
	return (varName) => {
		return varName in state.route.query ? state.route.query[varName] : null;
	};
}

export function getRouteTrainingVariables(state) {
	return () => {
		return state.route.query.training ? state.route.query.training : null;
	};
}
export function getRouteTrainingVariablesArray(state) {
	return () => {
		return state.route.query.training ? state.route.query.training.split(',') : [];
	};
}

export function getRouteTrainingVariablesMap(state, getters) {
	return () => {
		const training = getters.getRouteTrainingVariablesArray();
		const map = {};
		training.forEach(variable => {
			map[variable.toLowerCase()] = true;
		});
		return map;
	};
}

export function getRouteTargetVariable(state) {
	return () => {
		return state.route.query.target ? state.route.query.target : null;
	};
}

export function getRouteCreateRequestId(state) {
	return () => state.route.query.createRequestId;
}

export function getRouteResultId(state) {
	return () => state.route.query.resultId;
}

export function getRouteResultFilters(state) {
	return () => {
		return state.route.query.results;
	};
}

export function getRouteFilters(state) {
	return () => {
		return state.route.query.filters ? state.route.query.filters : [];
	};
}

export function getRouteFacetsPage(state) {
	return (pageKey) => {
		return state.route.query[pageKey];
	};
}

export function getResultsSummaries(state) {
	return () => {
		return state.resultsSummaries;
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

export function getAvailableVariables(state, getters) {
	return () => {
		const training = getters.getRouteTrainingVariablesMap();
		const target = getters.getRouteTargetVariable() || '';
		return state.variableSummaries.filter(variable => {
			return target.toLowerCase() !== variable.name.toLowerCase() &&
				!training[variable.name.toLowerCase()];
		});
	};
}

export function getTrainingVariables(state, getters) {
	return () => {
		const training = getters.getRouteTrainingVariablesMap();
		const target = getters.getRouteTargetVariable() || '';
		return state.variableSummaries.filter(variable => {
			return target.toLowerCase() !== variable.name.toLowerCase() &&
				training[variable.name.toLowerCase()];
		});
	};
}

export function getTargetVariable(state, getters) {
	return () => {
		const target = getters.getRouteTargetVariable();
		if (!target) {
			return null;
		}
		return state.variableSummaries.filter(variable => {
			return target.toLowerCase() === variable.name.toLowerCase();
		})[0];
	};
}

export function getFilteredData(state) {
	return () => state.filteredData;
}

export function getFilteredDataItems(state) {
	return () => {
		return state.filteredDataItems;
	};
}

export function getFilteredDataFields(state) {
	return () => {
		const data = state.filteredData;
		if (!_.isEmpty(data)) {
			const result = {};
			for (const col of data.columns) {
				result[col] = {
					label: col,
					sortable: true
				};
			}
			return result;
		} else {
			return {};
		}
	};
}

export function getResultData(state) {
	return () => {
		return state.resultData;
	};
}

export function getResultDataItems(state) {
	return () => {
		return state.resultDataItems;
	};
}

export function getResultDataFields(state, getters) {
	return () => {
		// const resultData = state.resultData;
		const dataFields = getters.getFilteredDataFields();
		const resultData = state.resultData;
		if (!_.isEmpty(resultData)) {
			for (const col of resultData.columns) {
				const label = `Predicted ${col}`;
				dataFields[label] = {
					label: label,
					sortable: true
				};
			}
			return dataFields;
		} else {
			return {};
		}
	};
}


export function getSelectedData(state) {
	return () => {
		return state.selectedData;
	};
}

export function getSelectedDataItems(state) {
	return () => {
		return state.selectedDataItems;
	};
}

export function getSelectedDataFields(state) {
	return () => {
		const data = state.selectedData;
		if (!_.isEmpty(data)) {
			const result = {};
			for (const col of data.columns) {
				result[col] = {
					label: col,
					sortable: true
				};
			}
			return result;
		} else {
			return {};
		}
	};
}

export function getPipelineResults(state) {
	return (requestId) => {
		return _.concat(
			_.values(state.runningPipelines[requestId]),
			_.values(state.completedPipelines[requestId]));
	};
}

export function getWebSocketConnection() {
	const conn = new Connection('/ws', err => {
		if (err) {
			console.warn(err);
			return;
		}
	});
	return () => {
		return conn;
	};
}

export function getPipelineSessionID(state) {
	return () => {
		if (!state.pipelineSession) {
			return window.localStorage.getItem('pipeline-session-id');
		}
		return state.pipelineSession.id;
	};
}

export function getPipelineSession(state) {
	return () => {
		return state.pipelineSession;
	};
}

export function getHighlightedFeatureValues(state) {
	return () => {
		return state.highlightedFeatureValues;
	};
}
