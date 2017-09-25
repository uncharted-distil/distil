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
	// results summary data for the selected pipeline run
	resultsSummaries: [
		// {
		//     name: '',
		// 	   pipelineId: '',
		//     buckets: [
		//     	   {
		//             key: '',
		//             count: 0
		//         }
		//     ]
		// }
	],
	// current set of pipeline results
	resultData: {
		// name: '',
		// columns:'[]',
		// types: '[]'
		// values: [
		//     []
		// ]
	},
	// result data items for the table view
	resultDataItems: [],
	// filtered data entries for the active dataset
	filteredData: {
		// name: '',
		// columns: [
		// types: '[]'
		// vaues: [
		//     []
		// ]
	},
	// filtered data items for the table view
	filteredDataItems: [],
	// running pipeline creation tasks grouped by parent create requestID
	runningPipelines: {
		// requestId: {
		//     pipelineId: {
		//         name: '',
		//         id: '',
		//         pipelineId: '',
		//         progress: '',
		//         pipeline: { // only present if progress === UPDATED,
		//             output: '',
		//             scores: [
		//                {
		//                    metric: '',
		//                    value: 0.1
		//                }
		//             ],
		//             resultUri: ''
		//         }
		//     }
		// }
	},
	// completed pipeline creation tasks grouped by parent create request ID
	completedPipelines: {
		// requestId: {
		//     pipelineId: {
		//         name: '',
		//         id: '',
		//         dataset: '',
		//         pipelineId: '',
		//         progress: '',
		//         pipeline: { // only present if progress === COMPLETE
		//             output: '',
		//             scores: [
		//                {
		//                    metric: '',
		//                    value: 0.1
		//                }
		//             ],
		//             resultUri: ''
		//         }
		//     }
		// }
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
