<template>
	<div class="filter-badge" v-bind:class="{ active: activeFilter }">
		{{filterName}}
		<span v-if="filter.type==='numerical'">
			{{filter.min}} : {{filter.max}}
		</span>
		<span v-if="filter.type==='categorical'">
			{{filter.categories.join(',')}}
		</span>

		<b-button class="remove-button" size="sm" @click="onClick">
			<i class="fa fa-times"></i>
		</b-button>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import { removeFilterFromRoute } from '../util/filters';
import { clearHighlightRoot } from '../util/highlights';
import { removeMetaPrefix } from '../util/types';

export default Vue.extend({
	name: 'filter-badge',

	props: {
		filter: Object,
		activeFilter: Boolean
	},

	computed: {
		filterName(): string {
			return removeMetaPrefix(this.filter.name);
		}
	},

	methods: {
		onClick() {
			if (!this.activeFilter) {
				removeFilterFromRoute(this, this.filter);
			} else {
				clearHighlightRoot(this);
			}
		}
	}
});
</script>

<style>
.filter-badge {
	position: relative;
	height: 28px;
	display: inline-block;
	color: #fff;
	padding-left: 8px;
	margin: 2px 4px;
	border-radius: 4px;
	background-color: #00c6e1;
}

.filter-badge.active {
	background-color: #00c6e1;
}

.remove-button {
	color: #fff;
	margin-left: 8px;
	background: none;
	border-radius: 0px;
	border-top-right-radius: 4px;
	border-bottom-right-radius: 4px;
	border: none;
	border-left: 1px solid #fff;
}
.remove-button:hover {
	color: #fff;
	background-color: #0089a4;
	border: none;
	border-left: 1px solid #fff;
}

.active .remove-button:hover {
	background-color: #0089a4;
}
</style>
