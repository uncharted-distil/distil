<template>
	<div class='result-summaries'>
		<div class="bg-faded rounded-top">
			<h6 class="nav-link">Results</h6>
		</div>
		<facets class="result-summaries-target" :groups="targetVariable">
		</facets>
		<result-facets
			:variables="variables"
			:dataset="dataset">
		</result-facets>
	</div>
</template>

<script>

import ResultFacets from '../components/ResultFacets';
import Facets from '../components/Facets';
import { createGroups } from '../util/facets';
import _ from 'lodash';
import 'font-awesome/css/font-awesome.css';

export default {
	name: 'result-summaries',

	components: {
		ResultFacets,
		Facets,
	},

	computed: {

		dataset() {
			return this.$store.getters.getRouteDataset();
		},

		targetVariable() {
			// Get the current target variable and the summary associated with it
			const targetVariable = this.$store.getters.getRouteTargetVariable();
			const varSummaries = this.$store.getters.getVariableSummaries();
			const targetSummary = _.find(varSummaries, v => v.name === targetVariable);
			// Create a facet for it - this will act as a basis of comparison for the result sets
			if (!_.isEmpty(targetSummary)) {
				return createGroups([targetSummary]);
			}
			return [];
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
.result-summaries-target {
	margin-bottom: 6px;
}
.result-summaries-target .facets-facet-horizontal .facet-histogram-bar-highlighted {
	fill: #00C851;
}
.result-summaries-target .facets-facet-horizontal .facet-histogram-bar-highlighted:hover {
	  fill: #007E33;
}
.result-summaries-target .facets-facet-vertical .facet-bar-selected {
	box-shadow: inset 0 0 0 1000px #00C851;
}

.result-summaries-target .facets-facet-horizontal .facet-range-filter {
	box-shadow: inset 0 0 0 1000px rgba(0, 225, 11, 0.15);
}
</style>
