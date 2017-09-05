<template>
	<div class='result-summaries'>
		<div class="nav bg-faded rounded-top">
			<h6 class="nav-link">Results</h6>
		</div>
		<div v-if="groups.length===0">
			No results
		</div>
		<facets v-if="groups.length>0"
			:groups="groups"
			root="result-summaries"></facets>
	</div>
</template>

<script>

import Facets from '../components/Facets';
import 'font-awesome/css/font-awesome.css';
import '../styles/spinner.css';
import _ from 'lodash';

const SPINNER_HTML = [
	'<div class="bounce1"></div>',
	'<div class="bounce2"></div>',
	'<div class="bounce3"></div>'].join('');

export default {

	name: 'result-summaries',

	components: {
		Facets
	},

	mounted() {
		// kick off a result fetch when the component is first displayed
		this.$store.dispatch('getResultsSummaries', {
			dataset: this.$store.getters.getRouteDataset(),
			resultsUri: this.$store.getters.getRouteResultsUri()
		});
	},

	computed: {
		groups() {
			// get the selected result summary and create a facet from it
			const results = this.$store.state.resultsSummaries;
			return this.createGroups(results);
		}
	},

	watch: {
		// watch the route and update the results if its modified
		'$route.query.results'() {
			this.$store.dispatch('getResultsSummaries', {
				dataset: this.$store.getters.getRouteDataset(),
				resultsUri: this.$store.getters.getRouteResultsUri()
			});
		}
	},

	methods: {
		// creates facet groups based on state of summary data
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

		// creates the summary facet when all required data is available
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

		// Creates a facet to display an error status message
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

		// Creates a facet to display an pending spinner while full facet
		// data is fetched
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
