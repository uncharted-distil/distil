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
import { getters as dataGetters } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import Vue from 'vue';

export default Vue.extend({
	name: 'variable-summaries',

	components: {
		VariableFacets
	},

	computed: {
		dataset() {
			return routeGetters.getRouteDataset(this.$store);
		},
		variables() {
			return dataGetters.getVariableSummaries(this.$store);
		}
	}
});
</script>

<style>
.variable-summaries {
	display: flex;
	flex-direction: column;
}
</style>
