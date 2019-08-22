<template>
	<div class='variable-facets row'>
		<div class="col-12 flex-column d-flex h-100">
			<div v-if="enableSearch" class="row align-items-center facet-filters">
				<div class="col-12 flex-column d-flex">
					<b-form-input size="sm" v-model="filter" placeholder="Search" />
				</div>
			</div>
			<!-- TODO: this should be passed in as title HTML -->
			<div v-if="enableTitle" class="row align-items-center">
				<div class="col-12 flex-column d-flex">
					<p><b>Select Feature to Predict</b> Select from potential features of interest below. Each feature tile shown summarizes count of records by value.</p>
				</div>
			</div>
			<div class="pl-1 pr-1">
				<!-- injectable slot -->
				<slot></slot>
			</div>
			<div class="row flex-1">
				<div class="col-12 flex-column variable-facets-container h-100">
					<div class="variable-facets-item" v-for="summary in paginatedSummaries" :key="summary.key">
						<template v-if="summary.varType === 'timeseries' || isTimeseriesAnalysis">
							<facet-timeseries
								:summary="summary"
								:highlight="highlight"
								:row-selection="rowSelection"
								:html="html"
								:enable-type-change="enableTypeChange"
								:enable-highlighting="[enableHighlighting, enableHighlighting]"
								:ignore-highlights="[ignoreHighlights, ignoreHighlights]"
								:instanceName="instanceName"
								@numerical-click="onNumericalClick"
								@categorical-click="onCategoricalClick"
								@range-change="onRangeChange"
								@histogram-numerical-click="onNumericalClick"
								@histogram-categorical-click="onCategoricalClick"
								@histogram-range-change="onRangeChange">
							</facet-timeseries>
						</template>
						<template v-else-if="summary.varType === 'geocoordinate'">
							<geocoordinate-facet :summary="summary"></geocoordinate-facet>
						</template>
						<template v-else>
							<facet-entry
								:summary="summary"
								:highlight="highlight"
								:row-selection="rowSelection"
								:html="html"
								:enable-type-change="enableTypeChange"
								:enable-highlighting="enableHighlighting"
								:ignore-highlights="ignoreHighlights"
								:instanceName="instanceName"
								@numerical-click="onNumericalClick"
								@categorical-click="onCategoricalClick"
								@range-change="onRangeChange"
								@facet-click="onFacetClick">
							</facet-entry>
						</template>
					</div>
				</div>
			</div>
			<div v-if="numSummaries > rowsPerPage" class="row align-items-center variable-page-nav">
				<div class="col-12 flex-column">
					<b-pagination size="sm" align="center" :total-rows="numSummaries" :per-page="rowsPerPage" v-model="currentPage" class="mb-0"/>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import FacetEntry from '../components/FacetEntry';
import FacetTimeseries from '../components/FacetTimeseries';
import GeocoordinateFacet from '../components/GeocoordinateFacet';
import { overlayRouteEntry, getRouteFacetPage } from '../util/routes';
import { Dictionary } from '../util/dict';
import { sortSummariesByImportance, filterVariablesByPage, getVariableImportance } from '../util/data';
import { Highlight, RowSelection, Variable, VariableSummary } from '../store/dataset/index';
import { getters as datasetGetters, actions as datasetActions } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';
import { ROUTE_PAGE_SUFFIX } from '../store/route/index';
import { Group } from '../util/facets';
import { LATITUDE_TYPE, LONGITUDE_TYPE } from '../util/types';

import { updateHighlight, clearHighlight } from '../util/highlights';
import Vue from 'vue';

export default Vue.extend({
	name: 'variable-facets',

	components: {
		FacetEntry,
		FacetTimeseries,
		GeocoordinateFacet
	},

	props: {
		enableSearch: Boolean as () => boolean,
		enableTitle: Boolean as () => boolean,
		enableTypeChange: Boolean as () => boolean,
		enableHighlighting: Boolean as () => boolean,
		ignoreHighlights: Boolean as () => boolean,
		summaries: Array as () => VariableSummary[],
		subtitle: String as () => string,
		html: [ String as () => string, Object as () => any, Function as () => Function ],
		instanceName: { type: String as () => string, default: 'variableFacets' },
		rowsPerPage: { type: Number as () => number, default: 10 },
	},

	data() {
		return {
			filter: ''
		};
	},

	computed: {

		currentPage: {
			set(page: number) {
				const entry = overlayRouteEntry(this.$route, {
					[this.routePageKey()]: page
				});
				this.$router.push(entry);
			},
			get(): number {
				return getRouteFacetPage(this.routePageKey(), this.$route);
			}
		},

		isTimeseriesAnalysis(): boolean {
			return !!routeGetters.getRouteTimeseriesAnalysis(this.$store);
		},

		variables(): Variable[] {
			return datasetGetters.getVariables(this.$store);
		},

		isGeocoordinateFacetAvailable(): boolean {
			const foo = datasetGetters.getGeocoordinateTypes(this.$store).length > 0;
			return datasetGetters.getGeocoordinateTypes(this.$store).length > 0;
		},

		filteredSummaries(): VariableSummary[] {
			return this.summaries.filter(summary => {
				return this.filter === '' || summary.key.toLowerCase().includes(this.filter.toLowerCase());
			});
		},

		sortedFilteredSummaries(): VariableSummary[] {
			return sortSummariesByImportance(this.filteredSummaries, this.variables);
		},

		paginatedSummaries(): VariableSummary[] {
			let filteredVariables = filterVariablesByPage(this.currentPage, this.rowsPerPage, this.sortedFilteredSummaries);
			if (this.isGeocoordinateFacetAvailable) {
				filteredVariables = filteredVariables.filter((variable) => {
					return variable.key !== LATITUDE_TYPE && variable.key !== LONGITUDE_TYPE;
				});
			}
			return filteredVariables;
		},

		numSummaries(): number {
			return this.filteredSummaries.length;
		},

		highlight(): Highlight {
			return routeGetters.getDecodedHighlight(this.$store);
		},

		rowSelection(): RowSelection {
			return routeGetters.getDecodedRowSelection(this.$store);
		},

		importance(): Dictionary<number> {
			const importance: Dictionary<number> = {};
			this.variables.forEach(variable => {
				importance[variable.colName] = getVariableImportance(variable);
			});
			return importance;
		}
	},

	methods: {

		// creates a facet key for the route from the instance-name component arg
		// or uses a default if unset
		routePageKey(): string {
			return `${this.instanceName}${ROUTE_PAGE_SUFFIX}`;
		},

		onRangeChange(context: string, key: string, value: { from: number, to: number }, dataset: string) {
			updateHighlight(this.$router, {
				context: context,
				dataset: dataset,
				key: key,
				value: value
			});
			this.$emit('range-change', key, value);
		},

		onFacetClick(context: string, key: string, value: string, dataset: string) {
			if (this.enableHighlighting) {
				if (key && value) {
					updateHighlight(this.$router, {
						context: context,
						dataset: dataset,
						key: key,
						value: value
					});
				} else {
					clearHighlight(this.$router);
				}
			}
			this.$emit('facet-click', context, key, value);
		},

		onCategoricalClick(context: string, key: string) {
			this.$emit('categorical-click', key);
		},

		onNumericalClick(context: string, key: string, value: { from: number, to: number }, dataset: string) {
			if (this.enableHighlighting) {
				if (!this.highlight || this.highlight.key !== key) {
					updateHighlight(this.$router, {
						context: this.instanceName,
						dataset: dataset,
						key: key,
						value: value
					});
				}
			}
			this.$emit('numerical-click', key);
		},

		availableVariables(): string[] {
			// NOTE: used externally, not internally by the component

			// filter by search
			const searchFiltered = this.summaries.filter(summary => {
				return this.filter === '' || summary.key.toLowerCase().includes(this.filter.toLowerCase());
			});
			return searchFiltered.map(v => v.key);
		}
	}
});

</script>

<style>
button {
	cursor: pointer;
}
.page-link {
	color: #868e96;
}
.page-item.active .page-link {
	z-index: 2;
	color: #fff;
	background-color: #868e96;
	border-color: #868e96;
}
.variable-facets-container {
	overflow-x: hidden;
	overflow-y: auto;
}
.variable-facets-container .variable-facets-item {
	margin: 2px 2px 4px 2px;
	min-height: 150px;
}
.variable-facets-container .facets-root-container .facets-group-container{
	background-color: inherit;
}
.variable-facets-container .facets-root-container .facets-group-container .facets-group {
	background: white;
	font-size: 0.867rem;
	color: rgba(0,0,0,0.87);
	box-shadow: 0 1px 2px 0 rgba(0,0,0,0.10);
	transition: box-shadow 0.3s ease-in-out;
}
.variable-facets-container .facets-root-container .facets-group-container .facets-group .group-header {
	padding: 4px 8px 6px 8px;
}
.variable-facets-container .facets-root-container .facets-group-container .facets-group .group-header .type-change-menu {
	float: right;
	margin-top: -4px;
	margin-right: -8px;
}
.facet-filters {
	margin: 0 -10px 4px -10px;
}
.facet-filters span {
	font-size: 0.9rem;
}
.variable-page-nav {
	padding-top: 10px;
}

.geocoordinate {
	max-width: 500px;
	height: 300px;
}

</style>
