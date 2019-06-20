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
		const params = JSON.stringify(args);
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
		const filterParams = context.getters.getDecodedSolutionRequestFilterParams;
		const highlight = context.getters.getDecodedHighlight;
		context.dispatch('fetchVariableSummaries', {
			dataset: dataset,
			variables: variables,
			filterParams: filterParams,
			highlight: highlight
		});
		context.dispatch('fetchTimeVariableSummaries', {
			dataset: dataset,
			filterParams: filterParams,
			highlight: highlight
		});
	});
});

const fetchVariableRankings = createCacheable(ParamCacheKey.VARIABLE_RANKINGS, (context, args) => {
	// if target or dataset has changed, clear previous rankings before re-fetch
	// this is needed because since user decides variable rankings to be updated, re-fetching doesn't always replace the previous data
	context.dispatch('updateVariableRankings', {
		dataset: args.dataset,
		rankings: {},
	});
	context.dispatch('fetchVariableRankings', {
		dataset: args.dataset,
		target: args.target
	});
});

const fetchSolutionRequests = createCacheable(ParamCacheKey.SOLUTION_REQUESTS, (context, args) => {
	return context.dispatch('fetchSolutionRequests', {
		dataset: args.dataset,
		target: args.target
	});
});

function clearVariablesParamCache(context: ViewContext) {
		// clear variable param cache to allow re-fetching variables
		mutations.setFetchParamsCache(context, { key: ParamCacheKey.VARIABLES, value: undefined });
}

function clearVariableSummaries(context: ViewContext) {
		context.commit('clearVariableSummaries');
		mutations.setFetchParamsCache(context, { key: ParamCacheKey.VARIABLE_SUMMARIES, value: undefined });
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
				return context.dispatch('updateJoinDatasetsData');
			});
	},

	updateJoinDatasetsData(context: ViewContext) {
		// clear any previous state
		context.commit('clearJoinDatasetsTableData');

		const datasetIDs = context.getters.getRouteJoinDatasets;
		const highlight = context.getters.getDecodedHighlight;
		const filterParams = context.getters.getDecodedJoinDatasetsFilterParams;
		const datasets = context.getters.getDatasets;
		const datasetIDA = datasetIDs[0];
		const datasetIDB = datasetIDs[1];

		// fetch new state
		const datasetA = _.find(datasets, d => {
			return d.id === datasetIDA;
		});
		const datasetB = _.find(datasets, d => {
			return d.id === datasetIDB;
		});

		return Promise.all([
			context.dispatch('fetchVariableSummaries', {
				dataset: datasetA.id,
				variables: datasetA.variables,
				filterParams:  filterParams,
				highlight: highlight
			}),
			context.dispatch('fetchVariableSummaries', {
				dataset: datasetB.id,
				variables: datasetB.variables,
				filterParams:  filterParams,
				highlight: highlight
			}),
			context.dispatch('fetchJoinDatasetsTableData', {
				datasets: datasetIDs,
				filterParams: filterParams,
				highlight: highlight
			}),
		]);
	},

	fetchSelectTargetData(context: ViewContext, clearSummaries: boolean) {
		// clear previous state
		if (clearSummaries) {
			clearVariableSummaries(context);
		}

		// fetch new state
		const dataset = context.getters.getRouteDataset;
		const args = {
			dataset: dataset
		};
		return fetchVariables(context, args).then(() => {
			return fetchVariableSummaries(context, args);
		});
	},

	clearJoinDatasetsData(context) {
		clearVariablesParamCache(context);
		clearVariableSummaries(context);
	},

	fetchSelectTrainingData(context: ViewContext, clearSummaries: boolean) {
		// clear any previous state
		context.commit('setIncludedTableData', null);
		context.commit('setExcludedTableData', null);

		if (clearSummaries) {
			clearVariableSummaries(context);
		}

		const dataset = context.getters.getRouteDataset;
		const target = context.getters.getRouteTargetVariable;

		fetchJoinSuggestions(context, {
			dataset: dataset
		});

		return fetchVariables(context, {
			dataset: dataset
		}).then(() => {
			fetchVariableRankings(context, { dataset, target });
			return context.dispatch('updateSelectTrainingData');
		});
	},

	updateSelectTrainingData(context: ViewContext) {
		// clear any previous state

		const dataset = context.getters.getRouteDataset;
		const highlight = context.getters.getDecodedHighlight;
		const filterParams = context.getters.getDecodedSolutionRequestFilterParams;

		return Promise.all([
			fetchVariableSummaries(context, {
				dataset: dataset,
				filterParams: filterParams,
				highlight: highlight
			}),
			context.dispatch('fetchIncludedTableData', {
				dataset: dataset,
				filterParams: filterParams,
				highlight: highlight
			}),
			context.dispatch('fetchExcludedTableData', {
				dataset: dataset,
				filterParams: filterParams,
				highlight: highlight
			})
		]);
	},

	fetchResultsData(context: ViewContext) {
		// clear previous state
		context.commit('clearTargetSummary');
		context.commit('clearTrainingSummaries');
		context.commit('clearResidualsExtrema');
		context.commit('setIncludedResultTableData', null);
		context.commit('setExcludedResultTableData', null);

		const dataset = context.getters.getRouteDataset;
		const target = context.getters.getRouteTargetVariable;
		// fetch new state
		return fetchVariables(context, {
			dataset: dataset
		}).then(() => {
			fetchVariableRankings(context, {
				dataset: dataset,
				target: target
			});
			return fetchSolutionRequests(context, {
				dataset: dataset,
				target: target
			});
		}).then(() => {
			return context.dispatch('updateResultsSolution');
		});
	},

	updateResultsSolution(context: ViewContext) {
		// clear previous state
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
		const trainingVariables = context.getters.getActiveSolutionTrainingVariables;
		const highlight = context.getters.getDecodedHighlight;

		context.dispatch('fetchResultTableData', {
			dataset: dataset,
			solutionId: solutionId,
			highlight: highlight
		});
		context.dispatch('fetchTargetSummary', {
			dataset: dataset,
			target: target,
			solutionId: solutionId,
			highlight: highlight
		});
		context.dispatch('fetchTrainingSummaries', {
			dataset: dataset,
			training: trainingVariables,
			solutionId: solutionId,
			highlight: highlight
		});
		context.dispatch('fetchPredictedSummaries', {
			dataset: dataset,
			target: target,
			requestIds: requestIds,
			highlight: highlight
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
				highlight: highlight
			});
		} else if (isClassification) {
			context.dispatch('fetchCorrectnessSummaries', {
				dataset: dataset,
				requestIds: requestIds,
				highlight: highlight
			});
		}
	}
};
