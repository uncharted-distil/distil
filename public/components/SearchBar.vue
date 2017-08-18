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

export default {
	name: 'search-bar',

	computed: {
		terms: {
			set: _.throttle(function(terms) {
				this.$router.push({
					path: '/search',
					query: {
						terms: terms
					}
				});
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
