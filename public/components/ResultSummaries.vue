<template>
	<div class='result-summaries'>
		<div class="bg-faded rounded-top">
			<h6 class="nav-link">Summaries</h6>
		</div>
		<variable-facets
			enable-filter="true"
			enable-toggle="true"
			:variables="variables"
			:dataset="dataset"></variable-facets>
	</div>
</template>

<script>

import VariableFacets from '../components/VariableFacets';
import 'font-awesome/css/font-awesome.css';

export default {
	name: 'result-summaries',

	components: {
		VariableFacets
	},

	mounted() {
		// kick off a result fetch when the component is first displayed
		this.$store.dispatch('getResultsSummaries', {
			dataset: this.$store.getters.getRouteDataset(),
			requestId: this.$store.getters.getRouteCreateRequestId()
		});
	},

	computed: {
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		variables() {
			return this.$store.state.resultsSummaries;
		}
	},

	watch: {
		// watch the route and update the results if its modified
		'$route.query.results'() {
			this.$store.dispatch('getResultsSummaries', {
				dataset: this.$store.getters.getRouteDataset(),
				requestId: this.$store.getters.getRouteCreateRequestId()
			});
		}
	}
};
</script>

<style>
.result-summaries {
	display: flex;
	flex-direction: column;
}
</style>
