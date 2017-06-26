import _ from 'lodash';
import axios from 'axios';
import * as index from './index';

// TODO: move this somewhere more appropriate.
const ES_INDEX = 'datasets';
const TABLE_DATA_LIMIT = 100;

// searches dataset descriptions and column names for supplied terms
export function searchDatasets(context, terms) {
	if (_.isEmpty(terms)) {
		context.commit('setDatasets', []);
	} else {
		axios.get(`/distil/datasets/${ES_INDEX}?search=${terms}`)
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
}

// fetches variable summary data for the given dataset and variables
export function getVariableSummaries(context, dataset) {
	// commit empty place holders
	const histograms = new Array(dataset.variables.length - 1);
	dataset.variables.forEach((variable, idx) => {
		histograms[idx] = {
			name: variable.name,
			pending: true
		};
	});
	context.commit('setVariableSummaries', histograms);
	// fill them in asynchronously
	dataset.variables.forEach((variable, idx) => {
		axios.get(`/distil/variable-summaries/${ES_INDEX}/${dataset.name}/${variable.name}`)
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
}

// update filtered data based on the  current filter state
export function updateFilteredData(context, datasetName) {
	// build up a map of var types so we can quickly look them up while we generate parameters
	// TODO: this should really be availabe through the store in some convenient fashion
	const variables = context.getters.getDataset(datasetName).variables;
	const varTypes = new Map();
	for (let variable of variables) {
		varTypes.set(variable.name, variable.type);
	}

	// initialize the url
	const filterState = context.state.filterState;
	var requestUrl = `distil/filtered-data/${datasetName}?`;

	// build up the parameter list from the current filter state
	var params = [];
	params.push(`size=${TABLE_DATA_LIMIT}`);
	_.forEach(filterState, varFilter => {
		if (varFilter.enabled) {			
			// numeric types have type,min,max or no additonal args if the value is unfiltered
			if (varFilter.type === index.NUMERICAL_SUMMARY_TYPE) {
				if (!_.isEmpty(varFilter, 'min') && !_.isEmpty(varFilter, 'max')) {
					params.push(varFilter.name + '=' + [encodeURIComponent(varTypes.get(varFilter.name)), varFilter.min, varFilter.max].join(','));
				} else {
					params.push(varFilter.name);
				}
			// categorical type shave type,cat1,cat2...catN or no additional args if the value is unfiltered
			} else if (varFilter.type === index.CATEGORICAL_SUMMARY_TYPE) {
				if (!_.isEmpty(varFilter.categories)) {
					var varParams = encodeURIComponent(varTypes.get(varFilter.name));
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
}
