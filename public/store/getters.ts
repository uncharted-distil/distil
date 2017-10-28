import _ from 'lodash';
import Connection from '../util/ws';
import { decodeFilters, FilterMap } from '../util/filters';
import { GetterTree } from 'vuex';
import { DistilState, Variable, Data } from './index';

function getTargetIndexFromPredicted(columns: string[], predictedIndex: number) {
	const targetName = columns[predictedIndex].replace('_res', '');
	return _.findIndex(columns, col => col === targetName);
}

function getPredictedIndex(columns: string[]) {
	return _.findIndex(columns, col => col.endsWith('_res'));
}

function getErrorIndex(columns: string[]) {
	return _.findIndex(columns, col => col === 'error');
}

function validateData(data: Data) {
	return  !_.isEmpty(data) &&
		!_.isEmpty(data.values) &&
		!_.isEmpty(data.columns);
}

export interface FieldInfo {
	label: string,
	type: string,
	suggested: string[],
	sortable: boolean
}

export const getters: GetterTree<DistilState, any> = {
	getRoute(state) {
		return () => state.route;
	},

	getRoutePath(state) {
		return () => state.route.path;
	},

	getRouteTerms(state) {
		return () => state.route.query.terms;
	},

	getRouteDataset(state) {
		return () => state.route.query.dataset;
	},

	getRouteTrainingVariables(state) {
		return () => state.route.query.training ? state.route.query.training : null;
	},

	getRouteTargetVariable(state) {
		return () => state.route.query.target ? state.route.query.target : null;
	},

	getRouteCreateRequestId(state) {
		return () => state.route.query.createRequestId;
	},

	getRouteResultId(state) {
		return () => state.route.query.resultId;
	},

	getRouteFilters(state) {
		return () => state.route.query.filters ? state.route.query.filters : [];
	},

	getRouteResultFilters(state) {
		return () => state.route.query.results ? state.route.query.results : [];
	},

	getRouteFacetsPage(state) {
		return (pageKey: string) => state.route.query[pageKey];
	},

	getRouteResidualThreshold(state) {
		return () => state.route.query.residualThreshold;
	},

	getFilters(state) {
		return () => decodeFilters(state.route.query.filters ? state.route.query.filters : "") as FilterMap;
	},

	getResultsFilters(state) {
		return decodeFilters(state.route.query.results ? state.route.query.results : "") as FilterMap;
	},

	getVariables(state) {
		return () => state.variables;
	},

	getVariablesMap(state) {
		return () => {
			const map: { [name: string]: Variable } = {};
			state.variables.forEach(variable => {
				map[variable.name.toLowerCase()] = variable;
			});
			return map;
		};
	},

	getDatasets(state) {
		return (ids: string[]) => {
			if (_.isUndefined(ids)) {
				return state.datasets;
			}
			const idSet = new Set(ids);
			return _.filter(state.datasets, d => idSet.has(d.name));
		};
	},

	getAvailableVariables(state, getters) {
		return () => {
			const training = getters.getTrainingVariablesMap();
			const target = getters.getTargetVariable();
			return state.variables.filter(variable => {
				return (!target || target.toLowerCase() !== variable.name.toLowerCase()) &&
					!training[variable.name.toLowerCase()];
			}).map(v => v.name);
		};
	},

	getAvailableVariablesMap(state, getters) {
		return () => {
			const available = getters.getAvailableVariables() as string[];
			const map: { [name: string]: boolean } = {};
			available.forEach(name => {
				map[name.toLowerCase()] = true;
			});
			return map;
		};
	},

	getTrainingVariables(state) {
		return () => state.route.query.training ? state.route.query.training.split(',') : [];
	},

	getTrainingVariablesMap(state, getters) {
		return () => {
			const training = getters.getTrainingVariables() as string[];
			const map: { [name: string]: boolean } = {};
			training.forEach(name => {
				map[name.toLowerCase()] = true;
			});
			return map;
		};
	},

	getTargetVariable(state) {
		return () => {
			return state.route.query.target ? state.route.query.target : null;
		};
	},

	getVariableSummaries(state) {
		return () => state.variableSummaries;
	},

	getResultsSummaries(state) {
		return () => {
			return state.resultsSummaries;
		};
	},

	getSelectedFilters(state, getters) {
		return () => {
			const training = getters.getTrainingVariables() as string[];
			const existing = getters.getFilters();
			const filters: { [name: string]: { name: string, enabled: boolean } } = {};

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
	},

	getAvailableVariableSummaries(state, getters) {
		return () => {
			const available = getters.getAvailableVariablesMap();
			return state.variableSummaries.filter(variable => {
				return available[variable.name.toLowerCase()];
			});
		};
	},

	getTrainingVariableSummaries(state, getters) {
		return () => {
			const training = getters.getTrainingVariablesMap();
			return state.variableSummaries.filter(variable => {
				return training[variable.name.toLowerCase()];
			});
		};
	},

	getTargetVariableSummaries(state, getters) {
		return () => {
			const target = getters.getTargetVariable();
			if (!target) {
				return [];
			}
			return state.variableSummaries.filter(variable => {
				return target.toLowerCase() === variable.name.toLowerCase();
			});
		};
	},

	getFilteredData(state) {
		return () => state.filteredData;
	},

	getFilteredDataItems(state) {
		return () => {
			if (validateData(state.filteredData)) {
				return _.map(state.filteredData.values, d => {
					const row: { [col: string]: any } = {};
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
	},

	getFilteredDataFields(state) {
		return () => {
			const data = state.filteredData;

			const variables = state.variables;
			const types = {};
			const suggested = {};
			variables.forEach(variable => {
			  types[variable.name] = variable.type;
			  suggested[variable.name] = variable.suggestedTypes;
			});

			if (!_.isEmpty(data)) {
				const result: { [col: string]: FieldInfo } = {};
				for (const col of data.columns) {
					result[col] = {
						label: col,
						type: types[col],
						suggested: suggested[col],
						sortable: true
					};
				}
				return result;
			} else {
				return {};
			}
		};
	},

	getResultData(state) {
		return () => state.resultData;
	},

	getResultDataItems(state) {
		return () => {
			const resultData = state.resultData;
			if (validateData(resultData)) {

				// look at first row and figure out the target, predicted, error values
				const predictedIdx = getPredictedIndex(resultData.columns);
				const targetName = resultData.columns[getTargetIndexFromPredicted(resultData.columns, predictedIdx)];
				const errorIdx = getErrorIndex(resultData.columns);

				// convert fetched result data rows into table data rows
				return _.map(resultData.values, resultRow => {
					const row: { [col: string]: any } = {};

					for (const [idx, colValues] of resultRow.entries()) {
						const colName = resultData.columns[idx];
						row[colName] = colValues;
					}
					row._target = { truth: targetName, predicted: resultData.columns[predictedIdx] };
					if (errorIdx >= 0) {
						row._target.error = resultData.columns[errorIdx];
					}
					// if row is in the current highlght range, set its style to info
					// TODO: this shouldn't be in the getter because it causes the entire
					// function to re-run whenever the high changes
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
	},

	getResultDataFields(state) {
		return () => {
			const data = state.resultData;

			// look at first row and figure out the target, predicted, error values
			const predictedIndex = getPredictedIndex(data.columns);
			const targetIndex = getTargetIndexFromPredicted(data.columns, predictedIndex);
			const errorIndex = getErrorIndex(data.columns);

			if (!_.isEmpty(data)) {
				const result: { [label: string]: {label: string, sortable: boolean} } = {};
				// assign column names, ignoring target, predicted and error
				for (const [idx, col] of data.columns.entries()) {
					if (idx !== predictedIndex && idx !== targetIndex && idx !== errorIndex) {
						result[col] = { label: col, sortable: true };
					}
				}
				// add target, predicted and error at end with customized labels
				const targetName = data.columns[targetIndex];
				result[targetName] = { label: targetName, sortable: true };
				result[data.columns[predictedIndex]] = { label: `Predicted ${targetName}`, sortable: true };
				if (errorIndex >= 0) {
					result[data.columns[errorIndex]] = { label: 'Error', sortable: true };
				}

				return result;
			} else {
				return {};
			}
		};
	},

	getSelectedData(state) {
		return () => {
			return state.selectedData;
		};
	},

	getSelectedDataItems(state) {
		return () => {
			if (validateData(state.selectedData)) {
				return _.map(state.selectedData.values, d => {
					const row: { [col: string]: any } = {};
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
	},

	getSelectedDataFields(state) {
		return () => {
			const data = state.selectedData;
			const variables = state.variables;
			const types = {};
			const suggested = {};
			variables.forEach(variable => {
			  types[variable.name] = variable.type;
			  suggested[variable.name] = variable.suggestedTypes;
			});

			if (!_.isEmpty(data)) {
				const result: { [label: string]: FieldInfo } = {};
				for (const col of data.columns) {
					result[col] = {
						label: col,
						type: types[col],
						suggested: suggested[col],
						sortable: true
					};
				}
				return result;
			} else {
				return {};
			}
		};
	},

	getPipelineResults(state) {
		return (requestId: string) => {
			return _.concat(
				_.values(state.runningPipelines[requestId]),
				_.values(state.completedPipelines[requestId]));
		};
	},

	getRunningPipelines(state) {
		return () => state.runningPipelines;
	},

	getCompletedPipelines(state) {
		return () => state.completedPipelines;
	},

	getWebSocketConnection() {
		const conn = new Connection('/ws', (err: string) => {
			if (err) {
				console.warn(err);
				return;
			}
		});
		return () => {
			return conn;
		};
	},

	getPipelineSessionID(state) {
		return () => {
			if (!state.pipelineSession) {
				return window.localStorage.getItem('pipeline-session-id');
			}
			return state.pipelineSession.id;
		};
	},

	getPipelineSession(state) {
		return () => state.pipelineSession;
	},

	getHighlightedFeatureValues(state) {
		return () => state.highlightedFeatureValues;
	},

	getRecentDatasets() {
		return () => {
			const datasets = window.localStorage.getItem('recent-datasets');
			return (datasets) ? datasets.split(',') : [];
		};
	}
};

