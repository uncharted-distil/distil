import _ from 'lodash';
import Connection from '../util/ws';
import { decodeFilters } from '../util/filters';

/**
 * ROUTE
 */

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

export function getRouteTrainingVariables(state) {
	return () => {
		return state.route.query.training ? state.route.query.training : null;
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

export function getRouteFilters(state) {
	return () => {
		return state.route.query.filters ? state.route.query.filters : [];
	};
}

export function getRouteResultFilters(state) {
	return () => {
		return state.route.query.results ? state.route.query.results : [];
	};
}

export function getRouteFacetsPage(state) {
	return (pageKey) => state.route.query[pageKey];
}

export function getRouteResidualThreshold(state) {
	return () => _.get(state.route.query, 'residualThreshold', 0.0);
}

export function getFilters(state) {
	return () => {
		return decodeFilters(state.route.query.filters ? state.route.query.filters : []);
	};
}

export function getResultsFilters(state) {
	return () => {
		return decodeFilters(state.route.query.results ? state.route.query.results : []);
	};
}

export function getVariables(state) {
	return () => state.variables;
}

export function getVariablesMap(state) {
	return () => {
		const map = {};
		state.variables.forEach(variable => {
			map[variable.name.toLowerCase()] = variable;
		});
		return map;
	};
}

export function getDatasets(state) {
	return (ids) => {
		if (_.isUndefined(ids)) {
			return state.datasets;
		}
		return _.intersectionWith(state.datasets, ids, (l, r) => l.name === r);
	};
}

export function getAvailableVariables(state, getters) {
	return () => {
		const training = getters.getTrainingVariablesMap();
		const target = getters.getTargetVariable();
		return state.variables.filter(variable => {
			return (!target || target.toLowerCase() !== variable.name.toLowerCase()) &&
				!training[variable.name.toLowerCase()];
		}).map(v => v.name);
	};
}

export function getAvailableVariablesMap(state, getters) {
	return () => {
		const available = getters.getAvailableVariables();
		const map = {};
		available.forEach(name => {
			map[name.toLowerCase()] = true;
		});
		return map;
	};
}

export function getTrainingVariables(state) {
	return () => {
		return state.route.query.training ? state.route.query.training.split(',') : [];
	};
}

export function getTrainingVariablesMap(state, getters) {
	return () => {
		const training = getters.getTrainingVariables();
		const map = {};
		training.forEach(name => {
			map[name.toLowerCase()] = true;
		});
		return map;
	};
}

export function getTargetVariable(state) {
	return () => {
		return state.route.query.target ? state.route.query.target : null;
	};
}

export function getVariableSummaries(state) {
	return () => state.variableSummaries;
}

export function getResultsSummaries(state) {
	return () => {
		return state.resultsSummaries;
	};
}

export function getSelectedFilters(state, getters) {
	return () => {
		const training = getters.getTrainingVariables();
		const filters = {};
		training.forEach(variable => {
			filters[variable] = {
				name: variable,
				enabled: false
			};
		});
		return filters;
	};
}

export function getAvailableVariableSummaries(state, getters) {
	return () => {
		const available = getters.getAvailableVariablesMap();
		return state.variableSummaries.filter(variable => {
			return available[variable.name.toLowerCase()];
		});
	};
}

export function getTrainingVariableSummaries(state, getters) {
	return () => {
		const training = getters.getTrainingVariablesMap();
		return state.variableSummaries.filter(variable => {
			return training[variable.name.toLowerCase()];
		});
	};
}

export function getTargetVariableSummaries(state, getters) {
	return () => {
		const target = getters.getTargetVariable();
		if (!target) {
			return [];
		}
		return state.variableSummaries.filter(variable => {
			return target.toLowerCase() === variable.name.toLowerCase();
		});
	};
}

export function getFilteredData(state) {
	return () => state.filteredData;
}

function validateData(data) {
	return  !_.isEmpty(data) &&
		!_.isEmpty(data.values) &&
		!_.isEmpty(data.columns);
}

export function getFilteredDataItems(state) {
	return () => {
		if (validateData(state.filteredData)) {
			return _.map(state.filteredData.values, d => {
				const row = {};
				for (const [index, col] of state.filteredData.columns.entries()) {
					row[col] = d[index];
				}
				_.forIn(state.highlightedFeatureRanges, (range, name) => {
					if (row[name] >= range.from && row[name] <= range.to) {
						row._rowVariant = 'info';
					}
				});
				return row;
			});
		}
		return [];
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
	return () => state.resultData;
}

export function getResultDataItems(state, getters) {
	return (computeResiduals) => {
		if (validateData(state.resultData)) {
			const resultDataItems = getters.getFilteredDataItems();
			// append the result variable data to the baseline variable data
			for (const [i, row] of resultDataItems.entries()) {
				for (const [j, colName] of state.resultData.columns.entries()) {
					// append the result value
					const label = `Predicted ${colName}`;
					row[label] = state.resultData.values[i][j];

					// append the residual value if necessary
					let residualLabel = null;
					if (computeResiduals) {
						residualLabel = 'Error';
						row[residualLabel] = row[colName] - state.resultData.values[i][j];
					}
					// save the names of the columns related to the target and predictions as metadata
					// for use at render time
					row._target = { truth: colName, predicted: label, error: residualLabel };
				}
			}
		}
		return [];
	};
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

export function getSelectedData(state) {
	return () => {
		return state.selectedData;
	};
}

export function getSelectedDataItems(state) {
	return () => {
		if (validateData(state.selectedData)) {
			return _.map(state.selectedData.values, d => {
				const row = {};
				for (const [index, col] of state.selectedData.columns.entries()) {
					row[col] = d[index];
				}
				_.forIn(state.highlightedFeatureRanges, (range, name) => {
					if (row[name] >= range.from && row[name] <= range.to) {
						row._rowVariant = 'info';
					}
				});
				return row;
			});
		}
		return [];
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
	return () => state.pipelineSession;
}

export function getHighlightedFeatureValues(state) {
	return () => state.highlightedFeatureValues;
}
