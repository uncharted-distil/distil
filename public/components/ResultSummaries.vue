<template>
	<div class='result-summaries'>
		<div class="bg-faded rounded-top">
			<h6 class="nav-link">Results</h6>
		</div>
		<div v-if="regressionEnabled" class="result-summaries-error">
			<div class="result-summaries-label">
				Error:
			</div>
			<div  class="result-summaries-slider">
				<vue-slider
					ref="slider"
					:v-model="value"
					:min="0"
					:max="maxVal"
					:lazy="true"
					width=100%
					tooltip-dir="bottom"
					@callback="onSlide"/>
			</div>
  		</div>
		<facets class="result-summaries-target" :groups="targetVariable">
		</facets>
		<result-facets
			enable-group-collapse
			enable-facet-filtering
			:variables="variables"
			:dataset="dataset">
		</result-facets>
	</div>
</template>

<script>

import ResultFacets from '../components/ResultFacets';
import Facets from '../components/Facets';
import { createGroups } from '../util/facets';
import { createRouteEntryFromRoute } from '../util/routes';
import { getTask } from '../util/pipelines';
import vueSlider from 'vue-slider-component';
import _ from 'lodash';
import 'font-awesome/css/font-awesome.css';

export default {
	name: 'result-summaries',

	components: {
		ResultFacets,
		Facets,
		vueSlider,
	},

	data() {
		return {
			value: this.minVal
		};
	},

	computed: {

		dataset() {
			return this.$store.getters.getRouteDataset();
		},

		minVal() {
			const resultItems = this.$store.getters.getResultDataItems(this.regressionEnabled);
			if (!_.isEmpty(resultItems) && _.has(resultItems[0], 'Error')) {
				return Math.abs(_.minBy(resultItems, r => r.Error).Error);
			}
			return 0.0;
		},

		// computes the absolute maximum residual
		maxVal() {
			const resultItems = this.$store.getters.getResultDataItems(this.regressionEnabled);
			if (!_.isEmpty(resultItems) && _.has(resultItems[0], 'Error')) {
				return Math.abs(_.maxBy(resultItems, r => r.Error).Error);
			}
			return 100.0;
		},

		targetVariable() {
			// Get the current target variable and the summary associated with it
			const targetVariable = this.$store.getters.getRouteTargetVariable();
			const varSummaries = this.$store.getters.getVariableSummaries();
			const targetSummary = _.find(varSummaries, v => _.toLower(v.name) === targetVariable);
			// Create a facet for it - this will act as a basis of comparison for the result sets
			if (!_.isEmpty(targetSummary)) {
				return createGroups([targetSummary]);
			}
			return [];
		},

		variables() {
			return this.$store.getters.getResultsSummaries();
		},

		regressionEnabled() {
			const targetVarName = this.$store.getters.getRouteTargetVariable();
			const targetVar = _.find(this.$store.getters.getVariables(), v => _.toLower(v.name) === targetVarName);
			if (_.isEmpty(targetVar)) {
				return false;
			}
			const task = getTask(targetVar.type);
			return task.schemaName === 'regression';
		}
	},

	methods: {
		onSlide(value) {
			const entry = createRouteEntryFromRoute(this.$route, { residualThreshold: value });
			this.$router.push(entry);
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
	margin-bottom: 12px;
}
.result-summaries-target .facets-group {
	box-shadow: none;
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
.result-summaries-error {
	display: flex;
	flex-direction: row;
	justify-content: flex-start;
	margin-bottom: 30px;
}
.result-summaries-label {
	display: flex;
	flex-basis:auto;
	margin-left:10px;
	margin-right:15px;
}
.result-summaries-slider {
	display: flex;
	flex-grow: 1;
}
.result-summaries-slider .vue-slider-component .vue-slider-process {
	background-color:#00C851;
}
.result-summaries-slider .vue-slider-component .vue-slider-tooltip {
	border: 1px solid #00C851;
    background-color: #00C851;
}
</style>
