import _ from 'lodash';
import axios from 'axios';
import * as index from './index';

// TODO: move this somewhere more appropriate.
const ES_INDEX = 'datasets';
const TABLE_DATA_LIMIT = 100;

export function getVariables(context, dataset) {
	return axios.get(`/distil/variables/${ES_INDEX}/${dataset}`)
		.then(response => {
			if (!_.isEmpty(response.data.variables)) {
				context.commit('setVariables', response.data.variables);
			} else {
				context.commit('setVariables', []);
			}
		})
		.catch(error => {
			console.error(error);
			context.commit('setVariables', []);
		});
}

// searches dataset descriptions and column names for supplied terms
export function searchDatasets(context, terms) {
	const params = (terms !== '') ? `?search=${terms}` : '';
	return axios.get(`/distil/datasets/${ES_INDEX}${params}`)
		.then(response => {
			if (!_.isEmpty(response.data.datasets)) {
				context.commit('setDatasets', response.data.datasets);
			} else {
				context.commit('setDatasets', []);
			}
		})
		.catch(error => {
			console.error(error);
			context.commit('setDatasets', []);
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
						const histogram = response.data.histograms[0];
						context.commit('updateVariableSummaries', {
							index: idx,
							histogram: histogram
						});
						// set the default filter state for the variable
						const filterState = {
							type: histogram.type,
							enabled: true
						};
						if (_.has(histogram, 'extrema')) {
							filterState.min = histogram.extrema.min;
							filterState.max = histogram.extrema.max;
						} else if (_.has(histogram, 'categories')) {
							filterState.categories = histogram.categories;
						}
						context.commit('updateVarFilterState', { name: histogram.name, filterState: filterState });
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
		.catch(() => {
			context.commit('setVariableSummaries', []);
		});
}

// update filtered data based on the  current filter state
export function updateFilteredData(context, datasetName) {

	return context.dispatch('getVariables', datasetName)
		.then(() => {
			// get variables map
			const variables = context.getters.getVariablesMap();

			// initialize the url
			const filterState = context.state.filterState;
			let requestUrl = `distil/filtered-data/${datasetName}?`;

			// build up the parameter list from the current filter state
			const params = [];
			params.push(`size=${TABLE_DATA_LIMIT}`);
			_.forEach(filterState, varFilter => {
				if (varFilter.enabled) {
					// numeric types have type,min,max or no additonal args if the value is unfiltered
					if (varFilter.type === index.NUMERICAL_SUMMARY_TYPE) {
						if (!_.isEmpty(varFilter, 'min') && !_.isEmpty(varFilter, 'max')) {
							params.push(varFilter.name + '=' + [encodeURIComponent(variables.get(varFilter.name)), varFilter.min, varFilter.max].join(','));
						} else {
							params.push(varFilter.name);
						}
					// categorical type shave type,cat1,cat2...catN or no additional args if the value is unfiltered
					} else if (varFilter.type === index.CATEGORICAL_SUMMARY_TYPE) {
						if (!_.isEmpty(varFilter.categories)) {
							let varParams = encodeURIComponent(variables.get(varFilter.name));
							varParams = ([varParams].concat(varFilter.categories)).join(',');
							params.push(encodeURIComponent(varFilter.name) + '=' + varParams);
						} else {
							params.push(varFilter.name);
						}
					}
				}
			});

			// construct the final URL
			requestUrl += params.join('&');

			// request filtered data from server - no data is valid given filter settings
			axios.get(requestUrl)
				.then(response => {
					if (_.isEmpty(response.data.metadata)) {
						context.commit('setFilteredData', {});
					} else {
						context.commit('setFilteredData', response.data);
					}
				})
				.catch(error => {
					console.error(error);
					context.commit('setFilteredData', {});
				});
		})
		.catch(() => {
			context.commit('setFilteredData', {});
		});
}
