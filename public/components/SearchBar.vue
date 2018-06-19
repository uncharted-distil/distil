<template>
	<div class='search-bar'>
		<b-form-input
			ref="searchbox"
			v-model="terms"
			debounce="500"
			type="text"
			placeholder="Search datasets"
			name="datasetsearch"></b-form-input>
		<i class="fa fa-search search-icon"></i>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import { createRouteEntry } from '../util/routes';
import { getters as routeGetters } from '../store/route/module';
import { SEARCH_ROUTE } from '../store/route/index';
import Vue from 'vue';

export default Vue.extend({
	name: 'search-bar',

	computed: {
		terms: {
			set(terms: string) {
				const path = !_.isEmpty(terms) ? SEARCH_ROUTE : routeGetters.getRoutePath(this.$store);
				const routeEntry = createRouteEntry(path, {
					terms: terms
				});
				this.$router.push(routeEntry);
			},
			get(): string {
				return routeGetters.getRouteTerms(this.$store);
			}
		}
	},

	mounted() {
		if (!_.isEmpty(this.terms)) {
			const component = this.$refs.searchbox as any;
			const elem = component.$el;
			// NOTE: hack to get the cursor at the end of the text after focus
			if (typeof elem.selectionStart === 'number') {
				elem.focus();
				elem.selectionStart = elem.selectionEnd = elem.value.length;
			} else if (typeof elem.createTextRange !== 'undefined') {
				elem.focus();
				const range = elem.createTextRange();
				range.collapse(false);
				range.select();
			}
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
