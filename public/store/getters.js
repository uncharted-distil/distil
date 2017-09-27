import _ from 'lodash';
import Connection from '../util/ws';

export function getRoute(state) {
	return () => state.route;
}

export function getRoutePath(state) {
	return () => state.route.path;
}

export function getRouteTerms(state) {
	return () => state.route.query.terms;
}

export function getRouteDataset(state) {
	return () => state.route.query.dataset;
}

export function getRouteFilter(state) {
	return (varName) => _.get(state.route.query, varName, null);
}

export function getRouteTrainingVariables(state) {
	return () => state.route.query.training ? state.route.query.training.split(',') : [];
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
	return () => state.route.query.target ? state.route.query.target.toLowerCase(): null;
}

export function getRouteCreateRequestId(state) {
	return () => state.route.query.createRequestId;
}

export function getRouteResultId(state) {
	return () => state.route.query.resultId;
}

export function getRouteResultFilters(state) {
	return () => state.route.query.results;
}

export function getRouteFilters(state) {
	return () => _.get(state.route.query, 'filters', []);
}

export function getRouteFacetsPage(state) {
	return (pageKey) => state.route.query[pageKey];
}

export function getRouteResidualThreshold(state) {
	return () => _.get(state.route.query, 'residualThreshold', 0.0);
}

export function getResultsSummaries(state) {
	return () => state.resultsSummaries;
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
	return () => state.filteredDataItems;
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
	return () => state.resultData;
}

export function getResultDataItems(state) {
	return () => state.resultDataItems;
}

export function getResultDataFields(state, getters) {
	return (regression) => {
		let dataFields = getters.getFilteredDataFields();

		// target field should be last displayed in table, next to predicted value and error
		// (if applicable)
		const resultData = state.resultData;
		if (!_.isEmpty(resultData)) {
			// add the result data to the baseline data
			for (const col of resultData.columns) {
				const truthValue = dataFields[col];
				dataFields = _.omit(dataFields, col);
				dataFields[col] = truthValue;

				const label = `Predicted ${col}`;
				dataFields[label] = {
					label: label,
					sortable: true
				};
				// add a field for the residuals for numeric predictions
				if (regression) {
					const errorLabel = 'Error';
					dataFields[errorLabel] = {
						label: errorLabel,
						sortable: true
					};
				}
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
	return () => state.pipelineSession;
}

export function getHighlightedFeatureValues(state) {
	return () => state.highlightedFeatureValues;
}
