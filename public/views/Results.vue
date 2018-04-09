<template>
	<div class="container-fluid d-flex flex-column h-100 results-view">
		<div class="row flex-0-nav">
		</div>
		<div class="row flex-1 align-items-center bg-white">
			<div class="col-12">
				<h5 class="header-label">Selected Features</h5>
			</div>
		</div>
		<div class="row flex-12 pb-3">
				<results-variable-summaries
					class="col-12 col-md-3 border-gray-right results-variable-summaries"
					:variables="summaries"
					:dataset="dataset"></results-variable-summaries>
				<results-comparison
					class="col-12 col-md-6 results-result-comparison"
					:exclude-non-training="excludeNonTraining"></results-comparison>
				<result-summaries
					class="col-12 col-md-3 border-gray-left results-result-summaries"></result-summaries>
		</div>
	</div>
</template>

<script lang="ts">
import ResultsComparison from '../components/ResultsComparison.vue';
import ResultsVariableSummaries from '../components/ResultsVariableSummaries.vue';
import ResultSummaries from '../components/ResultSummaries.vue';
import { getRequestIdsForDatasetAndTarget, getTrainingVariablesForPipelineId, regression, getTask } from '../util/pipelines';
import { getters as dataGetters, actions as dataActions } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { actions as pipelineActions, getters as pipelineGetters } from '../store/pipelines/module';
import { Variable, VariableSummary, Extrema } from '../store/data/index';
import { Dictionary } from '../util/dict';
import { HighlightRoot } from '../util/highlights';
import { Filter } from '../util/filters';
import Vue from 'vue';

export default Vue.extend({
	name: 'results-view',

	components: {
		ResultsComparison,
		ResultsVariableSummaries,
		ResultSummaries
	},

	data() {
		return {
			excludeNonTraining: true
		};
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},
		summaries(): VariableSummary[] {
			if (this.excludeNonTraining) {
				return dataGetters.getResultSummaries(this.$store).filter(summary => this.training[summary.name]);
			}
			return dataGetters.getResultSummaries(this.$store);
		},
		variables(): Variable[] {
			return dataGetters.getVariables(this.$store);
		},
		requestIds(): string[] {
			return getRequestIdsForDatasetAndTarget(this.$store.state.pipelineModule, this.dataset, this.target);
		},
		training(): Dictionary<boolean> {
			const training = getTrainingVariablesForPipelineId(this.$store.state.pipelineModule, this.pipelineId);
			const trainingMap = {};
			training.forEach(t => {
				trainingMap[t] = true;
			});
			return trainingMap;
		},
		pipelineId(): string {
			return routeGetters.getRoutePipelineId(this.$store);
		},
		sessionId(): string {
			return pipelineGetters.getPipelineSessionID(this.$store);
		},
		filters(): Filter[] {
			return routeGetters.getDecodedFilters(this.$store);
		},
		filterStr(): string {
			return routeGetters.getRouteFilters(this.$store);
		},
		highlightRoot(): HighlightRoot {
			return routeGetters.getDecodedHighlightRoot(this.$store);
		},
		highlightRootStr(): string {
			return routeGetters.getRouteHighlightRoot(this.$store);
		},
		predictedExtrema(): Extrema {
			return dataGetters.getPredictedExtrema(this.$store);
		},
		residualExtrema(): Extrema {
			return dataGetters.getResidualExtrema(this.$store);
		}
	},

	beforeMount() {
		this.fetch();
	},

	watch: {
		highlightRootStr() {
			dataActions.fetchResultHighlightValues(this.$store, {
				dataset: this.dataset,
				filters: this.filters,
				highlightRoot: this.highlightRoot,
				pipelineId: this.pipelineId
			});
		},
		pipelineId() {
			// if this is a regression task, pull extrema as a first step
			const isRegression = this.testRegression();
			let extremaFetches = [];
			if (isRegression) {
				extremaFetches = [
					dataActions.fetchResultExtrema(this.$store, {
						dataset: this.dataset,
						variable: this.target,
						pipelineId: this.pipelineId
					}),
					dataActions.fetchPredictedExtremas(this.$store, {
						dataset: this.dataset,
						requestIds: this.requestIds
					})
				];
			}
			Promise.all(extremaFetches).then(() => {
				dataActions.fetchResultSummaries(this.$store, {
					dataset: this.dataset,
					variables: this.variables,
					pipelineId: this.pipelineId,
					extrema: this.predictedExtrema
				});
			});
			dataActions.fetchResultHighlightValues(this.$store, {
				dataset: this.dataset,
				filters: this.filters,
				highlightRoot: this.highlightRoot,
				pipelineId: this.pipelineId
			});
			dataActions.fetchResultTableData(this.$store, {
				dataset: this.dataset,
				pipelineId: this.pipelineId,
				filters: this.filters,
			});
		},
		filterStr() {
			dataActions.fetchResultHighlightValues(this.$store, {
				dataset: this.dataset,
				filters: this.filters,
				highlightRoot: this.highlightRoot,
				pipelineId: this.pipelineId
			});
			dataActions.fetchResultTableData(this.$store, {
				dataset: this.dataset,
				pipelineId: this.pipelineId,
				filters: this.filters,
			});
		}
	},

	methods: {
		fetch() {
			Promise.all([
					dataActions.fetchVariables(this.$store, {
						dataset: this.dataset
					}),
					pipelineActions.startPipelineSession(this.$store, {
						sessionId: this.sessionId
					})
				])
				.then(() => {
					pipelineActions.fetchPipelines(this.$store, {
						sessionId: this.sessionId,
						dataset: this.dataset,
						target: this.target
					}).then(() => {
						const isRegression = this.testRegression();
						let extremaFetches = [];
						if (isRegression) {
							extremaFetches = [
								dataActions.fetchResultExtrema(this.$store, {
									dataset: this.dataset,
									variable: this.target,
									pipelineId: this.pipelineId
								}),
								dataActions.fetchPredictedExtremas(this.$store, {
									dataset: this.dataset,
									requestIds: this.requestIds
								})
							];
						}
						Promise.all(extremaFetches).then(() => {
							dataActions.fetchResultSummaries(this.$store, {
								dataset: this.dataset,
								variables: this.variables,
								pipelineId: this.pipelineId,
								extrema: this.predictedExtrema
							});
							dataActions.fetchPredictedSummaries(this.$store, {
								dataset: this.dataset,
								requestIds: this.requestIds,
								extrema: this.predictedExtrema
							});
						});

						if (isRegression) {
							dataActions.fetchResidualsExtremas(this.$store, {
								dataset: this.dataset,
								requestIds: this.requestIds
							}).then(() => {
								dataActions.fetchResidualsSummaries(this.$store, {
									dataset: this.dataset,
									requestIds: this.requestIds,
									extrema: this.residualExtrema
								});
							});
						}
						dataActions.fetchResultHighlightValues(this.$store, {
							dataset: this.dataset,
							filters: this.filters,
							highlightRoot: this.highlightRoot,
							pipelineId: this.pipelineId
						});
						dataActions.fetchResultTableData(this.$store, {
							dataset: this.dataset,
							pipelineId: this.pipelineId,
							filters: this.filters,
						});
					});
				});
		},
		// tests whether or not the results are for a regression or a classificiation
		testRegression(): boolean {
			const targetVariable = this.variables.find(s => s.name === this.target);
			const task = getTask(targetVariable.type);
			return task.schemaName === regression.schemaName;
		}
	}
});
</script>

<style>
.results-view .nav-link {
	padding: 1rem 0 0.25rem 0;
	border-bottom: 1px solid #E0E0E0;
	color: rgba(0,0,0,.87);
}
.header-label {
	padding: 1rem 0 0.5rem 0;
	font-weight: bold;
}
.results-data-table-container {
	background-color: white;
}
.results-view .table td {
	text-align: right;
}
.result-facets {
	margin-bottom: 12px;
}
</style>
