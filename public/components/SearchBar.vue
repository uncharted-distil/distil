<template>
	<div class='search-bar'>
		<b-form-input
			ref="searchbox"
			v-model="terms"
			type="text"
			placeholder="Search datasets"
			name="datasetsearch"></b-form-input>
		<i class="fa fa-search search-icon"></i>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import { createRouteEntry } from '../util/routes';
import { actions } from '../store/data/module';
import { getters } from '../store/route/module';
import Vue from 'vue';

export default Vue.extend({
	name: 'search-bar',

	computed: {
		terms: {
			set: _.throttle(function(terms) {
				const path = !_.isEmpty(terms) ? '/search' : getters.getRoutePath(this.$store);
				const routeEntry = createRouteEntry(path, {
					terms: terms
				});
				this.$router.push(routeEntry);
			}, 500),
			get: function() {
				return getters.getRouteTerms(this.$store);
			}
		}
	},

	mounted() {
		actions.searchDatasets(this.$store, this.terms);
		if (!_.isEmpty(this.terms)) {
			this.$refs.searchbox.focus();
		}
	},

	watch: {
		'$route.query.terms'() {
			actions.searchDatasets(this.$store, this.terms);
		}
	}
});
</script>

<style>
.search-bar {
	position: relative;
}
.search-icon {
	position: absolute;
	padding: 0.5rem 0.75rem;
	font-size: 1rem;
	line-height: 1.25;
	top: 0;
	right: 0;

}
</style>
