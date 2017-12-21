<template>
	<div class='variable-facets'>
		<div class="facet-filters">
			<div v-if="enableSearch">
				<b-form-fieldset size="sm" horizontal label="Filter" :label-cols="2">
					<b-form-input size="sm" v-model="filter" placeholder="Type to Search" />
				</b-form-fieldset>
			</div>
			<div v-if="enableToggle">
				<b-form-fieldset size="sm" horizontal label="Toggle" :label-cols="2">
					<b-button size="sm" variant="outline-secondary" @click="selectAll">All</b-button>
					<b-button size="sm" variant="outline-secondary" @click="deselectAll">None</b-button>
				</b-form-fieldset>
			</div>
		</div>
		<facets class="variable-facets-container"
			:groups="groups"
			:highlights="highlights"
			:html="html"
			:sort="sort"
			v-on:click="onClick"
			v-on:expand="onExpand"
			v-on:collapse="onCollapse"
			v-on:range-change="onRangeChange"
			v-on:facet-toggle="onFacetToggle"
			v-on:histogram-mouse-enter="onHistogramMouseEnter"
			v-on:histogram-mouse-leave="onHistogramMouseLeave"
			v-on:facet-mouse-enter="onFacetMouseEnter"
			v-on:facet-mouse-leave="onFacetMouseLeave">
		</facets>
		<div v-if="numRows > rowsPerPage" class="variable-page-nav">
			<b-pagination size="sm" align="center" :total-rows="numRows" :per-page="rowsPerPage" v-model="currentPage"/>
		</div>

	</div>
</template>

<script lang="ts">

import Facets from '../components/Facets';
import { Filter, decodeFiltersDictionary, updateFilter, getFilterType, isDisabled, CATEGORICAL_FILTER, NUMERICAL_FILTER, EMPTY_FILTER } from '../util/filters';
import { createRouteEntryFromRoute, getRouteFacetPage } from '../util/routes';
import { VariableSummary } from '../store/data/index';
import { Dictionary } from '../util/dict';
import { getters as dataGetters, mutations as dataMutations } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { createGroups, Group } from '../util/facets';
import 'font-awesome/css/font-awesome.css';
import '../styles/spinner.css';
import _ from 'lodash';
import Vue from 'vue';

const VARIABLE_FACET_HIGHLIGHTS = 'variable_facets';

export default Vue.extend({
	name: 'variable-facets',

	components: {
		Facets
	},

	props: {
		'enableSearch': Boolean,
		'enableToggle': Boolean,
		'enableGroupCollapse': Boolean,
		'enableFacetFiltering': Boolean,
		'variables': Array,
		'dataset': String,
		'html': [ String, Object, Function ],
		'instanceName': String
	},

	data() {
		return {
			filter: '',
			numRows: 1,
			rowsPerPage: 10,
			sortMethod: 'importanceDesc'
		};
	},

	computed: {
		currentPage: {
			set(page: number) {
				const entry = createRouteEntryFromRoute(this.$route, {
					[this.routePageKey()]: page
				});
				this.$router.push(entry);
			},
			get(): number {
				return getRouteFacetPage(this.routePageKey(), this.$route);
			}
		},

		groups(): Group[] {
			// filter by search
			const searchFiltered = (<VariableSummary[]>this.variables).filter(summary => {
				return this.filter === '' || summary.name.toLowerCase().includes(this.filter.toLowerCase());
			});

			// sort by current function - sort looks for key to hold sort key
			// TODO: this only needs to happen on re-order events once it has been sorted initially
			const sorted = searchFiltered.map(v => ({ key: v.name, variable: v }))
				.sort((a, b) => this[this.sortMethod](a, b))
				.map(v => v.variable);

			// if necessary, refilter applying pagination rules
			this.numRows = searchFiltered.length;
			let filtered = sorted;
			if (this.numRows > this.rowsPerPage) {
				const firstIndex = this.rowsPerPage * (this.currentPage - 1);
				const lastIndex = Math.min(firstIndex + this.rowsPerPage, this.numRows);
				filtered = sorted.slice(firstIndex, lastIndex);
			}

			// create the groups
			let groups = createGroups(filtered, this.enableGroupCollapse, this.enableFacetFiltering);

			// update collapsed state
			groups = this.updateGroupCollapses(groups);

			// update selections
			return this.updateGroupSelections(groups);
		},

		highlights(): Dictionary<any> {
			return dataGetters.getHighlightedFeatureValues(this.$store).values;
		},

		importance(): Dictionary<number> {
			const variables = dataGetters.getVariables(this.$store);
			const importance: Dictionary<number> = {};
			variables.forEach(variable => {
				importance[variable.name] = variable.importance;
			});
			return importance;
		},

		sort() {
			return (<any>this)[(<any>this).sortMethod];
		}
	},

	methods: {
		importanceDesc(a: { key: string }, b: { key: string }): number {
			const importance = this.importance;
			return importance[b.key] - importance[a.key];
		},

		// creates a facet key for the route from the instance-name component arg
		// or uses a default if unset
		routePageKey(): string {
			if (this.instanceName) {
				return `${this.instanceName}Page`;
			}
			return 'facetPage';
		},

		// updates route with current filter state
		updateFilterRoute(filter: Filter) {
			// retrieve the filters from the route
			const filters = routeGetters.getRouteFilters(this.$store);
			// merge the updated filters back into the route query params
			const updated = updateFilter(filters, filter);
			const entry = createRouteEntryFromRoute(routeGetters.getRoute(this.$store), {
				filters: updated,
			});
			this.$router.push(entry);
		},

		// handles facet group transition to active state
		onExpand(key: string) {
			// enable filter
			this.updateFilterRoute({
				name: key,
				type: EMPTY_FILTER,
				enabled: true
			});
			this.$emit('expand', key);
		},

		// handles facet group transitions to inactive (grayed out, reduced visuals) state
		onCollapse(key) {
			// disable filter
			this.updateFilterRoute({
				name: key,
				type: EMPTY_FILTER,
				enabled: false
			});
			this.$emit('collapse', key);
		},

		// handles range slider change events
		onRangeChange(key: string, value: { from: { label: string[] }, to: { label: string[] } }) {
			// set range filter
			this.updateFilterRoute({
				name: key,
				type: NUMERICAL_FILTER,
				enabled: true,
				min: parseFloat(value.from.label[0]),
				max: parseFloat(value.to.label[0])
			});
			this.$emit('range-change', key, value);
		},

		// handles individual category toggle events within a facet group
		onFacetToggle(key: string, values: string[]) {
			// set range filter
			this.updateFilterRoute({
				name: key,
				type: CATEGORICAL_FILTER,
				enabled: true,
				categories: values
			});
			this.$emit('facet-toggle', key, values);
		},

		onClick(key: string) {
			this.$emit('click', key);
		},

		onHistogramMouseEnter(key: string, value: any) {
			// extract the var name from the key
			dataMutations.highlightFeatureRange(this.$store, {
				context: VARIABLE_FACET_HIGHLIGHTS,
				ranges: {
					[key]: {
						from: _.toNumber(value.label[0]),
						to: _.toNumber(value.toLabel[value.toLabel.length-1])
					}
				}
			});
		},

		onHistogramMouseLeave(key: string) {
			dataMutations.clearFeatureHighlightRange(this.$store, key);
		},

		onFacetMouseEnter(key: string, value: any) {
			// extract the var name from the key
			dataMutations.highlightFeatureValues(this.$store, {
				context: VARIABLE_FACET_HIGHLIGHTS,
				values: {
					[key]: value
				}
			});
		},

		onFacetMouseLeave(key: string) {
			dataMutations.clearFeatureHighlightValues(this.$store);
		},

		// sets all facet groups to the active state - full size display + all controls, updates
		// route accordingly
		selectAll() {
			// enable all filters
			let filters = routeGetters.getRouteFilters(this.$store);
			this.groups.forEach(group => {
				filters = updateFilter(filters, {
					name: group.key,
					type: EMPTY_FILTER,
					enabled: true
				});
			});
			const entry = createRouteEntryFromRoute(routeGetters.getRoute(this.$store), {
				filters: filters,
			});
			this.$router.push(entry);
		},

		// sets all facet groups to the inactive state - minimized diplay , no controls,
		// and updates route accordingly
		deselectAll() {
			// enable all filters
			let filters = routeGetters.getRouteFilters(this.$store);
			this.groups.forEach(group => {
				filters = updateFilter(filters, {
					name: group.key,
					type: EMPTY_FILTER,
					enabled: false
				});
			});
			const entry = createRouteEntryFromRoute(routeGetters.getRoute(this.$store), {
				filters: filters
			});
			this.$router.push(entry);
		},

		// updates facet collapse/expand state based on route settings
		updateGroupCollapses(groups: Group[]): Group[] {
			const filters = routeGetters.getRouteFilters(this.$store);
			const decoded = decodeFiltersDictionary(filters);
			return groups.map(group => {
				// return if disabled
				group.collapsed = isDisabled(decoded[group.key]);
				return group;
			});
		},

		// updates numerical facet range controls or categorical selected state based on
		// route
		updateGroupSelections(groups): Group[] {
			const filters = routeGetters.getRouteFilters(this.$store);
			const decoded = decodeFiltersDictionary(filters);
			return groups.map(group => {
				// get filter
				const filter = decoded[group.key];
				switch (getFilterType(filter)) {
					case NUMERICAL_FILTER:
						// add selection to facets
						group.facets.forEach(facet => {
							facet.selection = {
								// NOTE: the `from` / `to` values MUST be strings.
								range: {
									from: `${filter.min}`,
									to: `${filter.max}`,
								}
							};
						});
						break;

					case CATEGORICAL_FILTER:
						// add selection to facets
						group.facets.forEach(facet => {
							if (filter.categories.indexOf(facet.value) !== -1) {
								// select
								facet.selected = {
									count: facet.count
								};
							} else {
								delete facet.selected;
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
button {
	cursor: pointer;
}
.variable-facets {
	display: flex;
	flex-direction: column;
	padding-right: 8px;
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
.facet-filters span {
	font-size: 0.9rem;
}
.facet-filters .form-group {
	margin-bottom: 4px;
	padding-right: 16px;
}

.variable-page-nav {
	margin-top: 10px;
}

</style>
