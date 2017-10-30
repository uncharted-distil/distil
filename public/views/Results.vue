<template>
	<div class="results-view">
		<flow-bar
			left-text="Return to Select"
			:on-left="gotoSelect"
			center-text="Examine Pipeline Results">
		</flow-bar>
		<div class="results-items">
			<variable-summaries class="results-variable-summaries"></variable-summaries>
			<results-comparison class="results-result-comparison"></results-comparison>
			<result-summaries class="results-result-summaries"></result-summaries>
		</div>
	</div>
</template>

<script lang="ts">
import FlowBar from '../components/FlowBar.vue';
import ResultsComparison from '../components/ResultsComparison.vue';
import VariableSummaries from '../components/VariableSummaries.vue';
import ResultSummaries from '../components/ResultSummaries.vue';
import { gotoSelect } from '../util/nav';

export default {
	name: 'results',

	components: {
		FlowBar,
		ResultsComparison,
		VariableSummaries,
		ResultSummaries
	},

	computed: {
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		variables() {
			return this.$store.getters.getVariables();
		},
		requestId() {
			return this.$store.getters.getRouteCreateRequestId();
		},
		sessionId() {
			return this.$store.getters.getPipelineSessionID();
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
					this.$store.dispatch('getVariables', {
						dataset: this.dataset
					}),
					this.$store.dispatch('getSession', {
						sessionId: this.sessionId
					})
				])
				.then(() => {
					this.$store.dispatch('getVariableSummaries', {
						dataset: this.dataset,
						variables: this.variables
					});
					this.$store.dispatch('getResultsSummaries', {
						dataset: this.dataset,
						requestId: this.requestId
					});
				});
		}
	}
};
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
