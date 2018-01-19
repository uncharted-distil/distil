<template>
	<div class="results-view">
		<flow-bar
			left-text="Return to Select"
			:on-left="gotoSelect"
			center-text="Examine Model Results">
		</flow-bar>
		<div class="results-items">
			<variable-summaries
				class="results-variable-summaries"
				:variables="summaries"
				:dataset="dataset"></variable-summaries>
			<results-comparison
				class="results-result-comparison"
				:exclude-non-training="excludeNonTraining"></results-comparison>
			<result-summaries
				class="results-result-summaries"></result-summaries>
		</div>
	</div>
</template>

<script lang="ts">
import FlowBar from '../components/FlowBar.vue';
import ResultsComparison from '../components/ResultsComparison.vue';
import VariableSummaries from '../components/VariableSummaries.vue';
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
		FlowBar,
		ResultsComparison,
		VariableSummaries,
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
.header-label {
	color: #333;
	margin: 0.75rem 0;
}
.results-view .nav-link {
    padding: 1rem 0 0.5rem 0;
} {
    padding: 1rem 0 0.5rem 0;
}
.results-view {
	display: flex;
	justify-content: space-around;
	flex-direction: column;
	align-items: center;
}
.results-items {
	display: flex;
	justify-content: space-around;
	width: 100%;
}
.results-variable-summaries {
	width: 25%;
	padding: 1rem;
}
.results-result-summaries {
	width: 25%;
	padding: 1rem;
}
.results-result-comparison {
	width: 50%;
	padding: 1rem;
}
.results-data-table-container {
	background-color: white;
}
</style>