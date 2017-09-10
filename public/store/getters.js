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

export function getRouteResultFilters(state) {
	return () => {
		return state.route.query.results;
	};
}

export function getResultsSummaries(state) {
	return () => {
		return state.resultsSummaries;
	};
}

export function getRouteFilters(state) {
	return () => {
		return state.route.query.filters ? state.route.query.filters : [];
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

function validateData(data) {
	return  !_.isEmpty(data) && !_.isEmpty(data.values) && !_.isEmpty(data.metadata);
}

export function getFilteredDataItems(state) {
	return () => {
		const data = state.filteredData;
		if (validateData(data)) {
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

export function getResultDataItems(state, getters) {
	return () => {
		// get the filtered data items
		const dataRows = getters.getFilteredDataItems(state);
		if (validateData(state.resultData.results)) {
			// append the result variable data to the baseline variable data
			for (const [i, dataObj] of dataRows.entries()) {
				const resultData = state.resultData.results;
				for (const [j, resultMeta] of resultData.metadata.entries()) {
					const label = `Predicted ${resultMeta.name}`;
					dataObj[label] = resultData.values[i][j];
					if (dataObj[resultMeta.name] !== resultData.values[i][j]) {
						dataObj._cellVariants = { [label]: 'danger'};
					}
				}
			}
			return dataRows;
		} else {
			return [];
		}
	};
}

export function getResultDataFields(state, getters) {
	return () => {
		// const resultData = state.resultData;
		const dataFields = getters.getFilteredDataFields();
		const resultData = state.resultData.results;
		if (!_.isEmpty(resultData)) {
			for (let resultMeta of resultData.metadata) {
				const label = `Predicted ${resultMeta.name}`; 
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
