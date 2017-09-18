<template>
	<div class='variable-facets'>
		<div class="facet-filters">
			<div v-if="enableFilter">
				<b-form-fieldset size="sm" horizontal label="Filter" :label-cols="2">
					<b-form-input size="sm" v-model="filter" placeholder="Type to Search" />
				</b-form-fieldset>
			</div>
			<div v-if="enableToggle">
				<b-form-fieldset size="sm" horizontal label="Toggle" :label-cols="2">
					<b-button size="sm" variant="outline-secondary" @click="selectAll">All</b-button>
					<b-button size="sm" variant="outline-secondary" @click="deselectAll">None</b-button>
				</b-form-fieldset>
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
					</div>
				</b-form-fieldset>
			</div>
		</div>
		<facets class="facets-container"
			:groups="groups"
			:html="html"
			:sort="sort"
			v-on:expand="onExpand"
			v-on:collapse="onCollapse"
			v-on:range-change="onRangeChange"
			v-on:facet-toggle="onFacetToggle"></facets>
	</div>
</template>

<script>

import Facets from '../components/Facets';
import { decodeFilters, updateFilter, getFilterType, isDisabled, CATEGORICAL_FILTER, NUMERICAL_FILTER } from '../util/filters';
import { createRouteEntry } from '../util/routes';
import { spinnerHTML } from '../util/spinner';
import 'font-awesome/css/font-awesome.css';
import '../styles/spinner.css';

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
		'html'
	],

	data() {
		return {
			filter: '',
			sortMethod: 'alphaAsc'
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
			return (textA < textB) ? -1 : (textA > textB) ? 1 : 0;
		},
		alphaDesc(a, b) {
			const textA = a.key.toLowerCase();
			const textB = b.key.toLowerCase();
			return (textA < textB) ? 1 : (textA > textB) ? -1 : 0;
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
					html: spinnerHTML()
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
			const filters = this.$store.getters.getRouteFilters();
			const decoded = decodeFilters(filters);
			return groups.map(group => {
				// return if disabled
				group.collapsed = isDisabled(decoded[group.key]);
				return group;
			});
		},
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
.sort-groups {
	display: flex;
	flex-direction: row;
	justify-content: space-between;
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
</style>
