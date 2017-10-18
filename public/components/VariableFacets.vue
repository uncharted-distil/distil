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

<script>

import Facets from '../components/Facets';
import { decodeFilters, updateFilter, getFilterType, isDisabled, CATEGORICAL_FILTER, NUMERICAL_FILTER } from '../util/filters';
import { createRouteEntryFromRoute } from '../util/routes';
import { createGroups } from '../util/facets';
import 'font-awesome/css/font-awesome.css';
import '../styles/spinner.css';

export default {
	name: 'variable-facets',

	components: {
		Facets
	},

	props: {
		'enable-search': Boolean,
		'enable-toggle': Boolean,
		'enable-sort': Boolean,
		'enable-group-collapse': Boolean,
		'enable-facet-filtering': Boolean,
		'variables': Array,
		'dataset': String,
		'html': [String, Object, Function],
		'instance-name': String
	},

	data() {
		return {
			filter: '',
			numRows: 1,
			rowsPerPage: 10,
			sortMethod: 'alphaAsc'
		};
	},

	computed: {
		currentPage: {
			set(page) {
				const entry = createRouteEntryFromRoute(this.$route, {
					[this.routePageKey()]: page
				});
				this.$router.push(entry);
			},
			get() {
				const routeFacetPage = this.$store.getters.getRouteFacetsPage(this.routePageKey());
				return routeFacetPage ? parseInt(routeFacetPage) : 1;
			}
		},
		groups() {
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
			groups = this.updateGroupCollapses(groups);
			// update selections
			return this.updateGroupSelections(groups);
		},
		highlights() {
			return this.$store.getters.getHighlightedFeatureValues();
		},
		importance() {
			const variables = this.$store.getters.getVariables();
			const importance = {};
			variables.forEach(variable => {
				importance[variable.name] = variable.importance;
			});
			return importance;
		},
		sort() {
			return this[this.sortMethod];
		}
	},

	methods: {
		alphaAsc(a, b) {
			const textA = a.key.toLowerCase();
			const textB = b.key.toLowerCase();
			return (textA <= textB) ? -1 : (textA > textB) ? 1 : 0;
		},
		alphaDesc(a, b) {
			const textA = a.key.toLowerCase();
			const textB = b.key.toLowerCase();
			return (textA <= textB) ? 1 : (textA > textB) ? -1 : 0;
		},
		importanceAsc(a, b) {
			const importance = this.importance;
			return importance[a.key] - importance[b.key];
		},
		importanceDesc(a, b) {
			const importance = this.importance;
			return importance[b.key] - importance[a.key];
		},
		noveltyAsc(a, b) {
			return a.novelty - b.novelty;
		},
		noveltyDesc(a, b) {
			return b.novelty - a.novelty;
		},

		// creates a facet key for the route from the instance-name component arg
		// or uses a default if unset
		routePageKey() {
			if (this.instanceName) {
				return `${this.instanceName}Page`;
			}
			return 'facetPage';
		},

		// updates route with current filter state
		updateFilterRoute(key, values) {
			// retrieve the filters from the route
			const filters = this.$store.getters.getRouteFilters();
			// merge the updated filters back into the route query params
			const updated = updateFilter(filters, key, values);
			const entry = createRouteEntryFromRoute(this.$store.getters.getRoute(), {
				filters: updated
			});
			this.$router.push(entry);
		},

		// handles facet group transition to active state
		onExpand(key) {
			// enable filter
			this.updateFilterRoute(key, {
				enabled: true
			});
		},

		// handles facet group transitions to inactive (grayed out, reduced visuals) state
		onCollapse(key) {
			// disable filter
			this.updateFilterRoute(key, {
				enabled: false
			});
		},

		// handles range slider change events
		onRangeChange(key, value) {
			// set range filter
			this.updateFilterRoute(key, {
				enabled: true,
				min: parseFloat(value.from.label[0]),
				max: parseFloat(value.to.label[0])
			});
		},

		// handles individual category toggle events within a facet group
		onFacetToggle(key, values) {
			// set range filter
			this.updateFilterRoute(key, {
				enabled: true,
				categories: values
			});
		},
		setSortMethod(type) {
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
			let filters = this.$store.getters.getRouteFilters();
			this.groups.forEach(group => {
				filters = updateFilter(filters, group.key, {
					enabled: true
				});
			});
			const entry = createRouteEntryFromRoute(this.$store.getters.getRoute(), {
				filters: filters
			});
			this.$router.push(entry);
		},

		// sets all facet groups to the inactive state - minimized diplay , no controls,
		// and updates route accordingly
		deselectAll() {
			// enable all filters
			let filters = this.$store.getters.getRouteFilters();
			this.groups.forEach(group => {
				filters = updateFilter(filters, group.key, {
					enabled: false
				});
			});
			const entry = createRouteEntryFromRoute(this.$store.getters.getRoute(), {
				filters: filters
			});
			this.$router.push(entry);
		},

		// updates facet collapse/expand state based on route settings
		updateGroupCollapses(groups) {
			const filters = this.$store.getters.getRouteFilters();
			const decoded = decodeFilters(filters);
			return groups.map(group => {
				// return if disabled
				group.collapsed = isDisabled(decoded[group.key]);
				return group;
			});
		},

		// updates numerical facet range controls or categorical selected state based on
		// route
		updateGroupSelections(groups) {
			const filters = this.$store.getters.getRouteFilters();
			const decoded = decodeFilters(filters);
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
									to: `${filter.max}`
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
};
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
.facet-filters .form-group {
	margin-bottom: 4px;
	padding-right: 16px;
}
.facet-filters label {
	font-size: 0.8rem;
	font-weight: bold;
}
.sort-groups {
	display: flex;
	flex-direction: row;
	text-align: center;
}
.sort-group {
	display: flex;
	flex-direction: column;
	justify-content: center;
	width: 33%;
	font-size: 12px;
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
