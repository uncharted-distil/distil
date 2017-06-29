<template>
	<div class='variable-summaries'>
		<div class="nav bg-faded rounded-top">
			<h6 class="nav-link">Summaries</h6>
		</div>
		<div v-if="summaries.length===0">
			No results
		</div>
		<div id="variable-facets"></div>
	</div>
</template>

<script>

import _ from 'lodash';

import Facets from '@uncharted.software/stories-facets';
import { decodeFilter, updateFilter, getFilterType, isDisabled, NUMERICAL_FILTER } from '../util/filters';
import '@uncharted.software/stories-facets/dist/facets.css';
import 'font-awesome/css/font-awesome.css';
import '../styles/spinner.css';

export default {
	name: 'variable-summaries',

	data() {
		return {
			facets: null,
			groups: new Map(),
			pending: new Map(),
			errors: new Map()
		};
	},

	computed: {
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		summaries() {
			return this.$store.getters.getVariableSummaries();
		}
	},

	mounted() {
		const component = this;

		this.$store.dispatch('getVariableSummaries', this.dataset);

		// instantiate the external facets widget
		const container = document.getElementById('variable-facets');
		this.facets = new Facets(container, []);

		// handle a facet going from collapsed to expanded by updating the state in
		// the store
		this.facets.on('facet-group:expand', (evt, key) => {
			// enable filter
			component.updateFilterRoute(key, {
				enabled: true
			});
		});

		// handle a facet going from expanded to collapsed by updating the state in
		// the store
		this.facets.on('facet-group:collapse', (evt, key) => {
			// disable filter
			component.updateFilterRoute(key, {
				enabled: false
			});
		});

		// handle a facet changing its filter range by updating the store
		this.facets.on(' facet-histogram:rangechangeduser', (evt, key, value) => {
			// set range filter
			component.updateFilterRoute(key, {
				enabled: true,
				min: parseFloat(value.from.label[0]),
				max: parseFloat(value.to.label[0])
			});
		});

		// update it's contents when the dataset changes
		this.$store.watch(() => this.$store.state.variableSummaries, (histograms) => {
			const bulk = [];
			// for each histogram
			histograms.forEach(histogram => {
				if (histogram.err) {
					// create error facet
					this.createErrorFacet(bulk, histogram);
					return;
				}
				if (histogram.pending) {
					// create pending facet
					this.createPendingFacet(bulk, histogram);
					return;
				}
				// create facet
				this.createFacet(bulk, histogram);
			});
			// add created facets
			if (bulk.length > 0) {
				this.facets.replace(bulk);
			}
		});
	},

	watch: {
		'$route.query.dataset'() {
			this.groups.clear();
			this.pending.clear();
			this.errors.clear();
			this.facets.replace([]);
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
		getHistogramKey(histogram) {
			return histogram.name;
		},
		createErrorFacet(bulk, histogram) {
			const key = this.getHistogramKey(histogram);
			// check if already added as error
			if (this.errors.has(key)) {
				return;
			}
			// add error group
			const group = {
				label: histogram.name,
				key: key,
				facets: [{
					placeholder: true,
					key: 'placeholder',
					html: `<div>${histogram.err}</div>`
				}]
			};
			bulk.push(group);
			this.errors.set(key, group);
			this.pending.delete(key);
			return;
		},
		createPendingFacet(bulk, histogram) {
			const key = this.getHistogramKey(histogram);
			// check if already added as placeholder
			if (this.pending.has(key)) {
				return;
			}
			// add placeholder
			const group = {
				label: histogram.name,
				key: key,
				facets: [
					{
						placeholder: true,
						key: 'placeholder',
						html: `
							<div class="bounce1"></div>
							<div class="bounce2"></div>
							<div class="bounce3"></div>`
					}
				]
			};
			bulk.push(group);
			this.pending.set(key, group);
		},
		createFacet(bulk, histogram) {
			const key = this.getHistogramKey(histogram);

			let group;
			const filter = this.$store.getters.getRouteFilter(histogram.name);

			switch (histogram.type) {
				case 'categorical':
					group = {
						label: histogram.name,
						key: key,
						facets: histogram.buckets.map(b => {
							return {
								value: b.key,
								count: b.count
							};
						})
					};
					break;

				case 'numerical':
					const selection = {};
					if (getFilterType(filter) === NUMERICAL_FILTER) {
						const decoded = decodeFilter(filter);
						selection.range = {
							from: decoded.min,
							to: decoded.max,
						};
					}
					group = {
						label: histogram.name,
						key: key,
						facets: [
							{
								selection: selection,
								histogram: {
									slices: histogram.buckets.map(b => {
										return {
											label: b.key,
											count: b.count
										};
									})
								}
							}
						]
					};
					break;

				default:
					console.warn('unrecognized histogram type', histogram.type);
					return;
			}

			// collapse if disabled
			group.collapsed = isDisabled(filter);

			// append
			this.facets.replaceGroup(group);
			// track
			this.groups.set(key, group);
			this.pending.delete(key);
			this.errors.delete(key);
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
