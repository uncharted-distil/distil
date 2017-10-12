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
	return () => state.route.query.residualThreshold;
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
		const existing = getters.getFilters();
		const filters = {};
		training.forEach(variable => {
			if (!existing[variable]) {
				filters[variable] = {
					name: variable,
					enabled: false
				};
			} else {
				filters[variable] = existing[variable];
			}
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

export function getResultDataItems(state) {
	return () => {
		const resultData = state.resultData; 
		if (validateData(resultData)) {
			
			// look at first row and figure out the target, predicted, error values
			const predictedIdx = _.findIndex(resultData.columns, col => col.endsWith('_res'));
			const targetName = resultData.columns[predictedIdx].replace('_res', '');
			const errorIdx = _.findIndex(resultData.columns, col => col === 'error');
			
			// convert fetched result data rows into table data rows 
			return _.map(resultData.values, resultRow => {
				const row = {};

				for (const [idx, colValues] of resultRow.entries()) {
					const colName = resultData.columns[idx];
					row[colName] = colValues;
				}
				row._target = { truth: targetName, predicted: resultData.columns[predictedIdx] };
				if (errorIdx >= 0) {
					row._target.error = resultData.columns[errorIdx];
				}

				// _.forIn(state.highlightedFeatureRanges, (range, name) => {
				// 	let col = row[name];
				// 	if (!row[name]) {
				// 		// row does not contain name, we ASSUME this is because it is a
				// 		// predicted field
				// 		col = row[label];
				// 	}
				// 	if (col >= range.from && col <= range.to) {
				// 		row._rowVariant = 'info';
				// 	}
				// });
				
				// if row is in the current highlght range, set its style to info
				// _.forIn(state.highlightedFeatureRanges, (range, name) => {
				// 	if (row[name] >= range.from && row[name] <= range.to) {
				// 		row._rowVariant = 'info';
				// 	}
				// });
				return row;
			});
		}
		return [];
	};
}

export function getResultDataFields(state) {
	return (computeResiduals) => {
		const data = state.resultData;
		if (!_.isEmpty(data)) {
			const result = {};
			for (const col of data.columns) {
				result[col] = {
					label: col,
					sortable: true
				};
			}
			if (computeResiduals) {
				result.Error = {
					label: 'Error', 
					sortable: true
				};
			}
			return result;
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
