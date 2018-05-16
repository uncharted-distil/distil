import { ViewState } from './index';
import { ActionContext } from 'vuex';
import { DistilState } from '../store';

export type ViewContext = ActionContext<ViewState, DistilState>;

export const actions = {

	fetchHomeData(context: ViewContext) {
		return context.dispatch('fetchSolutions', {});
	},

	fetchSearchData(context: ViewContext) {
		const terms = context.getters.getRouteTerms;
		return context.dispatch('datasetActions', terms);
	},

	fetchSelectTargetData(context: ViewContext) {
		// clear previous state
		context.commit('clearHighlightSummaries');
		context.commit('updateHighlightSamples', null);

		// fetch new state
		const dataset = context.getters.getRouteDataset;

		return context.dispatch('fetchVariables', {
			dataset: dataset
		}).then(() => {
			const variables = context.getters.getVariables;
			return context.dispatch('fetchVariableSummaries', {
				dataset: dataset,
				variables: variables
			});
		});
	},

	fetchSelectTrainingData(context: ViewContext) {
		// clear any previous state
		context.commit('clearHighlightSummaries');
		context.commit('updateHighlightSamples', null);
		context.commit('setIncludedTableData', null);
		context.commit('setExcludedTableData', null);

		// fetch new state
		const dataset = context.getters.getRouteDataset;

		return context.dispatch('fetchVariables', {
			dataset: dataset
		}).then(() => {
			const variables = context.getters.getVariables;
			return Promise.all([
				context.dispatch('fetchVariableSummaries', {
					dataset: dataset,
					variables: variables
				}),
				context.dispatch('updateSelectTrainingData')
			]);
		});
	},

	updateSelectTrainingData(context: ViewContext) {
		// clear any previous state
		context.commit('clearHighlightSummaries');
		context.commit('updateHighlightSamples', null);
		context.commit('setIncludedTableData', null);
		context.commit('setExcludedTableData', null);

		const dataset = context.getters.getRouteDataset;
		const highlightRoot = context.getters.getDecodedHighlightRoot;
		const filterParams = context.getters.getDecodedFilterParams;
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
		context.commit('clearHighlightSummaries');
		context.commit('updateHighlightSamples', null);
		context.commit('clearResultExtrema', null);
		context.commit('clearPredictedExtremas', null);
		context.commit('clearResidualsExtrema', null);
		context.commit('setIncludedResultTableData', null);
		context.commit('setExcludedResultTableData', null);

		// fetch new state
		const dataset = context.getters.getRouteDataset;

		return context.dispatch('fetchVariables', {
			dataset: dataset
		}).then(() => {
			const target = context.getters.getRouteTargetVariable;
			context.dispatch('fetchSolutions', {
				dataset: dataset,
				target: target
			}).then(() => {
				return context.dispatch('updateResultsSolution');
			});
		});
	},

	updateResultsSolution(context: ViewContext) {
		// clear previous state
		context.commit('clearResultExtrema', null);
		context.commit('clearPredictedExtremas', null);
		context.commit('clearResidualsExtrema', null);
		context.commit('setIncludedResultTableData', null);
		context.commit('setExcludedResultTableData', null);

		// fetch new state
		const dataset = context.getters.getRouteDataset;
		const target = context.getters.getRouteTargetVariable;
		const isRegression = context.getters.isRegression;
		const variables = context.getters.getVariables;
		const requestIds = context.getters.getSolutionRequestIds;
		const solutionId = context.getters.getRouteSolutionId;
		const paginatedVariables = context.getters.getResultsPaginatedVariables;
		const highlightRoot = context.getters.getDecodedHighlightRoot;

		let extremaFetches = [];
		if (isRegression) {
			extremaFetches = [
				context.dispatch('fetchResultExtrema', {
					dataset: dataset,
					variable: target,
					solutionId: solutionId
				}),
				context.dispatch('fetchPredictedExtremas', {
					dataset: dataset,
					requestIds: requestIds
				})
			];
		}
		Promise.all(extremaFetches).then(() => {
			const predictedExtrema = context.getters.getPredictedExtrema;
			context.dispatch('fetchTrainingResultSummaries', {
				dataset: dataset,
				variables: variables,
				solutionId: solutionId,
				extrema: predictedExtrema
			});
			context.dispatch('fetchPredictedSummaries', {
				dataset: dataset,
				requestIds: requestIds,
				extrema: predictedExtrema
			});
			context.dispatch('fetchResultHighlightValues', {
				dataset: dataset,
				highlightRoot: highlightRoot,
				solutionId: solutionId,
				requestIds: requestIds,
				extrema: predictedExtrema,
				variables: paginatedVariables
			});
		});

		if (isRegression) {
			context.dispatch('fetchResidualsExtremas', {
				dataset: dataset,
				requestIds: requestIds
			}).then(() => {
				const residualExtrema = context.getters.getResidualExtrema;
				context.dispatch('fetchResidualsSummaries', {
					dataset: dataset,
					requestIds: requestIds,
					extrema: residualExtrema
				});
			});
		} else {
			context.dispatch('fetchCorrectnessSummaries', {
				dataset: dataset,
				requestIds: requestIds
			});
		}
	},

	updateResultsActiveSolution(context: ViewContext) {
		// clear previous state
		context.commit('clearResultExtrema', null);
		context.commit('clearPredictedExtremas', null);
		context.commit('clearResidualsExtrema', null);
		context.commit('setIncludedResultTableData', null);
		context.commit('setExcludedResultTableData', null);

		// fetch new state
		const dataset = context.getters.getRouteDataset;
		const solutionId = context.getters.getRouteSolutionId;
		const highlightRoot = context.getters.getDecodedHighlightRoot;

		return Promise.all([
			context.dispatch('updateResultsSolution'),
			context.dispatch('fetchResultTableData', {
				dataset: dataset,
				solutionId: solutionId,
				highlightRoot: highlightRoot
			})
		]);
	},

	updateResultsHighlights(context: ViewContext) {
		// clear previous state
		context.commit('setIncludedResultTableData', null);
		context.commit('setExcludedResultTableData', null);

		const dataset = context.getters.getRouteDataset;
		const requestIds = context.getters.getSolutionRequestIds;
		const solutionId = context.getters.getRouteSolutionId;
		const predictedExtrema = context.getters.getPredictedExtrema;
		const paginatedVariables = context.getters.getResultsPaginatedVariables;
		const highlightRoot = context.getters.getDecodedHighlightRoot;

		return Promise.all([
			context.dispatch('fetchResultHighlightValues', {
				dataset: dataset,
				highlightRoot: highlightRoot,
				solutionId: solutionId,
				requestIds: requestIds,
				extrema: predictedExtrema,
				variables: paginatedVariables
			}),
			context.dispatch('fetchResultTableData', {
				dataset: dataset,
				solutionId: solutionId,
				highlightRoot: highlightRoot
			})
		]);
	}
}
