<template>
	<div class='row h-100 variable-facets'>
		<div class="col-12 flex-column d-flex">
			<div v-if="enableSearch" class="row flex-1 align-items-center facet-filters">
				<div class="col-12 flex-column d-flex">
					<b-form-fieldset size="sm" horizontal label="Filter" :label-cols="2">
						<b-form-input size="sm" v-model="filter" placeholder="Type to Search" />
					</b-form-fieldset>
				</div>
			</div>
			<div v-if="enableToggle" class="row flex-1 align-items-center facet-filters">
				<div class="col-12 flex-column d-flex">
					<b-form-fieldset size="sm" horizontal label="Toggle" :label-cols="2">
						<b-button size="sm" variant="outline-secondary" @click="selectAll">All</b-button>
						<b-button size="sm" variant="outline-secondary" @click="deselectAll">None</b-button>
					</b-form-fieldset>
				</div>
			</div>
			<div v-if="enableTitle" class="row flex-1 align-items-center">
				<div class="col-12 flex-column d-flex">
					<p>Select one of the following feature summaries showing count of records by feature value.</p>
				</div>
			</div>
			<div class="row flex-11">
				<facets class="col-12 flex-column d-flex variable-facets-container"
					:groups="groups"
					:filters="filters"
					:highlights="highlights"
					:html="html"
					:sort="sort"
					:type-change="typeChange"
					@click="onClick"
					@expand="onExpand"
					@collapse="onCollapse"
					@range-change="onRangeChange"
					@facet-toggle="onFacetToggle"
					@histogram-click="onHistogramClick"
					@facet-click="onFacetClick">
				</facets>
			</div>
			<div v-if="numRows > rowsPerPage" class="row flex-1 align-items-center variable-page-nav">
				<div class="col-12 flex-column">
					<b-pagination size="sm" align="center" :total-rows="numRows" :per-page="rowsPerPage" v-model="currentPage" class="mb-0"/>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts">

import Facets from '../components/Facets';
import { Filter, decodeFiltersDictionary, updateFilter, isDisabled, EMPTY_FILTER,
	createNumericalFilter, createCategoricalFilter, updateFilterRoute } from '../util/filters';
import { overlayRouteEntry, getRouteFacetPage } from '../util/routes';
import { Highlights, Range } from '../util/highlights';
import { Dictionary } from '../util/dict';
import { getters as dataGetters } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { createGroups, Group } from '../util/facets';
import { updateHighlightRoot, clearHighlightRoot, getHighlights } from '../util/highlights';
import 'font-awesome/css/font-awesome.css';
import '../styles/spinner.css';
import Vue from 'vue';

export default Vue.extend({
	name: 'variable-facets',

	components: {
		Facets
	},

	props: {
		enableSearch: Boolean,
		enableToggle: Boolean,
		enableTitle: Boolean,
		enableGroupCollapse: Boolean,
		enableFacetFiltering: Boolean,
		variables: Array,
		dataset: String,
		html: [ String, Object, Function ],
		instanceName: { type: String, default: 'variable-facets' },
		typeChange: Boolean
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
				const entry = overlayRouteEntry(this.$route, {
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
			const searchFiltered = this.variables.filter(summary => {
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
			return this.updateGroupCollapses(groups);
		},

		highlights(): Highlights {
			return getHighlights(this.$store);
		},

		filters(): Filter[] {
			return routeGetters.getDecodedFilters(this.$store);
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

		// handles facet group transition to active state
		onExpand(key: string) {
			// enable filter
			const filter = {
				name: key,
				type: EMPTY_FILTER,
				enabled: true
			};
			updateFilterRoute(this, filter);
			this.$emit('expand', key);
		},

		// handles facet group transitions to inactive (grayed out, reduced visuals) state
		onCollapse(key) {
		// disable filter
			const filter = {
				name: key,
				type: EMPTY_FILTER,
				enabled: false
			};
			updateFilterRoute(this, filter);
			this.$emit('collapse', key);
		},

		onRangeChange(key: string, value: { from: { label: string[] }, to: { label: string[] } }) {
			const filter = createNumericalFilter(key, value);
			updateFilterRoute(this, filter);
			this.$emit('range-change', key, value);
		},

		onFacetToggle(key: string, values: string[]) {
			const filter = createCategoricalFilter(key, values);
			updateFilterRoute(this, filter);
			this.$emit('facet-toggle', key, values);
		},

		onClick(key: string) {
			this.$emit('click', key);
		},

		onHistogramClick(context: string, key: string, value: Range) {
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

		onFacetClick(context: string, key: string, value: string) {
			if (key && value) {
				// extract the var name from the key
				updateHighlightRoot(this, {
					context: context,
					key: key,
					value: value
				});
			} else {
				clearHighlightRoot(this);
			}
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
			const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
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
			const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
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
.variable-facets-container .facets-root-container{
	margin: 2px;
}
.variable-facets-container .facets-root-container .facets-group-container{
	background-color: inherit;
}
.variable-facets-container .facets-root-container .facets-group-container .facets-group {
	background: white;
	margin: 2px 2px 4px 2px;
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
}
.facet-filters span {
	font-size: 0.9rem;
}
.facet-filters .form-group {
	margin: 0.25rem;
}
.variable-page-nav {
	padding-top: 10px;
}

</style>
