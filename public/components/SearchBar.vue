<template>
	<div class='search-bar'>
		<b-form-input
			ref="searchbox"
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
				const routeEntry = createRouteEntry('/search', {
					terms: terms,
				});
				this.$router.push(routeEntry);
				const component = this;
				// maintain focus on routing
				setTimeout(() => {
					console.log('sdf', component);
					component.$refs.searchbox.focus();
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
