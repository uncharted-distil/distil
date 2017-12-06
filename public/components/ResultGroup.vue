<template>
	<div v-bind:class="currentClass"
		@click="click()">
		{{ name }}
		<facets class="result-container"
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
</template>

<script lang="ts">

// Component that contains a histogram of regression predictions, a histogram of the
// of prediction-truth residuals, and scoring information.

import Facets from '../components/Facets';
import { createGroups, Group } from '../util/facets';
import { VariableSummary, Dictionary } from '../store/data/index';
import { createRouteEntryFromRoute } from '../util/routes';
import { updateFilter } from '../util/filters';
import { getters } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { NUMERICAL_FILTER, CATEGORICAL_FILTER, NumericalFilter, CategoricalFilter, getFilterType, decodeFilters } from '../util/filters';
import { NumericalFacet, CategoricalFacet } from '../util/facets';
import Vue from 'vue';

export default Vue.extend({
	name: 'result-group',

	props: {
		name: String,
		resultSummary: Object,
		residualsSummary: Object,
		resultHtml: String,
		residualHtml: String
	},

	components: {
		Facets
	},

	computed: {
		residualsGroups(): Group[] {
			if (this.residuals()) {
				return createGroups([this.residuals()], false, false);
			}
			return [];
		},

		resultGroups(): Group[] {
			return createGroups([this.results()], false, false);
		},

		highlights(): Dictionary<any> {
			return getters.getHighlightedFeatureValues(this.$store);
		},

		currentClass(): string {
			const selectedResults = atob(routeGetters.getRouteResultId(this.$store));
			return (this.results().resultId === selectedResults)
				? 'result-group-selected result-group' : 'result-group';
		}
	},

	methods: {
		click() {
			const routeEntry = createRouteEntryFromRoute(this.$route, { resultId: btoa(this.results().resultId) });
			this.$router.push(routeEntry);
		},

		results(): VariableSummary {
			return <VariableSummary>this.resultSummary;
		},

		residuals(): VariableSummary {
			return <VariableSummary>this.residualsSummary;
		},

		updateFilterRoute(key: string, values: Dictionary<any>, resultUri: string) {
			// merge the updated filters back into the route query params if set
			const filters = routeGetters.getRouteResultFilters(this.$store);
			let updatedFilters = filters;
			if (key && values) {
				updatedFilters = updateFilter(filters, key, values);
			}

			const entry = createRouteEntryFromRoute(routeGetters.getRoute(this.$store), {
				resultId: resultUri ? btoa(resultUri) : routeGetters.getRouteResultId(this.$store),
				results: updatedFilters
			});

			this.$router.push(entry);
		},

		onRangeChange(key: string, value: { from: { label: string[] }, to: { label: string[] } }) {
			// set range filter
			this.updateFilterRoute(key, {
					enabled: true,
					min: parseFloat(value.from.label[0]),
					max: parseFloat(value.to.label[0])
				}, null);
		},

		updateGroupSelections(groups: Group[]): Group[] {
			const filters = routeGetters.getRouteResultFilters(this.$store);
			const decoded = decodeFilters(filters);
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
										from: `${(<NumericalFilter>filter).min}`,
										to: `${(<NumericalFilter>filter).max}`,
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
								if ((<CategoricalFilter>filter).categories.indexOf(categoricalFacet.value) !== -1) {
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
