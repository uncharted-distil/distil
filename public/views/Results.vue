<template>
	<div class="results-view">
		<flow-bar
			left-text="Return to Select"
			:on-left="gotoSelect"
			center-text="Examine Pipeline Results">
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
import { getters as dataGetters, actions as dataActions } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { actions as pipelineActions } from '../store/pipelines/module';
import { getters as appGetters } from '../store/app/module';
import { Variable, VariableSummary } from '../store/data/index';
import Vue from 'vue';

export default Vue.extend({
	name: 'results',

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
		summaries(): VariableSummary[] {
			if (this.excludeNonTraining) {
				return dataGetters.getTrainingVariableSummaries(this.$store);
			}
			return dataGetters.getVariableSummaries(this.$store);
		},
		variables(): Variable[] {
			return dataGetters.getVariables(this.$store);
		},
		requestId(): string {
			return routeGetters.getRouteCreateRequestId(this.$store);
		},
		sessionId(): string {
			return appGetters.getPipelineSessionID(this.$store);
		}
	},

	mounted() {
		this.fetch();
	},

	watch: {
		// watch the route and update the results if its modified
		'$route.query.dataset'() {
			this.fetch();
		},
		'$route.query.requestId'() {
			this.fetch();
		}
	},

	methods: {
		gotoSelect() {
			gotoSelect(this.$store, this.$router);
		},
		fetch() {
			Promise.all([
					dataActions.getVariables(this.$store, {
						dataset: this.dataset
					}),
					pipelineActions.getSession(this.$store, {
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
						requestId: this.requestId
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
.results-view {
	display: flex;
	justify-content: space-around;
	flex-direction: column;
	align-items: center;
}
.results-items {
	display: flex;
	justify-content: space-around;
	padding: 8px;
	width: 100%;
}
.results-variable-summaries {
	width: 25%;
}
.results-result-summaries {
	width: 25%;
}
.results-result-comparison {
	width: 50%;
}
</style>
