import Vue from 'vue';
import Vuex from 'vuex';
import * as actions from './actions';
import * as getters from './getters';
import * as mutations from './mutations';

Vue.use(Vuex);

// shared data model
const state = {
	// description of matched datasets
	datasets: [
		// {
		//     name: '',
		//     description: '',
		//     variables: [
		//         {
		//             name: '',
		//            type: ''
		//         }
		//     ]
		// }
	],
	// variable summary data for the active dataset
	variableSummaries: {
		// histograms: [{
		//     name: '',
		//     buckets: [{
		//         key: '',
		//         count: 0
		//     }]
		// }]
	},
	// data entries for the active dataset
	data: {
		// name: '',
		// variables: [
		//     {
		//         name: '',
		//         type: '',
		//     }
		// ]
	},
	// name/id of the active dataset
	activeDataset: null
};

export default new Vuex.Store({
	state,
	getters,
	actions,
	mutations,
	strict: true
});
