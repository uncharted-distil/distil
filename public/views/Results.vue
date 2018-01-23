<template>
	<div class="container-fluid d-flex flex-column h-100 results-view">
		<div class="row flex-0-nav">
		</div>
		<div class="row flex-1 align-items-center bg-white">
			<div class="col-12">
				<h5 class="header-label">Select Features That May Predict</h5>
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
import { gotoSelect } from '../util/nav';
import { getRequestIdsForDatasetAndTarget, getTrainingVariablesForPipelineId } from '../util/pipelines';
import { getters as dataGetters, actions as dataActions } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { actions as pipelineActions, getters as pipelineGetters } from '../store/pipelines/module';
import { Variable, VariableSummary } from '../store/data/index';
import { Dictionary } from '../util/dict';
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
				return dataGetters.getVariableSummaries(this.$store).filter(summary => this.training[summary.name]);
			}
			return dataGetters.getVariableSummaries(this.$store);
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
		}
	},

	mounted() {
		this.fetch();
	},

	watch: {
		// watch the route and update the results if its modified
		'$route.query.dataset'() {
			this.fetch();
		}
	},

	// need session id to pull

	methods: {
		gotoSelect() {
			gotoSelect(this.$store, this.$router);
		},
		fetch() {
			Promise.all([
					dataActions.getVariables(this.$store, {
						dataset: this.dataset
					}),
					pipelineActions.fetchPipelines(this.$store, {
						sessionId: this.sessionId
					})
				])
				.then(() => {
					dataActions.getVariableSummaries(this.$store, {
						dataset: this.dataset,
						variables: this.variables
					});
					dataActions.getResultsSummaries(this.$store, {
						dataset: this.dataset,
						requestIds: this.requestIds
					});
					dataActions.getResidualsSummaries(this.$store, {
						dataset: this.dataset,
						requestIds: this.requestIds
					});
				});
		}
	}
});
</script>

<style>
.results-view .nav-link {
	padding: 1rem 0 0.25rem 0;
	border-bottom: 1px solid #E0E0E0;
}
.header-label {
	padding: 1rem 0 0.5rem 0;
	font-weight: bold;
}
.results-data-table-container {
	background-color: white;
}
</style>