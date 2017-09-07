<template>
	<div class='variable-summaries'>
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
	name: 'variable-summaries',

	components: {
		VariableFacets
	},

	computed: {
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		variables() {
			return this.$store.getters.getVariableSummaries();
		}
	}
};
</script>

<style>
.variable-summaries {
	display: flex;
	flex-direction: column;
}
</style>
