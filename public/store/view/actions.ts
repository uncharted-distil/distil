import _ from 'lodash';
import { ViewState } from './index';
import { ActionContext } from 'vuex';
import { DistilState } from '../store';
import { mutations } from './module';
import { Dictionary } from '../../util/dict';

enum ParamCacheKey {
	VARIABLES = 'VARIABLES',
	VARIABLE_SUMMARIES = 'VARIABLE_SUMMARIES',
	VARIABLE_RANKINGS = 'VARIABLE_RANKINGS',
	SOLUTION_REQUESTS = 'SOLUTION_REQUESTS',
	JOIN_SUGGESTIONS = 'JOIN_SUGGESTIONS',
}

function createCacheable(key: ParamCacheKey, func: (context: ViewContext, args: Dictionary<string>) => any) {
	return (context: ViewContext, args: Dictionary<string>) => {
		// execute provided function if params are not cached already or changed
		const params = _.values(args).join(':');
		const cachedParams = context.getters.getFetchParamsCache[key];
		if (cachedParams !== params) {
			mutations.setFetchParamsCache(context, { key, value: params});
			return Promise.resolve(func(context, args));
		}
		return Promise.resolve();
	};
}

const fetchJoinSuggestions = createCacheable(ParamCacheKey.JOIN_SUGGESTIONS, (context, args) => {
	context.dispatch('fetchJoinSuggestions', args);
});

const fetchVariables = createCacheable(ParamCacheKey.VARIABLES, (context, args) => {
	return context.dispatch('fetchVariables', args);
});

const fetchVariableSummaries = createCacheable(ParamCacheKey.VARIABLE_SUMMARIES, (context, args) => {
	return fetchVariables(context, args).then(() => {
		const dataset = args.dataset;
		const variables = context.getters.getVariables;
		context.dispatch('fetchVariableSummaries', { dataset, variables });
	});
});

const fetchVariableRankings = createCacheable(ParamCacheKey.VARIABLE_RANKINGS, (context, args) => {
	// if target or dataset has changed, clear previous rankings before re-fetch
	// this is needed because since user decides variable rankings to be updated, re-fetching doesn't always replace the previous data
	context.dispatch('updateVariableRankings', undefined);
	context.dispatch('fetchVariableRankings', args);
});

const fetchSolutionRequests = createCacheable(ParamCacheKey.SOLUTION_REQUESTS, (context, args) => {
	return context.dispatch('fetchSolutionRequests', args);
});

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

	fetchSelectTargetData(context: ViewContext, clearSummaries: boolean) {
		// clear previous state
		context.commit('clearHighlightSummaries');
		if (clearSummaries) {
			context.commit('clearVariableSummaries');
			mutations.setFetchParamsCache(context, { key: ParamCacheKey.VARIABLE_SUMMARIES, value: undefined });
		}

		// fetch new state
		const dataset = context.getters.getRouteDataset;
		const args = { dataset };

		return fetchVariables(context, args).then(() => {
			return fetchVariableSummaries(context, args);
		});
	},

	fetchSelectTrainingData(context: ViewContext, clearSummaries: boolean) {
		// clear any previous state
		context.commit('clearHighlightSummaries');
		context.commit('setIncludedTableData', null);
		context.commit('setExcludedTableData', null);
		if (clearSummaries) {
			context.commit('clearVariableSummaries');
			mutations.setFetchParamsCache(context, { key: ParamCacheKey.VARIABLE_SUMMARIES, value: undefined });
		}

		const dataset = context.getters.getRouteDataset;
		const target = context.getters.getRouteTargetVariable;

		fetchJoinSuggestions(context, { dataset });

		return fetchVariables(context, { dataset }).then(() => {
			fetchVariableRankings(context, { dataset, target });
			return fetchVariableSummaries(context, { dataset });
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

		const dataset = context.getters.getRouteDataset;
		const target = context.getters.getRouteTargetVariable;
		// fetch new state
		return fetchVariables(context, { dataset }).then(() => {
			fetchVariableRankings(context, { dataset, target });
			return fetchSolutionRequests(context, { dataset, target });
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
		const isForecasting = context.getters.isForecasting;
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
		} else if (isForecasting) {
			context.dispatch('fetchForecastingSummaries', {
				dataset: dataset,
				requestIds: requestIds
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
