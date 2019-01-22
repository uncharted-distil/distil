import _ from 'lodash';
import { ViewState } from './index';
import { ActionContext } from 'vuex';
import { DistilState } from '../store';

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

		return context.dispatch('searchDatasets', terms);
	},

	fetchJoinDatasetsData(context: ViewContext) {
		// clear previous state
		context.commit('clearHighlightSummaries');

		const datasetIDs = context.getters.getRouteJoinDatasets;
		Promise.all([
				context.dispatch('fetchDataset', datasetIDs[0]),
				context.dispatch('fetchDataset', datasetIDs[1]),
				context.dispatch('fetchJoinDatasetsVariables', {
					datasets: datasetIDs
				})
			])
			.then(() => {
				// fetch new state
				const datasets = context.getters.getDatasets;
				const datasetA = _.find(datasets, d => {
					return d.id === datasetIDs[0];
				});
				const datasetB = _.find(datasets, d => {
					return d.id === datasetIDs[1];
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
		const dataset = context.getters.getRouteDataset;
		const highlightRoot = context.getters.getDecodedHighlightRoot;
		const filterParams = context.getters.getDecodedJoinDatasetsFilterParams;
		// const paginatedVariables = context.getters.getSelectTrainingPaginatedVariables;

		return Promise.all([
			// context.dispatch('fetchDataHighlightValues', {
			// 	dataset: dataset,
			// 	variables: paginatedVariables,
			// 	highlightRoot: highlightRoot,
			// 	filterParams: filterParams
			// }),

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
		context.commit('setIncludedTableData', null);
		context.commit('setExcludedTableData', null);

		// fetch new state
		const dataset = context.getters.getRouteDataset;

		return context.dispatch('fetchVariables', {
			dataset: dataset
		}).then(() => {
			const variables = context.getters.getVariables;
			const target = context.getters.getRouteTargetVariable;
			return Promise.all([
				context.dispatch('fetchVariableSummaries', {
					dataset: dataset,
					variables: variables
				}),
				context.dispatch('fetchVariableRankings', {
					dataset: dataset,
					target: target
				})
			]).then(() => {
				return context.dispatch('updateSelectTrainingData');
			});
		});
	},

	updateSelectTrainingData(context: ViewContext) {
		// clear any previous state
		context.commit('clearHighlightSummaries');
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
		context.commit('clearTargetSummary');
		context.commit('clearTrainingSummaries');
		context.commit('clearHighlightSummaries');
		context.commit('clearResidualsExtrema');
		context.commit('setIncludedResultTableData', null);
		context.commit('setExcludedResultTableData', null);

		// fetch new state
		const dataset = context.getters.getRouteDataset;

		return context.dispatch('fetchVariables', {
			dataset: dataset
		}).then(() => {

			const target = context.getters.getRouteTargetVariable;

			context.dispatch('fetchVariableRankings', {
				dataset: dataset,
				target: target
			});

			context.dispatch('fetchSolutionRequests', {
				dataset: dataset,
				target: target
			}).then(() => {
				return context.dispatch('updateResultsSolution');
			});
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
