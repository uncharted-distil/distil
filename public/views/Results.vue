<template>
	<div class="results-view">
		<h4 class="header-label">Examine Pipeline Results</h4>
		<div class="results-items">
			<variable-summaries class="results-variable-summaries"></variable-summaries>
			<results-comparison class="results-result-comparison"></results-comparison>
			<result-summaries class="results-result-summaries"></result-summaries>
		</div>
	</div>
</template>

<script>
import FlowBar from '../components/FlowBar';
import ResultsComparison from '../components/ResultsComparison';
import VariableSummaries from '../components/VariableSummaries';
import ResultSummaries from '../components/ResultSummaries';

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
		fetch() {
			this.$store.dispatch('getVariables', this.dataset)
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
