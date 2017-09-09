<template>
	<div class='result-summaries'>
		<div class="bg-faded rounded-top">
			<h6 class="nav-link">Results</h6>
		</div>
		<result-facets
			:variables="variables"
			:dataset="dataset"></result-facets>
	</div>
</template>

<script>

import ResultFacets from '../components/ResultFacets';
import 'font-awesome/css/font-awesome.css';

export default {
	name: 'result-summaries',

	components: {
		ResultFacets
	},

	computed: {
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		variables() {
			return this.$store.getters.getResultsSummaries();
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
