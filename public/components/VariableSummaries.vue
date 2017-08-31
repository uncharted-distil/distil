<template>
	<div class='variable-summaries'>
		<div class="nav bg-faded rounded-top">
			<h6 class="nav-link">Summaries</h6>
		</div>
		<div class="nav row">
			<div class="col-md-6">
				<b-form-fieldset horizontal label="Filter" :label-cols="3">
					<b-form-input v-model="filter" placeholder="Type to Search" />
				</b-form-fieldset>
			</div>
			<div class="col-md-6">
				<b-form-fieldset horizontal label="Toggle" :label-cols="3">
					<b-button variant="outline-secondary" @click="selectAll">All</b-button>
					<b-button variant="outline-secondary" @click="deselectAll">None</b-button>
				</b-form-fieldset>
			</div>
		</div>
		<div v-if="groups.length===0">
			No results
		</div>
		<facets v-if="groups.length>0"
			:groups="groups"
			v-on:expand="onExpand"
			v-on:collapse="onCollapse"
			v-on:range-change="onRangeChange"
			v-on:facet-toggle="onFacetToggle"></facets>
	</div>
</template>

<script>

import Facets from '../components/Facets';
import { decodeFilter, updateFilter, getFilterType, isDisabled, CATEGORICAL_FILTER, NUMERICAL_FILTER } from '../util/filters';
import { createRouteEntry } from '../util/routes';
import 'font-awesome/css/font-awesome.css';
import '../styles/spinner.css';

const SPINNER_HTML = [
	'<div class="bounce1"></div>',
	'<div class="bounce2"></div>',
	'<div class="bounce3"></div>'].join('');

export default {
	name: 'variable-summaries',

	components: {
		Facets
	},

	data() {
		return {
			filter: ''
		};
	},

	computed: {
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		groups() {
			// get variable summaries
			const summaries = this.$store.getters.getVariableSummaries();
			// filter by search
			const filtered = summaries.filter(summary => {
				return this.filter === '' || summary.name.toLowerCase().includes(this.filter.toLowerCase());
			});
			// create the groups
			let groups = this.createGroups(filtered);
			// update collapsed state
			groups = this.updateGroupCollapses(groups);
			// update selections
			return this.updateGroupSelections(groups);
		}
	},

	mounted() {
		this.$store.dispatch('getVariableSummaries', this.dataset);
	},

	watch: {
		'$route.query.dataset'() {
			this.$store.dispatch('getVariableSummaries', this.dataset);
		}
	},

	methods: {
		updateFilterRoute(key, values) {
			// retrieve the filters from the route
			const filters = this.$store.getters.getRouteFilters();
			const path = this.$store.getters.getRoutePath();
			// merge the updated filters back into the route query params
			const updated = updateFilter(filters, key, values);
			const entry = createRouteEntry(path, {
				dataset: this.$store.getters.getRouteDataset(),
				filters: updated
			});
			this.$router.push(entry);
		},
		onExpand(key) {
			// enable filter
			this.updateFilterRoute(key, {
				enabled: true
			});
		},
		onCollapse(key) {
			// disable filter
			this.updateFilterRoute(key, {
				enabled: false
			});
		},
		onRangeChange(key, value) {
			// set range filter
			this.updateFilterRoute(key, {
				enabled: true,
				min: parseFloat(value.from.label[0]),
				max: parseFloat(value.to.label[0])
			});
		},
		onFacetToggle(key, values) {
			// set range filter
			this.updateFilterRoute(key, {
				enabled: true,
				categories: values
			});
		},
		selectAll() {
			// enable all filters
			let filters = this.$store.getters.getRouteFilters();
			this.groups.forEach(group => {
				filters = updateFilter(filters, group.key, {
					enabled: true
				});
			});
			const path = this.$store.getters.getRoutePath();
			const entry = createRouteEntry(path, {
				dataset: this.$store.getters.getRouteDataset(),
				filters: filters
			});
			this.$router.push(entry);
		},
		deselectAll() {
			// enable all filters
			let filters = this.$store.getters.getRouteFilters();
			this.groups.forEach(group => {
				filters = updateFilter(filters, group.key, {
					enabled: false
				});
			});
			const path = this.$store.getters.getRoutePath();
			const entry = createRouteEntry(path, {
				dataset: this.$store.getters.getRouteDataset(),
				filters: filters
			});
			this.$router.push(entry);
		},
		createErrorFacet(summary) {
			return {
				label: summary.name,
				key: summary.name,
				facets: [{
					placeholder: true,
					html: `<div>${summary.err}</div>`
				}]
			};
		},
		createPendingFacet(summary) {
			return {
				label: summary.name,
				key: summary.name,
				facets: [{
					placeholder: true,
					html: SPINNER_HTML
				}]
			};
		},
		createSummaryFacet(summary) {

			switch (summary.type) {

				case 'categorical':
					return {
						label: summary.name,
						key: summary.name,
						facets: summary.buckets.map(b => {
							return {
								value: b.key,
								count: b.count,
								selected: {
									count: b.count
								}
							};
						})
					};

				case 'numerical':
					return {
						label: summary.name,
						key: summary.name,
						facets: [
							{
								histogram: {
									slices: summary.buckets.map(b => {
										return {
											label: b.key,
											count: b.count
										};
									})
								}
							}
						]
					};
			}
			console.warn('unrecognized summary type', summary.type);
			return null;
		},
		createGroups(summaries) {
			return summaries.map(summary => {
				if (summary.err) {
					// create error facet
					return this.createErrorFacet(summary);
				}
				if (summary.pending) {
					// create pending facet
					return this.createPendingFacet(summary);
				}
				// create facet
				return this.createSummaryFacet(summary);
			}).filter(group => {
				// remove null groups
				return group;
			});
		},
		updateGroupCollapses(groups) {
			return groups.map(group => {
				// get filter
				const filter = this.$store.getters.getRouteFilter(group.key);
				// return if disabled
				group.collapsed = isDisabled(filter);
				return group;
			});
		},
		updateGroupSelections(groups) {
			return groups.map(group => {
				// get filter
				const filter = this.$store.getters.getRouteFilter(group.key);
				const decoded = decodeFilter(group.key, filter);
				switch (getFilterType(decoded)) {
					case NUMERICAL_FILTER:
						// add selection to facets
						group.facets.forEach(facet => {
							facet.selection = {
								// NOTE: the `from` / `to` values MUST be strings.
								range: {
									from: `${decoded.min}`,
									to: `${decoded.max}`,
								}
							};
						});
						break;

					case CATEGORICAL_FILTER:
						// add selection to facets
						group.facets.forEach(facet => {
							if (decoded.categories.indexOf(facet.value) !== -1) {
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
.variables-header {
	border: 1px solid #ccc;
}
#variable-facets {
	overflow-x: hidden;
	overflow-y: auto;
}
button {
	cursor: pointer;
}
</style>
