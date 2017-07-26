import _ from 'lodash';
import axios from 'axios';
import { decodeFilters, encodeQueryParams } from '../util/filters';

// TODO: move this somewhere more appropriate.
const ES_INDEX = 'datasets';

// searches dataset descriptions and column names for supplied terms
export function searchDatasets(context, terms) {
	const params = !_.isEmpty(terms) ? `?search=${terms}` : '';
	return axios.get(`/distil/datasets/${ES_INDEX}${params}`)
		.then(response => {
			context.commit('setDatasets', response.data.datasets);
		})
		.catch(error => {
			console.error(error);
			context.commit('setDatasets', []);
		});
}

// fetches all variables for a single dataset.
export function getVariables(context, dataset) {
	return axios.get(`/distil/variables/${ES_INDEX}/${dataset}`)
		.then(response => {
			context.commit('setVariables', response.data.variables);
		})
		.catch(error => {
			console.error(error);
			context.commit('setVariables', []);
		});
}

// fetches variable summary data for the given dataset and variables
export function getVariableSummaries(context, datasetName) {
	return context.dispatch('getVariables', datasetName)
		.then(() => {
			const variables = context.getters.getVariables();
			// commit empty place holders
			const histograms = variables.map(variable => {
				return {
					name: variable.name,
					pending: true
				};
			});
			context.commit('setVariableSummaries', histograms);
			// fill them in asynchronously
			variables.forEach((variable, idx) => {
				axios.get(`/distil/variable-summaries/${ES_INDEX}/${datasetName}/${variable.name}`)
					.then(response => {
						// save the variable summary data
						const histogram = response.data.histogram;
						context.commit('updateVariableSummaries', {
							index: idx,
							histogram: histogram
						});
					})
					.catch(error => {
						console.error(error);
						context.commit('updateVariableSummaries', {
							index: idx,
							histogram: {
								name: variable.name,
								err: error
							}
						});
					});
			});
		})
		.catch(error => {
			console.error(error);
		});
}

// update filtered data based on the  current filter state
export function updateFilteredData(context, datasetName) {
	const filters = context.getters.getRouteFilters();
	const decoded = decodeFilters(filters);
	const queryParams = encodeQueryParams(decoded);
	const url = `distil/filtered-data/${datasetName}${queryParams}`;
	// request filtered data from server - no data is valid given filter settings
	return axios.get(url)
		.then(response => {
			context.commit('setFilteredData', response.data);
		})
		.catch(error => {
			console.error(error);
				context.commit('setFilteredData', []);
		});
}

// starts a pipeline session.
export function getPipelineSession(context) {
	const conn = context.getters.getWebSocketConnection();
	const sessionID = context.getters.getPipelineSessionID();
	return conn.send({
			type: 'GET_SESSION',
			session: sessionID
		}).then(res => {
			if (sessionID && res.created) {
				console.warn('previous session', sessionID, 'could not be resumed, new session created');
			}
			context.commit('setPipelineSession', {
				id: res.session,
				uuids: res.uuids
			});
		}).catch(err => {
			console.warn(err);
		});
}

// end a pipeline session.
export function endPipelineSession(context) {
	const conn = context.getters.getWebSocketConnection();
	const sessionID = context.getters.getPipelineSessionID();
	if (!sessionID) {
		return;
	}
	return conn.send({
			type: 'END_SESSION',
			session: sessionID
		}).then(() => {
			context.commit('setPipelineSession', null);
		}).catch(err => {
			console.warn(err);
		});
}
