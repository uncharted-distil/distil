<template>
	<div class="filter-badge" v-bind:class="{ active: activeFilter }">
		{{filterName}}
		<span v-if="filter.type===NUMERICAL_FILTER">
			{{filter.min.toFixed(2)}} : {{filter.max.toFixed(2)}}
		</span>
		<span v-if="filter.type===BIVARIATE_FILTER">
			[{{filter.minX.toFixed(2)}}, {{filter.minY.toFixed(2)}}] to [{{filter.maxX.toFixed(2)}}, {{filter.maxY.toFixed(2)}}]
		</span>
		<span v-if="filter.type===CATEGORICAL_FILTER || filter.type===FEATURE_FILTER || filter.type===CLUSTER_FILTER">
			{{filter.categories.join(',')}}
		</span>

		<b-button class="remove-button" size="sm" @click="onClick">
			<i class="fa fa-times"></i>
		</b-button>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import { removeFilterFromRoute, Filter, NUMERICAL_FILTER, BIVARIATE_FILTER, CATEGORICAL_FILTER, FEATURE_FILTER, CLUSTER_FILTER } from '../util/filters';
import { clearHighlight } from '../util/highlights';
import { getVarType, isFeatureType, removeFeaturePrefix, isClusterType, removeClusterPrefix } from '../util/types';

export default Vue.extend({
	name: 'filter-badge',

	props: {
		filter: Object as () => Filter,
		activeFilter: Boolean as () => boolean
	},

	computed: {
		filterName(): string {
			const type = getVarType(this.filter.key);
			if (isFeatureType(type)) {
				return removeFeaturePrefix(this.filter.key);
			}
			if (isClusterType(type)) {
				return removeClusterPrefix(this.filter.key);
			}
			return this.filter.key;
		},
		NUMERICAL_FILTER(): string {
			return NUMERICAL_FILTER;
		},
		CATEGORICAL_FILTER(): string {
			return CATEGORICAL_FILTER;
		},
		FEATURE_FILTER(): string {
			return FEATURE_FILTER;
		},
		CLUSTER_FILTER(): string {
			return CLUSTER_FILTER;
		},
		BIVARIATE_FILTER(): string {
			return BIVARIATE_FILTER;
		}
	},

	methods: {
		onClick() {
			if (!this.activeFilter) {
				removeFilterFromRoute(this.$router, this.filter);
			} else {
				clearHighlight(this.$router);
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
	margin: 2px 4px 2px 0;
	border-radius: 4px;
	background-color: #333;
}

.filter-badge.active {
	background-color: #00c6e1;
}

button.remove-button {
	color: #fff;
	margin-left: 8px;
	background: none;
	border-radius: 0px;
	border-top-right-radius: 4px;
	border-bottom-right-radius: 4px;
	border: none;
	border-left: 1px solid #fff;
}
button.remove-button:hover {
	color: #fff;
	background-color: #0089a4;
	border: none;
	border-left: 1px solid #fff;
}

.active button.remove-button:hover {
	background-color: #0089a4;
}
</style>
