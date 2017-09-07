<template>
	<div class='variable-facets'>
		<div>
			<div v-if="enableFilter==true">
				<b-form-fieldset horizontal label="Filter" :label-cols="3">
					<b-form-input v-model="filter" placeholder="Type to Search" />
				</b-form-fieldset>
			</div>
			<div v-if="enableToggle==true">
				<b-form-fieldset horizontal label="Toggle" :label-cols="3">
					<b-button variant="outline-secondary" @click="selectAll">All</b-button>
					<b-button variant="outline-secondary" @click="deselectAll">None</b-button>
				</b-form-fieldset>
			</div>
		</div>
		<facets class="facets-container"
			:groups="groups"
			:html="html"
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
	name: 'variable-facets',

	components: {
		Facets
	},

	props: [
		'enable-filter',
		'enable-toggle',
		'variables',
		'dataset',
		'html',
	],

	data() {
		return {
			filter: ''
		};
	},

	computed: {
		groups() {
			// filter by search
			const filtered = this.variables.filter(summary => {
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

	methods: {
		updateFilterRoute(key, values) {
			// retrieve the filters from the route
			const filters = this.$store.getters.getRouteFilters();
			const path = this.$store.getters.getRoutePath();
			// merge the updated filters back into the route query params
			const updated = updateFilter(filters, key, values);
			const entry = createRouteEntry(path, {
				target: this.$store.getters.getRouteTargetVariable(),
				training: this.$store.getters.getRouteTrainingVariables(),
				dataset: this.dataset,
				filters: updated,
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
button {
	cursor: pointer;
}
.variable-facets {
	display: flex;
	flex-direction: column;
	padding: 8px;
}

.facets-container {
	overflow-x: hidden;
	overflow-y: auto;
}
</style>
