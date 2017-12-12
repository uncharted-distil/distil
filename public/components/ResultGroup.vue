<template>
	<div v-bind:class="currentClass"
		@click="click()">
		{{ name }}<br>Status: {{ pipelineStatus }}
		<div v-if="pipelineStatus === 'COMPLETED' || pipelineStatus === 'UPDATED'">
			<facets v-if="resultGroups.length" class="result-container"
				:groups="resultGroups"
				:highlights="highlights"
				:html="residualHtml">
			</facets>
			<facets v-if="residualsGroups.length" class="residual-container"
				:groups="residualsGroups"
				:highlights="highlights"
				:html="resultHtml">
			</facets>
		</div>
		<div v-if="pipelineStatus === 'COMPLETED'">
			<b-progress v-if="pipelineStatus !== 'COMPLETED'"
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
import { createGroups, Group } from '../util/facets';
import { VariableSummary } from '../store/data/index';
import { Dictionary } from '../util/dict';
import { createRouteEntryFromRoute } from '../util/routes';
import { updateFilter } from '../util/filters';
import { getters } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { getters as pipelineGetters } from '../store/pipelines/module';
import { NUMERICAL_FILTER, CATEGORICAL_FILTER, Filter, getFilterType, decodeFiltersDictionary } from '../util/filters';
import { NumericalFacet, CategoricalFacet } from '../util/facets';
import Vue from 'vue';

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
			return getters.getHighlightedFeatureValues(this.$store);
		},

		currentClass(): string {
			const selectedResults = routeGetters.getRouteResultId(this.$store);
			const results = this.results();
			return (results && results.resultId === selectedResults)
				? 'result-group-selected result-group' : 'result-group';
		}
	},

	methods: {
		click() {
			const routeEntry = createRouteEntryFromRoute(this.$route, {
				resultId: this.results().resultId
			});
			this.$router.push(routeEntry);
		},

		results(): VariableSummary {
			return <VariableSummary>this.resultSummary;
		},

		residuals(): VariableSummary {
			return <VariableSummary>this.residualsSummary;
		},

		updateFilterRoute(filter: Filter, resultUri: string) {
			// merge the updated filters back into the route query params if set
			const filters = routeGetters.getRouteResultFilters(this.$store);
			let updatedFilters = filters;
			if (filter) {
				updatedFilters = updateFilter(filters, filter);
			}
			const entry = createRouteEntryFromRoute(routeGetters.getRoute(this.$store), {
				resultId: resultUri ? resultUri : routeGetters.getRouteResultId(this.$store),
				results: updatedFilters
			});
			this.$router.push(entry);
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
