import _ from 'lodash';
import axios from 'axios';

// searches dataset descriptions and column names for supplied terms
export function searchDatasets(context, terms) {
	if (_.isEmpty(terms)) {
		context.commit('setDatasets', []);
	} else {
		axios.get('/distil/datasets?search=' + terms)
		.then(response => {
			if (!_.isEmpty(response.data.datasets)) {
				context.commit('setDatasets', response.data.datasets);
			} else {
				context.commit('setDatasets', []);
			}
		})
		.catch(error => {
			console.log(error);
			context.commit('setDatasets', []);
		});
	}
}

// fetches variable summary data for the given dataset
export function getVariableSummaries(context, name) {
	axios.get('/distil/variable-summaries/' + name)
		.then(response => {
			context.commit('setVariableSummaries', response.data);
		})
		.catch(error => {
			console.log(error);
			context.commit('setVariableSummaries', {});
		});
}

// fetches data entries for the given dataset
export function getData(context, name) {
	axios.get('/distil/data/' + name)
		.then(response => {
			context.commit('setData', response.data);
		})
		.catch(error => {
			console.log(error);
			context.commit('setData', {});
		});
}
