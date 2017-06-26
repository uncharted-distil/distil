import Vue from 'vue';
import Vuex from 'vuex';
import * as actions from './actions';
import * as getters from './getters';
import * as mutations from './mutations';

Vue.use(Vuex);

export const NUMERICAL_SUMMARY_TYPE = 'numerical';
export const CATEGORICAL_SUMMARY_TYPE = 'categorical';

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
	variableSummaries: [
		//     name: '',
		//     buckets: [{
		//         key: '',
		//         count: 0
		//     }]
	],
	// filtered data entries for the active dataset
	filteredData: {
		// name: '',
		// metadata: [
		//     {
		//         name: '',
		//         type: ''
		//     }
		// ]
		// values: [
		//     []
		// ]
	},
	filterState: {
		// On_base_pct: {
		//     type: 'numerical',
		//     enabled: true,
		//     min: '10',
		//     max: '100',
		// },
		// Position: {
		//     type: 'categorical',
		//     enabled: false,
		//     categories: ['pitcher', 'catcher']
		// },
		// ...
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
