<template>
	<div class='variable-summaries'>
		<div class="nav bg-faded rounded-top">
			<h6 class="nav-link">Summaries</h6>
		</div>
		<div v-if="groups.length===0">
			No results
		</div>
		<facets v-if="groups.length>0"
			:groups="groups"
			v-on:expand="onExpand"
			v-on:collapse="onCollapse"
			v-on:range-change="onRangeChange"></facets>
	</div>
</template>

<script>

import _ from 'lodash';

import Facets from '../components/Facets';
import { decodeFilter, updateFilter, getFilterType, isDisabled, NUMERICAL_FILTER } from '../util/filters';
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

	computed: {
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		groups() {
			// get variable summaries
			const summaries = this.$store.getters.getVariableSummaries();
			// create the groups
			let groups = this.createGroups(summaries);
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
			// update the filters
			const updated = updateFilter(filters, key, values);
			// merge the updated filters back into the route query params
			this.$router.push({
				path: '/dataset',
				query: _.merge({
					dataset: this.$store.getters.getRouteDataset(),
					terms: this.$store.getters.getRouteTerms(),
				}, updated)
			});
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
								count: b.count
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
				const decoded = decodeFilter(filter);
				// check if numeric filter
				if (getFilterType(decoded) === NUMERICAL_FILTER) {
					// add selection to facets
					group.facets.forEach(facet => {
						facet.selection = {
							range: {
								from: decoded.min,
								to: decoded.max
							}
						};
					});
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
</style>
