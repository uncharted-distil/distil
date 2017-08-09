<template>
	<div class='search-bar'>
		<b-form-input
			v-model="terms"
			type="text"
			placeholder="Search datasets"
			name="datasetsearch"></b-form-input>
	</div>
</template>

<script>
import _ from 'lodash';
import {createRouteEntry} from '../util/routes';

export default {
	name: 'search-bar',

	computed: {
		terms: {
			set: _.throttle(function(terms) {
				const routeEntry = createRouteEntry('/search', null, terms, this.$store.getters.getRouteFilters());
				this.$router.push(routeEntry);
			}, 500),
			get: function() {
				return this.$store.getters.getRouteTerms();
			}
		}
	},

	mounted() {
		this.$store.dispatch('searchDatasets', this.terms);
	},

	watch: {
		'$route.query.terms'() {
			this.$store.dispatch('searchDatasets', this.terms);
		}
	}
};
</script>

<style>
</style>
