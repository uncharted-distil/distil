<template>
	<div v-bind:class="currentClass"
		@click="click()">
		{{ name }}
		<facets class="result-container"
			:groups="results">
		</facets>
		<facets class="residual-container"
			:groups="residuals"
			:highlights="highlights"
			:html="html">
		</facets>
	</div>
</template>

<script lang="ts">

// Component that contains a histogram of regression predictions, a histogram of the
// of prediction-truth residuals, and scoring information.

import { createGroups, Group } from '../util/facets';
import Facets from '../components/Facets';
import { VariableSummary, Dictionary } from '../store/data/index';
import { getters } from '../store/data/module';
import Vue from 'vue';

export default Vue.extend({
	name: 'result-group',

	props: {
		name: String,
		resultSummary: Object,
		residualsSummary: Object,
		html: String,
		selectedId: String
	},

	components: {
		Facets
	},

	computed: {
		residuals(): Group[] {
			return createGroups([<VariableSummary>this.residualsSummary], false, true);
		},

		results(): Group[] {
			return createGroups([<VariableSummary>this.resultSummary], false, true);
		},

		highlights(): Dictionary<any> {
			return getters.getHighlightedFeatureValues(this.$store);
		},

		currentClass(): string {
			return this.resultSummary.pipelineId === this.selectedId ? 'result-group-selected' : 'result-group';
		}
	},

	methods: {
		click() {
			this.$emit('selected', this.residualsSummary.pipelineId);
		},
	}
});
</script>

<style>
.result-group {
	margin-left: 5px;
	padding: 10px;
	border-bottom-style: solid;
	border-bottom-color:lightgray;
	border-bottom-width: 1px;
}

.result-group-selected {
	padding:9px;
	border-style: solid;
	border-color: #03c6e1;
	border-width: 1px;
	border-radius: 2px;
	padding-bottom: 10px;
}

.result-group:not(.result-group-selected):hover {
	padding:9px;
	border-style: solid;
	border-color: lightgray;
	border-width: 1px;
	border-radius: 2px;
	padding-bottom: 10px;
}

.result-container {
	box-shadow: none;
}

.result-container {
	box-shadow: none;
}

.result-container .facets-group {
	box-shadow: none;
}

.residual-container .facets-group {
	box-shadow: none;
}

.residual-container .facets-facet-horizontal .facet-histogram-bar-highlighted {
	fill: #e05353
}

.residual-container .facets-facet-horizontal .facet-histogram-bar-highlighted:hover {
	fill: #662424;
}

.residual-container .facets-facet-vertical .facet-bar-selected {
	box-shadow: inset 0 0 0 1000px #e0535e;
}

.residual-container .facets-facet-horizontal .facet-range-filter {
	box-shadow: inset 0 0 0 1000px rgba(225, 0, 11, 0.15);
}

</style>
