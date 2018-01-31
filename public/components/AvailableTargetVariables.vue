<template>
	<div class="available-target-variables">
		<variable-facets
			enable-search
			type-change
			instance-name="availableVars"
			:variables="variables"
			:dataset="dataset"
			:html="html">
		</variable-facets>
	</div>
</template>

<script lang="ts">

import 'jquery';
import { getters as dataGetters } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { createRouteEntry } from '../util/routes';
import { VariableSummary } from '../store/data/index';
import VariableFacets from '../components/VariableFacets.vue';
import { CREATE_ROUTE } from '../store/route/index';
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
			return dataGetters.getVariableSummaries(this.$store);
		},
		html(): ( { key: string } ) => HTMLDivElement {
			return (group: { key: string }) => {
				const container = document.createElement('div');
				const targetElem = document.createElement('button');
				targetElem.className += 'btn btn-sm btn-outline-secondary ml-2 mr-2 mb-2';
				targetElem.innerHTML = 'Set as Target Feature';
				targetElem.addEventListener('click', () => {
					const target = group.key;
					// remove from training
					const trainingStr = routeGetters.getRouteTrainingVariables(this.$store);
					const training = trainingStr ? trainingStr.split(',') : [];
					const index = training.indexOf(target);
					if (index !== -1) {
						training.splice(index, 1);
					}
					const entry = createRouteEntry(CREATE_ROUTE, {
						target: group.key,
						dataset: routeGetters.getRouteDataset(this.$store),
						filters: routeGetters.getRouteFilters(this.$store),
						training: training.join(',')
					});
					this.$router.push(entry);
				});
				container.appendChild(targetElem);
				return container;
			};
		}
	}

});
</script>

<style>
.available-target-variables {
	display: flex;
	flex-direction: column;
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
.available-target-variables .facet-filters {
	padding: 2rem;
}
</style>
