<template>
	<div class="available-target-variables">
		<h6 class="nav-link">Available features</h6>
		<variable-facets
			enable-search
			enable-facet-filtering
			type-change
			instance-name="availableVars"
			:variables="variables"
			:dataset="dataset"
			v-on:click="onClick">
		</variable-facets>
	</div>
</template>

<script lang="ts">

import { createRouteEntryFromRoute } from '../util/routes';
import 'jquery';
import { getters as dataGetters } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { VariableSummary } from '../store/data/index';
import VariableFacets from '../components/VariableFacets.vue';
import 'font-awesome/css/font-awesome.css';
import Vue from 'vue';

export default Vue.extend({
	name: 'available-target-variables',

	components: {
		VariableFacets
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		variables(): VariableSummary[] {
			return dataGetters.getAvailableVariableSummaries(this.$store);
		}
	},

	methods: {
		onClick(key: string) {
			const entry = createRouteEntryFromRoute(routeGetters.getRoute(this.$store), {
				target: key,
			});
			this.$router.push(entry);
		}
	}
});
</script>

<style>
.available-target-variables {
	display: flex;
	flex-direction: column;
}
.available-target-variables .variable-facets-container {
	overflow: visible;
}
.available-target-variables .facets-group,
.available-target-variables .facets-group .group-header,
.available-target-variables .facets-group .group-facet-container,
.available-target-variables .facets-group .facets-facet-horizontal {
	cursor: pointer !important;
}

.available-target-variables .facets-group {
	z-index: 0;
	margin: 5px;
}

.available-target-variables .facets-group:hover {
	border-style: solid;
	border-color: #03c6e1;
	box-shadow: 0 0 10px #03c6e1;
	border-width: 1px;
	border-radius: 2px;
	z-index: 1;
}

</style>
