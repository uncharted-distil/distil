<template>
	<div class="results">
		<variable-summaries class="results-variable-summaries"></variable-summaries>
		<results-comparison class="results-result-comparison"></results-comparison>
		<result-summaries class="results-result-summaries"></result-summaries>
	</div>
</template>

<script>
import ResultsComparison from '../components/ResultsComparison';
import VariableSummaries from '../components/VariableSummaries';
import ResultSummaries from '../components/ResultSummaries';

export default {
	name: 'results',
	components: {
		ResultsComparison,
		VariableSummaries,
		ResultSummaries
	},

	mounted() {
		// kick off a result fetch when the component is first displayed
		this.$store.dispatch('getResultsSummaries', {
			dataset: this.$store.getters.getRouteDataset(),
			requestId: this.$store.getters.getRouteCreateRequestId()
		});
		this.$store.dispatch('getVariableSummaries', this.$store.getters.getRouteDataset());
	},

	watch: {
		// watch the route and update the results if its modified
		'$route.query.dataset'() {
			this.$store.dispatch('getResultsSummaries', {
				dataset: this.$store.getters.getRouteDataset(),
				requestId: this.$store.getters.getRouteCreateRequestId()
			});
			this.$store.dispatch('getVariableSummaries', this.$store.getters.getRouteDataset());
		},
		'$route.query.requestId'() {
			this.$store.dispatch('getResultsSummaries', {
				dataset: this.$store.getters.getRouteDataset(),
				requestId: this.$store.getters.getRouteCreateRequestId()
			});
		}
	}
};
</script>

<style>
.results {
	display: flex;
	justify-content: space-around;
	padding: 8px;
}
.results-variable-summaries {
	width: 20%;
}
.results-result-summaries {
	width: 20%;
}
.results-result-comparison {
	width: 60%;
}
</style>
