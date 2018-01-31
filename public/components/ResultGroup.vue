<template>
	<div v-bind:class="currentClass"
		@click="click()">
		{{name}} <sup>{{index}}</sup> {{timestamp}}
		<div v-if="pipelineStatus === 'COMPLETED' || pipelineStatus === 'UPDATED'">
			<b-badge variant="info" v-bind:key="`${score.metric}-${pipelineId}`" v-for="score in scores">
				{{metricName(score.metric)}}: {{score.value}}
			</b-badge>
			<facets v-if="resultGroups.length" class="result-container"
				@histogram-click="onResultHistogramClick"
				@facet-click="onResultFacetClick"
				:groups="resultGroups"
				:highlights="highlights"
				:filters="filters"
				:html="residualHtml">
			</facets>
			<facets v-if="residualsGroups.length" class="residual-container"
				@histogram-click="onResidualsHistogramClick"
				:groups="residualsGroups"
				:highlights="highlights"
				:filters="filters"
				:html="resultHtml">
			</facets>
		</div>
		<div v-if="pipelineStatus !== 'COMPLETED' && pipelineStatus !== 'ERRORED'">
			<b-badge variant="info">{{pipelineStatus}}</b-badge>
			<b-progress
				:value="100"
				variant="secondary"
				striped
				:animated="true"></b-progress>
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

import Facets from '../components/Facets';
import { createGroups, Group, NumericalFacet, CategoricalFacet } from '../util/facets';
import { isPredicted, isError, getVarFromPredicted, getVarFromError, getPredictedFacetKey,
	getErrorFacetKey, getErrorCol, getPredictedCol } from '../util/data';
import { VariableSummary } from '../store/data/index';
import { Highlights, Range, getHighlights } from '../util/highlights';
import { overlayRouteEntry } from '../util/routes';
import { Filter } from '../util/filters';
import { getters as routeGetters } from '../store/route/module';
import { getPipelineById, getMetricDisplayName } from '../util/pipelines';
import { NUMERICAL_FILTER, CATEGORICAL_FILTER, getFilterType, decodeFiltersDictionary } from '../util/filters';
import { updateHighlightRoot, clearHighlightRoot } from '../util/highlights';
import { Dictionary } from '../util/dict';
import _ from 'lodash';
import Vue from 'vue';

export default Vue.extend({
	name: 'result-group',

	props: {
		name: String,
		index: Number,
		timestamp: String,
		requestId: String,
		pipelineId: String,
		scores: Array,
		resultSummary: Object,
		residualsSummary: Object,
		resultExtrema: Object,
		residualExtrema: Object,
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

		pipelineStatus(): String {
			const pipeline = getPipelineById(this.$store.state.pipelineModule, this.pipelineId);
			if (pipeline) {
				return pipeline.progress;
			}
			return 'unknown';
		},

		residualsGroups(): Group[] {
			if (this.residuals()) {
				return createGroups([this.residuals()], false, false, this.residualExtrema);
			}
			return [];
		},

		resultGroups(): Group[] {
			if (this.results()) {
				return createGroups([this.results()], false, false, this.resultExtrema);
			}
			return [];
		},

		highlights(): Highlights {
			// Remap highlights to facet key names, filtering out anything other than
			// the predicted and error values (since that's all that is displayed in this
			// component)
			const highlights = getHighlights(this.$store);
			const facetHighlights = <Highlights>{
				root: _.cloneDeep(highlights.root),
				values: <Dictionary<string[]>>{}
			};
			_.forEach(highlights.values, (values, varName) => {
				if (isPredicted(varName)) {
					facetHighlights.values[getPredictedFacetKey(getVarFromPredicted(varName))] = values;
				} else if (isError(varName)) {
					facetHighlights.values[getErrorFacetKey(getVarFromError(varName))] = values;
				}
			});
			// Remap the selection root as well.
			if (highlights.root) {
				if (isPredicted(highlights.root.key)) {
					facetHighlights.root.key = 'Predicted';
				} else if (isError(highlights.root.key)) {
					facetHighlights.root.key = 'Error';
				}
			}
			return facetHighlights;
		},

		filters(): Filter[] {
			return routeGetters.getDecodedResultsFilters(this.$store);
		},

		currentClass(): string {
			const selectedId = routeGetters.getRoutePipelineId(this.$store);
			const results = this.results();
			return (results && results.pipelineId === selectedId)
				? 'result-group-selected result-group' : 'result-group';
		}
	},

	methods: {
		metricName(metric): string {
			return getMetricDisplayName(metric);
		},

		onResultHistogramClick(context: string, key: string, value: any) {
			const targetVar = routeGetters.getRouteTargetVariable(this.$store);
			this.histogramHighlights(context, key ? getPredictedCol(targetVar) : key, value);
		},

		onResidualsHistogramClick(context, key: string, value: any) {
			const targetVar = routeGetters.getRouteTargetVariable(this.$store);
			this.histogramHighlights(context, key ? getErrorCol(targetVar) : key, value);
		},

		histogramHighlights(context: string, key: string, value: Range) {
			if (key && value) {
				updateHighlightRoot(this, {
					context: context,
					key: key,
					value: value
				});
			} else {
				clearHighlightRoot(this);
			}
		},

		onResultFacetClick(context: string, key: string, value: string) {
			if (key && value) {
				// extract the var name from the key
				const targetVar = routeGetters.getRouteTargetVariable(this.$store);
				const varName = getPredictedCol(targetVar);
				updateHighlightRoot(this, {
					context: context,
					key: varName,
					value: value
				});
			} else {
				clearHighlightRoot(this);
			}
		},

		resultFacetMouseLeave(key: string) {
			clearHighlightRoot(this);
		},

		click() {
			if (this.results()) {
				const routeEntry = overlayRouteEntry(this.$route, {
					pipelineId: this.results().pipelineId
				});
				this.$router.push(routeEntry);
			}
		},

		results(): VariableSummary {
			if (this.resultSummary) {
				return this.resultSummary as VariableSummary;
			}
			return null;
		},

		residuals(): VariableSummary {
			if (this.residualsSummary) {
				return this.residualsSummary as VariableSummary;
			}
			return null;
		},

		updateGroupSelections(groups: Group[]): Group[] {
			const filters = routeGetters.getRouteResultFilters(this.$store);
			const decoded = decodeFiltersDictionary(filters);
			return groups.map(group => {
				// get filter
				const filter = decoded[group.key];
				switch (getFilterType(filter)) {
					case NUMERICAL_FILTER:
						// add selection to facets
						group.facets.forEach(facet => {
							if ((<NumericalFacet>facet).selection) {
								(<NumericalFacet>facet).selection = {
									// NOTE: the `from` / `to` values MUST be strings.
									range: {
										from: `${filter.min}`,
										to: `${filter.max}`,
									}
								};
							}
						});
						break;

					case CATEGORICAL_FILTER:
						// add selection to facets
						group.facets.forEach(facet => {
							if ((<CategoricalFacet>facet).value) {
								const categoricalFacet = <CategoricalFacet>facet;
								if (filter.categories.indexOf(categoricalFacet.value) !== -1) {
									// select
									categoricalFacet.selected = {
										count: categoricalFacet.count
									};
								} else {
									delete categoricalFacet.selected;
								}
							}
						});
						break;
				}
				return group;
			});
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
