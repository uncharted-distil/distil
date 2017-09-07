import _ from 'lodash';
import Connection from '../util/ws';

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
		return state.route.query.training ? state.route.query.training.split(',') : [];
	};
}

export function getRouteTrainingVariablesMap(state, getters) {
	return () => {
		const training = getters.getRouteTrainingVariables();
		const map = {};
		training.forEach(variable => {
			map[variable.toLowerCase()] = true;
		});
		return map;
	};
}

export function getRouteTargetVariable(state) {
	return () => {
		return state.route.query.target ? state.route.query.target.toLowerCase(): null;
	};
}

export function getRouteFilters(state) {
	return () => {
		const result = {};
		_.forEach(state.route.query, (value, key) => {
			// TODO: this is awful and error prone
			if (key !== 'dataset' &&
				key !== 'terms' &&
				key !== 'training' &&
				key !== 'target' &&
				key !== 'createRequestId') {
				result[key] = value;
			}
		});
		return result;
	};
}

export function getRouteCreateRequestId(state) {
	return () => state.route.query.createRequestId;
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
		const target = getters.getRouteTargetVariable();
		return state.variableSummaries.filter(variable => {
			return target !== variable.name.toLowerCase() &&
				!training[variable.name.toLowerCase()];
		});
	};
}

export function getTrainingVariables(state, getters) {
	return () => {
		const training = getters.getRouteTrainingVariablesMap();
		const target = getters.getRouteTargetVariable();
		return state.variableSummaries.filter(variable => {
			return target !== variable.name.toLowerCase() &&
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
			return target === variable.name.toLowerCase();
		})[0];
	};
}

export function getFilteredData(state) {
	return () => state.filteredData;
}

export function getFilteredDataItems(state) {
	return () => {
		const data = state.filteredData;
		if (!_.isEmpty(data) &&
			!_.isEmpty(data.values) &&
			!_.isEmpty(data.metadata)) {
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
					label: varMeta.name,
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

export function getPipelineSessionUUIDs(state) {
	return () => {
		return (state.pipelineSession && state.pipelineSession.uuids)
			? state.pipelineSession.uuids
			: [];
	};
}
