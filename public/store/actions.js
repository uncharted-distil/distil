import _ from 'lodash';
import axios from 'axios';

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
	dataset.variables.forEach((variable, index) => {
		histograms[index] = {
				name: variable.name,
				pending: true
		};
	});
	context.commit('setVariableSummaries', histograms);
	// fill them in asynchronously
	dataset.variables.forEach((variable, index) => {
		axios.get(`/distil/variable-summaries/${ES_INDEX}/${dataset.name}/${variable.name}`)
			.then(response => {
				context.commit('updateVariableSummaries', {
					index: index,
					histogram: response.data.histograms[0]

				});
			})
			.catch(error => {
				console.error(error);
				context.commit('updateVariableSummaries', {
					index: index,
					histogram: {
						name: variable.name,
						err: error
					}

				});
			});
	});
}

// fetches data entries for the given dataset
export function getFilteredData(context, name) {
	// should be updated to take params based on facet state, but will just get all
	// data for now
	axios.get(`/distil/filtered-data/${name}?size=${TABLE_DATA_LIMIT}`)
		.then(response => {
			context.commit('setFilteredData', response.data);
		})
		.catch(error => {
			console.error(error);
			context.commit('setFilteredData', {});
		});
}
