import _ from 'lodash';
import axios from 'axios';

// TODO: move this somewhere more appropriate.
const ES_INDEX = 'datasets';

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
		if (index > 20) {
			return;
		}
		histograms[index] = {
			pending: true,
			histogram: {
				name: variable.name
			}
		};
	});
	context.commit('setVariableSummaries', histograms);
	// fill them in asynchronously
	/*
	dataset.variables.forEach((variable, index) => {
		if (index > 20) {
			return;
		}
		axios.get(`/distil/variable-summaries/${ES_INDEX}/${dataset.name}/${variable.name}`)
			.then(response => {
				context.commit('updateVariableSummaries', {
					index: index,
					summary: {
						histogram: response.data.histograms[0]
					}
				});
			})
			.catch(error => {
				console.error(error);
				context.commit('updateVariableSummaries', {
					index: index,
					summary: {
						err: new Error(error)
					}
				});
			});
	});
	*/
}

// fetches data entries for the given dataset
export function getData(context, datasetName) {
	axios.get(`/distil/data/${ES_INDEX}/${datasetName}`)
		.then(response => {
			context.commit('setData', response.data);
		})
		.catch(error => {
			console.error(error);
			context.commit('setData', {});
		});
}
