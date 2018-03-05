<template>
	<div v-bind:class="currentClass"
		@click="click()">
		{{name}} <sup>{{index}}</sup> {{timestamp}}
		<div v-if="pipelineStatus !== 'COMPLETED' && pipelineStatus !== 'ERRORED'">
			<b-badge variant="info">{{pipelineStatus}}</b-badge>
			<b-progress
				:value="100"
				variant="outline-secondary"
				striped
				:animated="true"></b-progress>
		</div>
		<div v-if="pipelineStatus === 'COMPLETED' || pipelineStatus === 'UPDATED'">
			<b-badge variant="info" v-bind:key="`${score.metric}-${pipelineId}`" v-for="score in scores">
				{{metricName(score.metric)}}: {{score.value}}
			</b-badge>
			<facets v-if="resultGroups.length" class="result-container"
				@facet-click="onResultFacetClick"
				@range-change="onResultRangeChange"
				:groups="resultGroups"
				:highlights="highlights"
				:filters="filters"
				:html="residualHtml">
			</facets>
			<facets v-if="residualGroups.length" class="residual-container"
				@histogram-click="onResidualsHistogramClick"
				@range-change="onResidualRangeChange"
				:groups="residualGroups"
				:highlights="highlights"
				:filters="filters"
				:html="resultHtml">
			</facets>
		</div>
		<div v-if="pipelineStatus === 'ERRORED'">
			<b-badge variant="danger">
				ERROR
			</b-badge>
		</div>
	</div>
</template>

<script lang="ts">

// Component that contains a histogram of regression predictions, a histogram of the
// of prediction-truth residuals, and scoring information.

import Vue from 'vue';
import Facets from '../components/Facets';
import { createGroups, Group } from '../util/facets';
import { getPredictedCol, getErrorCol } from '../util/data';
import { VariableSummary, Highlight } from '../store/data/index';
import { getHighlights } from '../util/highlights';
import { overlayRouteEntry } from '../util/routes';
import { Filter } from '../util/filters';
import { getters as routeGetters } from '../store/route/module';
import { getPipelineById, getMetricDisplayName } from '../util/pipelines';
import { createNumericalFilter, createCategoricalFilter, updateFilterRoute } from '../util/filters';
import { updateHighlightRoot, clearHighlightRoot } from '../util/highlights';

export default Vue.extend({
	name: 'result-group',

	props: {
		name: String,
		index: Number,
		timestamp: String,
		requestId: String,
		pipelineId: String,
		scores: Array,
		predictedSummary: Object,
		residualsSummary: Object,
		resultHtml: String,
		residualHtml: String,
		instanceName: {
			type: String,
			default: 'result-group'
		}
	},

	components: {
		Facets
	},

	computed: {

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},

		predictedColumnName(): string {
			return getPredictedCol(this.target);
		},

		errorColumnName(): string {
			return getErrorCol(this.target);
		},

		pipelineStatus(): String {
			const pipeline = getPipelineById(this.$store.state.pipelineModule, this.pipelineId);
			if (pipeline) {
				return pipeline.progress;
			}
			return 'unknown';
		},

		resultGroups(): Group[] {
			if (this.predicted()) {
				return createGroups([this.predicted()], false, true);
			}
			return [];
		},

		residualGroups(): Group[] {
			if (this.residuals()) {
				return createGroups([this.residuals()], false, true);
			}
			return [];
		},

		filters(): Filter[] {
			return routeGetters.getDecodedFilters(this.$store);
		},

		highlights(): Highlight {
			return getHighlights(this.$store);
		},

		currentClass(): string {
			const selectedId = routeGetters.getRoutePipelineId(this.$store);
			const predicted = this.predicted();
			return (predicted && predicted.pipelineId === selectedId)
				? 'result-group-selected result-group' : 'result-group';
		}
	},

	methods: {
		metricName(metric): string {
			return getMetricDisplayName(metric);
		},

		onResultRangeChange(key: string, value: { from: { label: string[] }, to: { label: string[] } }) {
			const filter = createNumericalFilter(this.predictedColumnName, value);
			updateFilterRoute(this, filter);
		},

		onResidualRangeChange(key: string, value: { from: { label: string[] }, to: { label: string[] } }) {
			const filter = createNumericalFilter(this.errorColumnName, value);
			updateFilterRoute(this, filter);
		},

		onResultHistogramClick(context: string, key: string, value: any) {
			this.histogramHighlights(context, this.predictedColumnName, value);
		},

		onResidualsHistogramClick(context, key: string, value: any) {
			this.histogramHighlights(context, this.errorColumnName, value);
		},

		onResultFacetToggle(key: string, values: string[]) {
			const filter = createCategoricalFilter(this.predictedColumnName, values);
			updateFilterRoute(this, filter);
		},

		onResultFacetClick(context: string, key: string, value: string) {
			this.histogramHighlights(context, this.predictedColumnName, value);
		},

		histogramHighlights(context: string, key: string, value: any) {
			if (value) {
				updateHighlightRoot(this, {
					context: context,
					key: key,
					value: value
				});
			} else {
				clearHighlightRoot(this);
			}
		},

		click() {
			if (this.predicted()) {
				const routeEntry = overlayRouteEntry(this.$route, {
					pipelineId: this.predicted().pipelineId
				});
				this.$router.push(routeEntry);
			}
		},

		predicted(): VariableSummary {
			if (this.predictedSummary) {
				return this.predictedSummary as VariableSummary;
			}
			return null;
		},

		residuals(): VariableSummary {
			if (this.residualsSummary) {
				return this.residualsSummary as VariableSummary;
			}
			return null;
		}
	}
});
</script>

<style>
.result-group {
	margin: 5px;
	padding: 10px;
	border-bottom-style: solid;
	border-bottom-color:lightgray;
	border-bottom-width: 1px;
}

.result-group-selected {
	padding:9px;
	border-style: solid;
	border-color: #007bff;
	box-shadow: 0 0 10px #007bff;
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

.result-container .facets-group,
.residual-container .facets-group {
	box-shadow: none;
}

.result-group,
.result-container .facets-group,
.result-container .facets-group .group-header,
.residual-container .facets-group,
.residual-container .facets-group .group-header {
	cursor: pointer !important;
}

.residual-container .facets-facet-horizontal .facet-histogram-bar-highlighted {
	fill: #e05353
}

.residual-container .facets-facet-horizontal .facet-histogram-bar-highlighted:hover {
	fill: #662424;
}

.residual-container .facets-facet-horizontal .facet-histogram-bar-highlighted.select-highlight {
	fill: #007bff;
}

.residual-container .facets-facet-vertical .facet-bar-selected {
	box-shadow: inset 0 0 0 1000px #e0535e;
}

.residual-container .facets-facet-horizontal .facet-range-filter {
	box-shadow: inset 0 0 0 1000px rgba(225, 0, 11, 0.15);
}

</style>
