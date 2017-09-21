<template>
	<div class='result-facets'>
		<facets class="facets-container" :groups="groups" :html="html" v-on:expand="onExpand" v-on:collapse="onCollapse" v-on:range-change="onRangeChange" v-on:facet-toggle="onFacetToggle"></facets>
	</div>
</template>

<script>

import _ from 'lodash';
import Facets from '../components/Facets';
import { decodeFilters, updateFilter, getFilterType, isDisabled, CATEGORICAL_FILTER, NUMERICAL_FILTER } from '../util/filters';
import { createRouteEntryFromRoute } from '../util/routes';
import { spinnerHTML } from '../util/spinner';
import 'font-awesome/css/font-awesome.css';
import '../styles/spinner.css';

export default {
	name: 'result-facets',

	components: {
		Facets
	},

	props: [
		'variables',
		'dataset',
		'html'
	],

	computed: {
		groups() {
			// create the groups
			let groups = this.createGroups(this.variables);

			// sort alphabetically
			groups.sort((a, b) => {
				const textA = a.key.toLowerCase();
				const textB = b.key.toLowerCase();
				return (textA < textB) ? -1 : (textA > textB) ? 1 : 0;
			});

			// find pipeline result with the uri specified in the route and
			// flag it as the currently active result
			const requestId = this.$store.getters.getRouteCreateRequestId();
			const pipelineResults = this.$store.getters.getPipelineResults(requestId);
			const activeResult = _.find(pipelineResults, p => {
				return p.pipeline.resultUri === atob(this.$store.getters.getRouteResultId());
			});

			const filters = this.$store.getters.getRouteResultFilters();

			// if filters are empty this is the first group call - initialize
			// filter and group state
			if (_.isEmpty(filters)) {
				// set the selected value to the route value
				groups.forEach((group) => {
					if (group.key !== activeResult.name) {
						this.updateFilterRoute({
							key: group.key,
							values: {
								enabled: false
							}
						});
					} else {
						this.updateFilterRoute({
							key: group.key,
							values: {
								enabled: true
							}
						}, activeResult.pipeline.resultUri);
					}
				});
			}
			// update collapsed state
			groups = this.updateGroupCollapses(groups);
			// update selections
			return this.updateGroupSelections(groups);
		}

	},

	methods: {
		updateFilterRoute(filterArgs, resultUri) {
			const path = this.$store.getters.getRoutePath();

			// merge the updated filters back into the route query params if set
			const filters = this.$store.getters.getRouteResultFilters(); ;
			let updatedFilters = filters;
			if (filterArgs) {
				updatedFilters = updateFilter(filters, filterArgs.key, filterArgs.values);
			}

			const entry = createRouteEntryFromRoute(this.$store.getters.getRoute(), {
				resultId: resultUri ? btoa(resultUri) : this.$store.getters.getRouteResultId(),
				results: updatedFilters
			});

			this.$router.push(entry);
		},

		onExpand(key) {

			const createReqId = this.$store.getters.getRouteCreateRequestId();
			const pipelineRequests = this.$store.getters.getPipelineResults(createReqId);
			const completedReq = _.find(pipelineRequests, p => p.name === key);

			// disable all filters except this one
			this.groups.forEach(group => {
				if (group.key !== key) {
					this.updateFilterRoute({
						key: group.key,
						values: {
							enabled: false
						}
					});
				}
			});

			// enable filter
			this.updateFilterRoute({
				key: key,
				values: {
					enabled: true
				}
			}, completedReq.pipeline.resultUri);
		},
		onCollapse() {
			// TODO: prevent disabling?
			// no-op
		},
		onRangeChange(key, value) {
			// set range filter
			this.updateFilterRoute({
				key: key,
				values: {
					enabled: true,
					min: parseFloat(value.from.label[0]),
					max: parseFloat(value.to.label[0])
				}
			});
		},
		onFacetToggle(key, values) {
			// set range filter
			this.updateFilterRoute({
				key: key,
				values: {
					enabled: true,
					categories: values
				}
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
			const filters = this.$store.getters.getRouteResultFilters();
			const decoded = decodeFilters(filters);
			return groups.map(group => {
				// return if disabled
				group.collapsed = isDisabled(decoded[group.key]);
				return group;
			});
		},
		updateGroupSelections(groups) {
			const filters = this.$store.getters.getRouteResultFilters();
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
</style>
