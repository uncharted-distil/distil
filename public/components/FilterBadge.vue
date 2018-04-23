<template>
	<div class="filter-badge">
		{{filter.name}}
		<span v-if="filter.type==='numerical'">
			{{filter.min}} : {{filter.max}}
		</span>
		<span v-if="filter.type==='categorical'">
			{{filter.categories.join(',')}}
		</span>

		<b-button v-if="!noRemove" size="sm" @click="onClick">
			<i class="fa fa-times"></i>
		</b-button>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import { removeFilterFromRoute } from '../util/filters';

export default Vue.extend({
	name: 'filter-badge',

	props: {
		filter: Object,
		noRemove: Boolean
	},

	methods: {
		onClick() {
			removeFilterFromRoute(this, this.filter);
		}
	}
});
</script>

<style>
.filter-badge {
	padding: 2px 4px;
	background-color: #00c6e1;
}
</style>
