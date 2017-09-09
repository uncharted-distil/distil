<template>
	<div class="results">
		<variable-summaries class="results-variable-summaries"></variable-summaries>
		<data-table class="results-table"></data-table>
		<result-summaries class="results-result-summaries"></result-summaries>
	</div>
</template>

<script>
import DataTable from '../components/DataTable';
import VariableSummaries from '../components/VariableSummaries';
import ResultSummaries from '../components/ResultSummaries';

export default {
	name: 'results',
	components: {
		DataTable,
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
.results-table {
	display: flex;
	flex-direction: column;
	width: 60%;
}
</style>
