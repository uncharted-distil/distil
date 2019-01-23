<template>
	<div class='search-bar'>
		<b-form-input
			ref="searchbox"
			v-model="terms"
			type="text"
			placeholder="Search datasets"
			name="datasetsearch"
			@change="submitSearch"></b-form-input>
		<i class="fa fa-search search-icon" @click="submitSearch"></i>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import { createRouteEntry, overlayRouteEntry } from '../util/routes';
import { getters as routeGetters } from '../store/route/module';
import { SEARCH_ROUTE } from '../store/route/index';
import Vue from 'vue';

export default Vue.extend({
	name: 'search-bar',

	data() {
		return {
			uncommittedInput: false,
			uncommittedTerms: ''
		};
	},

	computed: {
		terms: {
			set(terms: string) {
				this.uncommittedTerms = terms;
				this.uncommittedInput = true;
			},
			get(): string {
				if (this.uncommittedInput) {
					return this.uncommittedTerms;
				}
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
	},

	methods: {
		submitSearch(arg) {
			const path = routeGetters.getRoutePath(this.$store);
			if (path !== SEARCH_ROUTE) {
				const routeEntry = createRouteEntry(SEARCH_ROUTE, {
					terms: this.uncommittedTerms
				});
				this.$router.push(routeEntry);
			} else {
				const routeEntry = overlayRouteEntry(this.$route, {
					terms: this.uncommittedTerms
				});
				this.$router.push(routeEntry);
			}
			this.uncommittedInput = false;
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
	cursor: pointer;
}
</style>
