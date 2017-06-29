<template>
	<div class='variable-summaries'>
		<div class="nav bg-faded rounded-top">
			<h6 class="nav-link">Summaries</h6>
		</div>
		<div id="variable-facets"></div>
	</div>
</template>

<script>

import _ from 'lodash';

import Facets from '@uncharted.software/stories-facets';
import { encodeFilters, decodeFilter, decodeFilters, isEmpty } from '../util/filters';
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
		dataset: function() {
			return this.$store.getters.getRouteDataset();
		}
	},

	methods: {
		updateFilterRoute: function(key, values) {
			const filters = this.$store.getters.getRouteFilters();
			const decoded = decodeFilters(filters);
			let filter = decoded[key];
			if (!filter) {
				filter = {
					name: key,
					enabled: true
				};
				decoded[key] = filter;
			}
			_.forIn(values, (v, k) => {
				filter[k] = v;
			});
			const encoded = encodeFilters(decoded);
			const query = _.merge({
					dataset: this.$store.getters.getRouteDataset(),
					terms: this.$store.getters.getRouteTerms(),
				}, encoded);
			// remove filter if it is empty
			if (isEmpty(filter)) {
				query[key] = undefined;
			}
			this.$router.push({
				path: '/dataset',
				query: query
			});
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
				type: 'numerical',
				enabled: true,
				min: parseFloat(value.from.label[0]),
				max: parseFloat(value.to.label[0])
			});
		});

		// update it's contents when the dataset changes
		this.$store.watch(() => this.$store.state.variableSummaries, histograms => {

			const bulk = [];
			// for each histogram
			histograms.forEach(histogram => {

				const key = histogram.name;

				if (histogram.err) {
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
					this.facets.replaceGroup(group);
					this.errors.set(key, group);
					this.pending.delete(key);
					return;
				}

				if (histogram.pending) {
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
					return;
				}

				// check if already added
				if (this.groups.has(key)) {
					return;
				}

				let group;
				const filter = this.$store.getters.getRouteFilter(histogram.name);
				const decoded = decodeFilter(filter);
				const collapsed = decoded && !decoded.enabled;

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
						if (decoded && _.has(decoded, 'min') && _.has(decoded, 'max')) {
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

				group.collapsed = collapsed;

				// append
				this.facets.replaceGroup(group);
				// track
				this.groups.set(key, group);
				this.pending.delete(key);
				this.errors.delete(key);
			});

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
			this.$store.commit('setFilterState', {});
			this.$store.dispatch('getVariableSummaries', this.dataset);
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
