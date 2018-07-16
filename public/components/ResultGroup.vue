<template>
	<div v-bind:class="currentClass"
		@click="click()">
		{{name}} <sup>{{solutionIndex}}</sup> {{timestamp}}
		<div v-if="isPending">
			<b-badge variant="info">{{solutionStatus}}</b-badge>
			<b-progress
				:value="100"
				variant="outline-secondary"
				striped
				:animated="true"></b-progress>
		</div>
		<div v-if="isCompleted">
			<b-badge variant="info" v-bind:key="`${score.metric}-${solutionId}`" v-for="score in scores">
				{{metricName(score.metric)}}: {{score.value.toFixed(2)}}
			</b-badge>
			<facets v-if="predictedGroups.length" class="result-container"
				@facet-click="onResultCategoricalClick"
				@numerical-click="onResultNumericalClick"
				@range-change="onResultRangeChange"
				:solution-id="solutionId"
				:groups="predictedGroups"
				:highlights="highlights"
				:instanceName="predictedInstanceName"
				:row-selection="rowSelection"
				:html="residualHtml">
			</facets>
			<div class="residual-group-container">
				<facets v-if="residualGroups.length" class="residual-container"
					@numerical-click="onResidualNumericalClick"
					@range-change="onResidualRangeChange"
					:solution-id="solutionId"
					:groups="residualGroups"
					:highlights="highlights"
					:deemphasis="residualThreshold"
					:instanceName="residualInstanceName"
					:row-selection="rowSelection"
					:html="resultHtml">
				</facets>
			</div>
			<facets v-if="correctnessGroups.length" class="result-container"
				@facet-click="onCorrectnessCategoricalClick"
				:solution-id="solutionId"
				:groups="correctnessGroups"
				:highlights="highlights"
				:instanceName="correctnessInstanceName"
				:row-selection="rowSelection"
				:html="residualHtml">
			</facets>
		</div>
		<div v-if="isErrored">
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
import { Extrema, VariableSummary } from '../store/dataset/index';
import { Highlight, RowSelection } from '../store/highlights/index';
import { SOLUTION_COMPLETED, SOLUTION_ERRORED } from '../store/solutions/index';
import { getters as routeGetters } from '../store/route/module';
import { getters as solutionGetters } from '../store/solutions/module';
import { getSolutionById, getMetricDisplayName } from '../util/solutions';
import { overlayRouteEntry } from '../util/routes';
import { getHighlights, updateHighlightRoot, clearHighlightRoot } from '../util/highlights';
import _ from 'lodash';

export default Vue.extend({
	name: 'result-group',

	props: {
		name: String,
		timestamp: String,
		requestId: String,
		solutionId: String,
		scores: Array,
		predictedSummary: Object,
		residualsSummary: Object,
		correctnessSummary: Object,
		resultHtml: String,
		residualHtml: String
	},

	components: {
		Facets
	},

	computed: {

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},

		predictedInstanceName(): string {
			return `predicted-result-facet-${this.solutionId}`;
		},

		residualInstanceName(): string {
			return `residual-result-facet-${this.solutionId}`;
		},

		correctnessInstanceName(): string {
			return `correctness-result-facet-${this.solutionId}`;
		},

		solutionStatus(): String {
			const solution = getSolutionById(this.$store.state.solutionModule, this.solutionId);
			if (solution) {
				return solution.progress;
			}
			return 'unknown';
		},

		rowSelection(): RowSelection {
			return routeGetters.getDecodedRowSelection(this.$store);
		},

		solutionIndex(): number {
			const solutions = solutionGetters.getRelevantSolutions(this.$store);
			return _.findIndex(solutions, solution => {
				return solution.solutionId === this.solutionId;
			});
		},

		predictedGroups(): Group[] {
			return this.getAndActivateGroups(this.predictedSummary, this.predictedInstanceName);
		},

		correctnessGroups(): Group[] {
			return this.getAndActivateGroups(this.correctnessSummary, this.correctnessInstanceName);
		},

		residualGroups(): Group[] {
			const groups = this.getAndActivateGroups(this.residualsSummary, this.residualInstanceName);
			groups.forEach(group => {
				group.facets.forEach((facet: any) => {
					if (facet.histogram) {
						facet.histogram.showOrigin = true;
					}
				});
			});
			return groups;
		},

		highlights(): Highlight {
			return getHighlights();
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
		},

		isPending(): boolean {
			return this.solutionStatus !== SOLUTION_COMPLETED && this.solutionStatus !== SOLUTION_ERRORED;
		},

		isCompleted(): boolean {
			return this.solutionStatus === SOLUTION_COMPLETED;
		},

		isErrored(): boolean {
			return this.solutionStatus === SOLUTION_ERRORED;
		}

	},

	methods: {

		metricName(metric): string {
			return getMetricDisplayName(metric);
		},

		onResultCategoricalClick(context: string, key: string, value: string) {
			if (key && value) {
				// extract the var name from the key
				updateHighlightRoot(this.$router, {
					context: context,
					key: key,
					value: value
				});
			} else {
				clearHighlightRoot(this.$router);
			}
		},

		onCorrectnessCategoricalClick(context: string, key: string, value: string) {
			if (key && value) {
				// extract the var name from the key
				updateHighlightRoot(this.$router, {
					context: context,
					key: key,
					value: value
				});
			} else {
				clearHighlightRoot(this.$router);
			}
		},

		onResultNumericalClick(context: string, key: string, value: { from: number, to: number }) {
			if (!this.highlights.root || this.highlights.root.key !== key) {
				updateHighlightRoot(this.$router, {
					context: context,
					key: key,
					value: value
				});
			}
		},

		onResultRangeChange(context: string, key: string, value: { from: { label: string[] }, to: { label: string[] } }) {
			updateHighlightRoot(this.$router, {
				context: context,
				key: key,
				value: value
			});
			this.$emit('range-change', key, value);
		},

		onResidualNumericalClick(context: string, key: string, value: { from: number, to: number }) {
			if (!this.highlights.root || this.highlights.root.key !== key) {
				updateHighlightRoot(this.$router, {
					context: context,
					key: key,
					value: value
				});
			}
		},

		onResidualRangeChange(context: string, key: string, value: { from: number, to: number }) {
			updateHighlightRoot(this.$router, {
				context: context,
				key: key,
				value: value
			});
			this.$emit('range-change', key, value);
		},

		click() {
			if (this.predictedSummary && this.solutionId !== this.predictedSummary.solutionId) {
				const routeEntry = overlayRouteEntry(this.$route, {
					solutionId: this.predictedSummary.solutionId,
					highlights: null
				});
				this.$router.push(routeEntry);
			}
		},

		getAndActivateGroups(summary: VariableSummary, contextName: string): Group[] {
			if (summary) {
				const groups = createGroups([ summary ]);
				if (this.highlights.root && this.highlights.root.context === contextName) {
					const group = groups[0];
					if (group.key === this.highlights.root.key) {
						group.facets.forEach(facet => {
							facet.filterable = true;
						});
					}
				}
				return groups;
			}
			return [];
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
	position: relative;
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
</style>
