<template>
	<div class="container-fluid d-flex flex-column h-100 results-view">
		<div class="row flex-0-nav">
		</div>

		<div class="row align-items-center justify-content-center bg-white">

			<div class="col-12 col-md-6 d-flex flex-column">
				<h5 class="header-label">Select Model That Best Predicts {{target.toUpperCase()}}</h5>

				<div class="row col-12 pl-4">
					<div>
						{{target.toUpperCase()}} is being modeled as a {{targetType}}
					</div>
				</div>
				<div class="row col-12 pl-4">
					<p>
						Use interactive feature highlighting to analyze models. Go back to revise features, if needed.
					</p>
				</div>
			</div>

			<result-target-variable
				class="col-12 col-md-6 d-flex flex-column"></result-target-variable>
		</div>


		<div class="row flex-12 pb-3">
			<variable-summaries
				class="col-12 col-md-3 border-gray-right results-variable-summaries"
				enable-search
				enable-highlighting
				instance-name="result-summary-facets"
				:groups="groups"
				:dataset="dataset"></variable-summaries>
			<results-comparison
				class="col-12 col-md-6 results-result-comparison"></results-comparison>
			<result-summaries
				class="col-12 col-md-3 border-gray-left results-result-summaries"></result-summaries>
		</div>
	</div>
</template>

<script lang="ts">
import VariableSummaries from '../components/VariableSummaries.vue';
import ResultsComparison from '../components/ResultsComparison.vue';
import ResultSummaries from '../components/ResultSummaries.vue';
import ResultTargetVariable from '../components/ResultTargetVariable.vue';
import { regression, getTask } from '../util/pipelines';
import { getters as dataGetters, actions as dataActions } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { actions as pipelineActions, getters as pipelineGetters } from '../store/pipelines/module';
import { Variable, Extrema } from '../store/data/index';
import { Dictionary } from '../util/dict';
import { HighlightRoot } from '../store/data/index';
import { Group, createGroups } from '../util/facets';
import Vue from 'vue';

export default Vue.extend({
	name: 'results-view',

	components: {
		VariableSummaries,
		ResultTargetVariable,
		ResultsComparison,
		ResultSummaries
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},
		targetType(): string {
			const variables = dataGetters.getVariablesMap(this.$store);
			if (variables && variables[this.target]) {
				return variables[this.target].type;
			}
			return '';
		},
		groups(): Group[] {
			const summaries = dataGetters.getResultSummaries(this.$store).filter(summary => this.training[summary.name]);
			return createGroups(summaries);
		},
		variables(): Variable[] {
			return pipelineGetters.getActivePipelineVariables(this.$store);
		},
		requestIds(): string[] {
			return pipelineGetters.getPipelineRequestIds(this.$store);
		},
		training(): Dictionary<boolean> {
			return pipelineGetters.getActivePipelineTrainingMap(this.$store);
		},
		pipelineId(): string {
			return routeGetters.getRoutePipelineId(this.$store);
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
				highlightRoot: this.highlightRoot,
				pipelineId: this.pipelineId,
				requestIds: this.requestIds,
				extrema: this.predictedExtrema,
				variables: this.variables
			});
			dataActions.fetchResultTableData(this.$store, {
				dataset: this.dataset,
				pipelineId: this.pipelineId,
				highlightRoot: this.highlightRoot
			});
		},
		pipelineId() {
			// if this is a regression task, pull extrema as a first step
			const isRegression = this.testRegression();
			let extremaFetches = [];
			if (isRegression) {
				extremaFetches = [
					dataActions.fetchTargetResultExtrema(this.$store, {
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
				dataActions.fetchTrainingResultSummaries(this.$store, {
					dataset: this.dataset,
					variables: this.variables,
					pipelineId: this.pipelineId,
					extrema: this.predictedExtrema
				});
				dataActions.fetchResultHighlightValues(this.$store, {
					dataset: this.dataset,
					highlightRoot: this.highlightRoot,
					pipelineId: this.pipelineId,
					requestIds: this.requestIds,
					extrema: this.predictedExtrema,
					variables: this.variables
				});
			});
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
			dataActions.fetchResultTableData(this.$store, {
				dataset: this.dataset,
				pipelineId: this.pipelineId,
				highlightRoot: this.highlightRoot
			});
		}
	},

	methods: {
		fetch() {
			Promise.all([
					dataActions.fetchVariables(this.$store, {
						dataset: this.dataset
					}),
				])
				.then(() => {
					pipelineActions.fetchPipelines(this.$store, {
						dataset: this.dataset,
						target: this.target
					}).then(() => {
						const isRegression = this.testRegression();
						let extremaFetches = [];
						if (isRegression) {
							extremaFetches = [
								dataActions.fetchTargetResultExtrema(this.$store, {
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
							dataActions.fetchTrainingResultSummaries(this.$store, {
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
							dataActions.fetchResultHighlightValues(this.$store, {
								dataset: this.dataset,
								highlightRoot: this.highlightRoot,
								pipelineId: this.pipelineId,
								requestIds: this.requestIds,
								extrema: this.predictedExtrema,
								variables: this.variables
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
						};
						dataActions.fetchResultTableData(this.$store, {
							dataset: this.dataset,
							pipelineId: this.pipelineId,
							highlightRoot: this.highlightRoot
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
