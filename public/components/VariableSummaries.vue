<template>
	<div class='variable-summaries'>
		<h6 class="nav-link">Feature Summaries</h6>
		<variable-facets
			enable-search
			enable-toggle
			enable-sort
			enable-group-collapse
			enable-facet-filtering
			instance-name="varSummaries"
			:variables="variables"
			:dataset="dataset">
		</variable-facets>
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
