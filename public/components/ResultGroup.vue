<template>
	<div v-bind:class="currentClass"
		@click="click()">
		{{ name }}<!--<br>Status: {{ pipelineStatus }}-->
		<div v-if="pipelineStatus === 'COMPLETED' || pipelineStatus === 'UPDATED'">
			<facets v-if="resultGroups.length" class="result-container"
				v-on:histogram-mouse-enter="resultHistogramMouseEnter"
				v-on:histogram-mouse-leave="resultHistogramMouseLeave"
				v-on:facet-mouse-enter="resultFacetMouseEnter"
				v-on:facet-mouse-leave="resultFacetMouseLeave"
				:groups="resultGroups"
				:highlights="highlights"
				:html="residualHtml">
			</facets>
			<facets v-if="residualsGroups.length" class="residual-container"
				v-on:histogram-mouse-enter="residualsHistogramMouseEnter"
				v-on:histogram-mouse-leave="residualsHistogramMouseLeave"
				:groups="residualsGroups"
				:highlights="highlights"
				:html="resultHtml">
			</facets>
		</div>
		<div v-if="pipelineStatus !== 'COMPLETED'">
			<b-progress
				:value="100"
				variant="secondary"
				striped
				:animated="true"></b-progress>
		</div>
	</div>
</template>

<script lang="ts">

// Component that contains a histogram of regression predictions, a histogram of the
// of prediction-truth residuals, and scoring information.

import Facets from '../components/Facets';
import { createGroups, Group, NumericalFacet, CategoricalFacet } from '../util/facets';
import { isPredicted, isError, getVarFromPredicted, getVarFromError, getPredictedFacetKey,
	getErrorFacetKey, getPredictedColFromFacetKey, getErrorColFromFacetKey } from '../util/data';
import { VariableSummary } from '../store/data/index';
import { Dictionary } from '../util/dict';
import { overlayRouteEntry } from '../util/routes';
import { getters } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { getters as pipelineGetters } from '../store/pipelines/module';
import { mutations as dataMutations } from '../store/data/module';
import { NUMERICAL_FILTER, CATEGORICAL_FILTER, getFilterType, decodeFiltersDictionary } from '../util/filters';
import _ from 'lodash';
import Vue from 'vue';

const RESULT_GROUP_HIGHLIGHTS = 'result-group';

export default Vue.extend({
	name: 'result-group',

	props: {
		name: String,
		requestId: String,
		pipelineId: String,
		resultSummary: Object,
		residualsSummary: Object,
		resultHtml: String,
		residualHtml: String
	},

	components: {
		Facets
	},

	computed: {
		pipelineStatus(): String {
			const pipelines = pipelineGetters.getPipelines(this.$store);
			let pipeline = null;
			if (pipelines[this.requestId] && pipelines[this.requestId][this.pipelineId]) {
				pipeline = pipelines[this.requestId][this.pipelineId];
			}
			if (pipeline) {
				return pipeline.progress;
			}
			return 'unknown';
		},

		residualsGroups(): Group[] {
			if (this.residuals()) {
				return createGroups([this.residuals()], false, false);
			}
			return [];
		},

		resultGroups(): Group[] {
			if (this.results()) {
				return createGroups([this.results()], false, false);
			}
			return [];
		},

		highlights(): Dictionary<any> {
			// Facets highlights are keyed by name - map the published highligh
			// key to the facet key
			const highlights = getters.getHighlightedFeatureValues(this.$store);
			const facetHighlights = <Dictionary<any>>{};
			_.forEach(highlights.values, (value, varName) => {
				if (isPredicted(varName)) {
					facetHighlights[getPredictedFacetKey(getVarFromPredicted(varName))] = value;
				} else if (isError(varName)) {
					facetHighlights[getErrorFacetKey(getVarFromError(varName))] = value;
				}
			});
			return facetHighlights;
		},

		currentClass(): string {
			const selectedId = routeGetters.getRoutePipelineId(this.$store);
			const results = this.results();
			return (results && results.pipelineId === selectedId)
				? 'result-group-selected result-group' : 'result-group';
		}
	},

	methods: {
		resultHistogramMouseEnter(key: string, value: any) {
			// extract the var name from the key
			const varName = getPredictedColFromFacetKey(key);
			dataMutations.highlightFeatureRange(this.$store, {
				context: RESULT_GROUP_HIGHLIGHTS,
				ranges: {
					[varName]: {
						from: _.toNumber(value.label[0]),
						to: _.toNumber(value.toLabel[value.toLabel.length-1])
					}
				}
			});
		},

		resultHistogramMouseLeave(key: string) {
			const varName = getPredictedColFromFacetKey(key);
			dataMutations.clearFeatureHighlightRange(this.$store, varName);
		},

		residualsHistogramMouseEnter(key: string, value: any) {
			// convert the residual histogram key name into the proper variable ID
			const varName =getErrorColFromFacetKey(key);
			dataMutations.highlightFeatureRange(this.$store, {
				context: RESULT_GROUP_HIGHLIGHTS,
				ranges: {
					[varName]: {
						from: _.toNumber(value.label[0]),
						to: _.toNumber(value.toLabel[value.toLabel.length-1])
					}
				}
			});
		},

		residualsHistogramMouseLeave(key: string) {
			const varName = getErrorColFromFacetKey(key);
			dataMutations.clearFeatureHighlightRange(this.$store, varName);
		},

		resultFacetMouseEnter(key: string, value: any) {
			// extract the var name from the key
			const varName = getPredictedColFromFacetKey(key);
			dataMutations.highlightFeatureValues(this.$store, {
				context: RESULT_GROUP_HIGHLIGHTS,
				values: {
					[varName]: value
				}
			});
		},

		resultFacetMouseLeave(key: string) {
			dataMutations.clearFeatureHighlightValues(this.$store);
		},

		click() {
			const routeEntry = overlayRouteEntry(this.$route, {
				pipelineId: this.results().pipelineId
			});
			this.$router.push(routeEntry);
		},

		results(): VariableSummary {
			return <VariableSummary>this.resultSummary;
		},

		residuals(): VariableSummary {
			return <VariableSummary>this.residualsSummary;
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
	border-color: #03c6e1;
	box-shadow: 0 0 10px #03c6e1;
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
