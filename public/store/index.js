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
	// variable list for the active dataset
	variables: [
		// {
		//     name: '',
		//     type: ''
		// }
	],
	// variable summary data for the active dataset
	variableSummaries: [
		// {
		//     name: '',
		//     buckets: [
		//     {
		//             key: '',
		//             count: 0
		//         }
		//     ]
		// }
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
	// the underlying websocket connection
	wsConnection: null,
	// the pipeline session id
	pipelineSession: null
};

export default new Vuex.Store({
	state,
	getters,
	actions,
	mutations,
	strict: true
});
