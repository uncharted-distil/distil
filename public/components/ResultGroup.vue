<template>
	<div v-bind:class="currentClass"
		@click="click()">
		{{name}} <sup>{{index}}</sup> {{timestamp}}
		<div v-if="solutionStatus !== 'COMPLETED' && solutionStatus !== 'ERRORED'">
			<b-badge variant="info">{{solutionStatus}}</b-badge>
			<b-progress
				:value="100"
				variant="outline-secondary"
				striped
				:animated="true"></b-progress>
		</div>
		<div v-if="solutionStatus === 'COMPLETED' || solutionStatus === 'UPDATED'">
			<b-badge variant="info" v-bind:key="`${score.metric}-${solutionId}`" v-for="score in scores">
				{{metricName(score.metric)}}: {{score.value}}
			</b-badge>
			<facets v-if="resultGroups.length" class="result-container"
				@facet-click="onResultCategoricalClick"
				@numerical-click="onResultNumericalClick"
				@range-change="onResultRangeChange"
				:groups="resultGroups"
				:highlights="highlights"
				:instanceName="predictedInstanceName"
				:html="residualHtml">
			</facets>
			<div class="residual-group-container">
				<facets v-if="residualGroups.length" class="residual-container"
					ignore-highlights
					:groups="residualGroups"
					:highlights="highlights"
					:deemphasis="residualThreshold"
					:instanceName="residualInstanceName"
					:html="resultHtml">
				</facets>
				<div class="residual-center-line"></div>
				<div class="residual-center-label">0</div>
			</div>
			<facets v-if="accuracyGroups.length" class="result-container"
				@facet-click="onResultCategoricalClick"
				:groups="accuracyGroups"
				:highlights="highlights"
				:instanceName="accuracyInstanceName"
				:html="residualHtml">
			</facets>
		</div>
		<div v-if="solutionStatus === 'ERRORED'">
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
import { Extrema } from '../store/data/index';
import { getPredictedCol, getErrorCol } from '../util/data';
import { Highlight } from '../store/data/index';
import { getters as routeGetters } from '../store/route/module';
import { getSolutionById, getMetricDisplayName } from '../util/solutions';
import { overlayRouteEntry } from '../util/routes';
import { getHighlights, updateHighlightRoot, clearHighlightRoot } from '../util/highlights';
import _ from 'lodash';

export default Vue.extend({
	name: 'result-group',

	props: {
		name: String,
		index: Number,
		timestamp: String,
		requestId: String,
		solutionId: String,
		scores: Array,
		predictedSummary: Object,
		residualsSummary: Object,
		accuracySummary: Object,
		resultHtml: String,
		residualHtml: String
	},

	data() {
		return {
			predictedInstanceName: 'predicted-result-facet',
			residualInstanceName: 'residual-result-facet',
			accuracyInstanceName: 'accuracy-result-facet'
		};
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

		solutionStatus(): String {
			const solution = getSolutionById(this.$store.state.solutionModule, this.solutionId);
			if (solution) {
				return solution.progress;
			}
			return 'unknown';
		},

		resultGroups(): Group[] {
			if (this.predictedSummary) {
				const predicted = createGroups([ this.predictedSummary ]);
				if (this.highlights.root) {
					const group = predicted[0];
					if (group.key === this.highlights.root.key) {
						group.facets.forEach(facet => {
							facet.filterable = true;
						});
					}
				}
				return predicted;
			}
			return [];
		},

		accuracyGroups(): Group[] {
			if (this.accuracySummary) {
				const accuracy = createGroups([ this.accuracySummary ]);
				if (this.highlights.root) {
					const group = accuracy[0];
					if (group.key === this.highlights.root.key) {
						group.facets.forEach(facet => {
							facet.filterable = true;
						});
					}
				}
				return accuracy;
			}
			return [];
		},

		residualGroups(): Group[] {
			if (this.residualsSummary) {
				return createGroups([this.residualsSummary]);
			}
			return [];
		},

		highlights(): Highlight {
			return getHighlights(this.$store);
		},

		currentClass(): string {
			const selectedId = routeGetters.getRouteSolutionId(this.$store);
			const predicted = this.predictedSummary;
			return (predicted && predicted.solutionId === selectedId)
				? 'result-group-selected result-group' : 'result-group';
		},

		residualThreshold(): Extrema {
			return {
				min: _.toNumber(routeGetters.getRouteResidualThresholdMin(this.$store)),
				max: _.toNumber(routeGetters.getRouteResidualThresholdMax(this.$store))
			};
		}
	},

	methods: {

		metricName(metric): string {
			return getMetricDisplayName(metric);
		},

		onResultCategoricalClick(context: string, key: string, value: string) {
			if (key && value) {
				// extract the var name from the key
				updateHighlightRoot(this, {
					context: context,
					key: this.predictedColumnName,
					value: value
				});
			} else {
				clearHighlightRoot(this);
			}
		},

		onResultNumericalClick(context: string, key: string) {
			if (!this.highlights.root || this.highlights.root.key !== key) {
				updateHighlightRoot(this, {
					context: context,
					key: this.predictedColumnName,
					value: null
				});
			}
		},

		onResultRangeChange(context: string, key: string, value: { from: { label: string[] }, to: { label: string[] } }) {
			updateHighlightRoot(this, {
				context: context,
				key: this.predictedColumnName,
				value: value
			});
			this.$emit('range-change', key, value);
		},

		click() {
			if (this.predictedSummary) {
				const routeEntry = overlayRouteEntry(this.$route, {
					solutionId: this.predictedSummary.solutionId
				});
				this.$router.push(routeEntry);
			}
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

.residual-group-container {
	position: relative;
}

.residual-center-line {
	position: absolute;
	left: 50%;
	top: 20px;
	height: 42px;
	width: 1px;
	background-color: #666;
}

.residual-center-label {
	position: absolute;
	top: 68px;
	width: 100%;
	color: #666;
	text-align: center;
	font-family: Helvetica Neue,Helvetica,Arial,sans-serif;
    font-size: 11px;
}

</style>
