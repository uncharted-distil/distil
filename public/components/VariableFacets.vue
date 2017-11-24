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
			<div v-if="enableSort">
				<b-form-fieldset size="sm" horizontal label="Sort" :label-cols="2">
					<div class="sort-groups">
						<span class="sort-group">
							importance
							<div class="sort-buttons">
								<b-button size="sm" variant="outline-secondary" @click="setSortMethod('importance-asc')">
									<i class="fa fa-sort-numeric-asc"></i>
								</b-button>
								<b-button size="sm" variant="outline-secondary" @click="setSortMethod('importance-desc')">
									<i class="fa fa-sort-numeric-desc"></i>
								</b-button>
							</div>
						</span>
						<span class="sort-group">
							alphanumeric
							<div class="sort-buttons">
								<b-button size="sm" variant="outline-secondary" @click="setSortMethod('alpha-asc')">
									<i class="fa fa-sort-alpha-asc"></i>
								</b-button>
								<b-button size="sm" variant="outline-secondary" @click="setSortMethod('alpha-desc')">
									<i class="fa fa-sort-alpha-desc"></i>
								</b-button>
							</div>
						</span>
						<!--
						<span class="sort-group">
							novelty
							<div class="sort-buttons">
								<b-button size="sm" variant="outline-secondary" @click="setSortMethod('novelty-asc')">
									<i class="fa fa-sort-amount-asc"></i>
								</b-button>
								<b-button size="sm" variant="outline-secondary" @click="setSortMethod('novelty-desc')">
									<i class="fa fa-sort-amount-desc"></i>
								</b-button>
							</div>
						</span>
						-->
					</div>
				</b-form-fieldset>
			</div>
		</div>
		<facets class="variable-facets-container"
			:groups="groups"
			:highlights="highlights"
			:html="html"
			:sort="sort"
			v-on:expand="onExpand"
			v-on:collapse="onCollapse"
			v-on:range-change="onRangeChange"
			v-on:facet-toggle="onFacetToggle">
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
import { getters as dataGetters } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { createGroups, Group } from '../util/facets';
import 'font-awesome/css/font-awesome.css';
import '../styles/spinner.css';
import Vue from 'vue';

export default Vue.extend({
	name: 'variable-facets',

	components: {
		Facets
	},

	props: {
		'enableSearch': Boolean,
		'enableToggle': Boolean,
		'enableSort': Boolean,
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
			let groups = createGroups(filtered, this.enableGroupCollapse, this.enableFacetFiltering, '');

			// update collapsed state
			groups = this.updateGroupCollapses(groups);

			// update selections
			return this.updateGroupSelections(groups);
		},
		highlights(): Dictionary<string> {
			return dataGetters.getHighlightedFeatureValues(this.$store);
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
		alphaAsc(a: { key: string }, b: { key: string }): number {
			const textA = a.key.toLowerCase();
			const textB = b.key.toLowerCase();
			return (textA <= textB) ? -1 : (textA > textB) ? 1 : 0;
		},
		alphaDesc(a: { key: string }, b: { key: string }): number {
			const textA = a.key.toLowerCase();
			const textB = b.key.toLowerCase();
			return (textA <= textB) ? 1 : (textA > textB) ? -1 : 0;
		},
		importanceAsc(a: { key: string }, b: { key: string }) {
			const importance = this.importance;
			return importance[a.key] - importance[b.key];
		},
		importanceDesc(a: { key: string }, b: { key: string }): number {
			const importance = this.importance;
			return importance[b.key] - importance[a.key];
		},
		noveltyAsc(a: { novelty: number }, b: { novelty: number }): number {
			return a.novelty - b.novelty;
		},
		noveltyDesc(a: { novelty: number }, b: { novelty: number }): number {
			return b.novelty - a.novelty;
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
		},

		// handles facet group transitions to inactive (grayed out, reduced visuals) state
		onCollapse(key) {
			// disable filter
			this.updateFilterRoute({
				name: key,
				type: EMPTY_FILTER,
				enabled: false
			});
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
		},

		setSortMethod(type: string) {
			switch (type) {
				case 'alpha-asc':
					this.sortMethod = 'alphaAsc';
					break;
				case 'alpha-desc':
					this.sortMethod = 'alphaDesc';
					break;
				case 'importance-asc':
					this.sortMethod = 'importanceAsc';
					break;
				case 'importance-desc':
					this.sortMethod = 'importanceDesc';
					break;
				case 'novelty-asc':
					this.sortMethod = 'noveltyAsc';
					break;
				case 'novelty-desc':
					this.sortMethod = 'noveltyDesc';
					break;
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
.sort-groups {
	display: flex;
	flex-direction: row;
	text-align: center;
}
.sort-groups .sort-group {
	display: flex;
	flex-direction: column;
	justify-content: center;
	width: 33%;
	font-size: 0.7rem;
	font-weight: bold;
}
.sort-buttons {
	display: flex;
	flex-direction: row;
	justify-content: center;
}
.sort-buttons > button {
	margin-right: 4px;
}

.variable-page-nav {
	margin-top: 10px;
}

</style>
