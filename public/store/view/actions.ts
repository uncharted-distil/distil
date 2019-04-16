import _ from 'lodash';
import { ViewState } from './index';
import { ActionContext } from 'vuex';
import { DistilState } from '../store';
import { mutations } from './module';

function updateViewState(context: ViewContext) {
	const routeDataset = context.getters.getRouteDataset;
	const currentViewDataset = context.getters.getViewActiveDataset;

	const routeTarget = context.getters.getRouteTargetVariable;
	const currentViewTarget = context.getters.getViewSelectedTarget;

	const viewStateChangeResult = {
		dataset: routeDataset,
		isDatasetUpdated: false,
		target: routeTarget,
		isTargetUpdated: false,
	};

	if (routeDataset && (routeDataset !== currentViewDataset)) {
		mutations.setViewActiveDataset(context, routeDataset);
		viewStateChangeResult.isDatasetUpdated = true;
		viewStateChangeResult.isTargetUpdated = true;
	}

	if (routeTarget && (routeTarget !== currentViewTarget)) {
		mutations.setViewSelectedTarget(context, routeTarget);
		viewStateChangeResult.isTargetUpdated = true;
	}
	return viewStateChangeResult;
}

const cache = {
	fetchVariables: '',
	fetchVariableSummaries: '',
	fetchVariableRankings: '',
	fetchSolutionRequests: '',
};

function updateVariables(context: ViewContext, dataset: string) {

	const result = { dataset, variables: context.getters.getVariables };

	console.log(cache);
	if (cache.fetchVariables !== dataset) {
		cache.fetchVariables = dataset;
		console.log('fetchVariables');
		return context.dispatch('fetchVariables', { dataset }).then(() => {
			result.variables = context.getters.getVariables;
			return result;
		});
	}
	return Promise.resolve(result);
}

function updateVariableSummaries(context: ViewContext, dataset: string) {
	if (cache.fetchVariableSummaries !== dataset) {
		cache.fetchVariableSummaries = dataset;
		return updateVariables(context, dataset).then(result => {
		console.log('fetchVariablesSummaries');
			context.dispatch('fetchVariableSummaries', { dataset: result.dataset, variables: result.variables });
		});
	}
	return Promise.resolve();
}

function updateVariableRankings(context: ViewContext, dataset: string, target: string) {
	const cacheParams = `${dataset}:${target}`;
	if (cache.fetchVariableRankings !== cacheParams) {
		cache.fetchVariableRankings = cacheParams;
		console.log('fetchVariableRankings');
		context.dispatch('fetchVariableRankings', { dataset: dataset, target });
	}
}

function updateSolutionRequests(context: ViewContext, dataset: string, target: string) {
	const cacheParams = `${dataset}:${target}`;
	if (cache.fetchSolutionRequests !== cacheParams) {
		cache.fetchSolutionRequests = cacheParams;
		console.log('fetchSolutionRequests');
		return context.dispatch('fetchSolutionRequests', { dataset, target, });
	}
}

export type ViewContext = ActionContext<ViewState, DistilState>;

export const actions = {

	fetchHomeData(context: ViewContext) {
		// clear any previous state
		context.commit('clearSolutionRequests');

		// fetch new state
		return context.dispatch('fetchSolutionRequests', {});
	},

	fetchSearchData(context: ViewContext) {
		const terms = context.getters.getRouteTerms;
		const datasetIDs = context.getters.getRouteJoinDatasets;

		const promises = datasetIDs.map((id: string) => {
			return context.dispatch('fetchDataset', {
				dataset: id
			});
		});
		promises.push(context.dispatch('searchDatasets', terms));

		return Promise.all(promises);
	},

	fetchJoinDatasetsData(context: ViewContext) {
		// clear previous state
		context.commit('clearHighlightSummaries');

		const datasetIDs = context.getters.getRouteJoinDatasets;
		const datasetIDA = datasetIDs[0];
		const datasetIDB = datasetIDs[1];
		Promise.all([
				context.dispatch('fetchDataset', {
					dataset: datasetIDA
				}),
				context.dispatch('fetchDataset', {
					dataset: datasetIDB
				}),
				context.dispatch('fetchJoinDatasetsVariables', {
					datasets: datasetIDs
				})
			])
			.then(() => {
				// fetch new state
				const datasets = context.getters.getDatasets;
				const datasetA = _.find(datasets, d => {
					return d.id === datasetIDA;
				});
				const datasetB = _.find(datasets, d => {
					return d.id === datasetIDB;
				});
				return Promise.all([
					context.dispatch('fetchVariableSummaries', {
						dataset: datasetA.id,
						variables: datasetA.variables
					}),
					context.dispatch('fetchVariableSummaries', {
						dataset: datasetB.id,
						variables: datasetB.variables
					})
				]).then(() => {
					return context.dispatch('updateJoinDatasetsData');
				});
			});
	},

	updateJoinDatasetsData(context: ViewContext) {
		// clear any previous state
		context.commit('clearHighlightSummaries');
		context.commit('clearJoinDatasetsTableData');

		const datasetIDs = context.getters.getRouteJoinDatasets;
		const highlightRoot = context.getters.getDecodedHighlightRoot;
		const filterParams = context.getters.getDecodedJoinDatasetsFilterParams;
		const paginatedVariables = context.getters.getJoinDatasetsPaginatedVariables;

		const datasets = context.getters.getDatasets;

		const joinDatasets = datasets.filter(d => {
			return d.id === datasetIDs[0] || d.id === datasetIDs[1];
		});

		return Promise.all([
			context.dispatch('fetchJoinDatasetsHighlightValues', {
				datasets: joinDatasets,
				variables: paginatedVariables,
				highlightRoot: highlightRoot,
				filterParams: filterParams
			}),
			context.dispatch('fetchJoinDatasetsTableData', {
				datasets: datasetIDs,
				filterParams: filterParams,
				highlightRoot: highlightRoot
			}),
		]);
	},

	fetchSelectTargetData(context: ViewContext) {
		// clear previous state
		context.commit('clearHighlightSummaries');

		const {dataset} = updateViewState(context);

		return updateVariables(context, dataset).then(result => {
			return updateVariableSummaries(context, dataset);
		});
	},

	fetchSelectTrainingData(context: ViewContext) {
		// clear any previous state
		context.commit('clearHighlightSummaries');
		context.commit('setIncludedTableData', null);
		context.commit('setExcludedTableData', null);

		const {isTargetUpdated, target, dataset} = updateViewState(context);
		// fetch new state
		return updateVariables(context, dataset).then(result => {
			updateVariableRankings(context, dataset, target);
			return updateVariableSummaries(context, dataset);
		}).then(() => {
			return context.dispatch('updateSelectTrainingData');
		});
	},

	updateSelectTrainingData(context: ViewContext) {
		// clear any previous state
		context.commit('clearHighlightSummaries');
		context.commit('setIncludedTableData', null);
		context.commit('setExcludedTableData', null);

		const dataset = context.getters.getRouteDataset;
		const highlightRoot = context.getters.getDecodedHighlightRoot;
		const filterParams = context.getters.getDecodedSolutionRequestFilterParams;
		const paginatedVariables = context.getters.getSelectTrainingPaginatedVariables;

		return Promise.all([
			context.dispatch('fetchDataHighlightValues', {
				dataset: dataset,
				variables: paginatedVariables,
				highlightRoot: highlightRoot,
				filterParams: filterParams
			}),
			context.dispatch('fetchIncludedTableData', {
				dataset: dataset,
				filterParams: filterParams,
				highlightRoot: highlightRoot
			}),
			context.dispatch('fetchExcludedTableData', {
				dataset: dataset,
				filterParams: filterParams,
				highlightRoot: highlightRoot
			})
		]);
	},

	fetchResultsData(context: ViewContext) {
		// clear previous state
		context.commit('clearTargetSummary');
		context.commit('clearTrainingSummaries');
		context.commit('clearHighlightSummaries');
		context.commit('clearResidualsExtrema');
		context.commit('setIncludedResultTableData', null);
		context.commit('setExcludedResultTableData', null);

		const {target, dataset} = updateViewState(context);
		// fetch new state
		return updateVariables(context, dataset).then(() => {
			updateVariableRankings(context, dataset, target);
			return updateSolutionRequests(context, dataset, target);
		}).then(() => {
			return context.dispatch('updateResultsSolution');
		});
	},

	updateResultsSolution(context: ViewContext) {
		// clear previous state
		context.commit('clearTargetSummary');
		context.commit('clearTrainingSummaries');
		context.commit('clearHighlightSummaries');
		context.commit('clearResidualsExtrema', null);
		context.commit('setIncludedResultTableData', null);
		context.commit('setExcludedResultTableData', null);

		// fetch new state
		const dataset = context.getters.getRouteDataset;
		const target = context.getters.getRouteTargetVariable;
		const isRegression = context.getters.isRegression;
		const isClassification = context.getters.isClassification;
		const requestIds = context.getters.getRelevantSolutionRequestIds;
		const solutionId = context.getters.getRouteSolutionId;
		const paginatedVariables = context.getters.getResultsPaginatedVariables;
		const trainingVariables = context.getters.getActiveSolutionTrainingVariables;
		const highlightRoot = context.getters.getDecodedHighlightRoot;

		context.dispatch('fetchResultTableData', {
			dataset: dataset,
			solutionId: solutionId,
			highlightRoot: highlightRoot
		});
		context.dispatch('fetchTargetSummary', {
			dataset: dataset,
			target: target,
			solutionId: solutionId
		});
		context.dispatch('fetchTrainingSummaries', {
			dataset: dataset,
			training: trainingVariables,
			solutionId: solutionId
		});
		context.dispatch('fetchPredictedSummaries', {
			dataset: dataset,
			target: target,
			requestIds: requestIds
		});
		context.dispatch('fetchResultHighlightValues', {
			dataset: dataset,
			target: target,
			training: paginatedVariables,
			highlightRoot: highlightRoot,
			solutionId: solutionId,
			requestIds: requestIds,
			includeCorrectness: isClassification,
			includeResidual: isRegression
		});

		if (isRegression) {
			context.dispatch('fetchResidualsExtrema', {
				dataset: dataset,
				target: target,
				solutionId: solutionId
			});
			context.dispatch('fetchResidualsSummaries', {
				dataset: dataset,
				target: target,
				requestIds: requestIds,
			});
		} else if (isClassification) {
			context.dispatch('fetchCorrectnessSummaries', {
				dataset: dataset,
				requestIds: requestIds
			});
		}
	},

	updateResultsHighlights(context: ViewContext) {
		// clear previous state
		context.commit('clearHighlightSummaries');
		context.commit('setIncludedResultTableData', null);
		context.commit('setExcludedResultTableData', null);

		const dataset = context.getters.getRouteDataset;
		const target = context.getters.getRouteTargetVariable;
		const requestIds = context.getters.getRelevantSolutionRequestIds;
		const solutionId = context.getters.getRouteSolutionId;
		const isClassification = context.getters.isClassification;
		const isRegression = context.getters.isRegression;
		const paginatedVariables = context.getters.getResultsPaginatedVariables;
		const highlightRoot = context.getters.getDecodedHighlightRoot;

		return Promise.all([
			context.dispatch('fetchResultHighlightValues', {
				dataset: dataset,
				target: target,
				training: paginatedVariables,
				highlightRoot: highlightRoot,
				solutionId: solutionId,
				requestIds: requestIds,
				includeCorrectness: isClassification,
				includeResidual: isRegression
			}),
			context.dispatch('fetchResultTableData', {
				dataset: dataset,
				solutionId: solutionId,
				highlightRoot: highlightRoot
			})
		]);
	}
};
