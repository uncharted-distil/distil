<template>
	<div class="available-training-variables">
		<h6 class="nav-link">Available features</h6>
		<variable-facets
			enable-search
			enable-facet-filtering
			type-change
			instance-name="availableVars"
			:variables="variables"
			:dataset="dataset"
			:html="html">
		</variable-facets>
	</div>
</template>

<script lang="ts">

import { overlayRouteEntry } from '../util/routes';
import { getters as dataGetters } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { VariableSummary } from '../store/data/index';
import VariableFacets from '../components/VariableFacets.vue';
import 'font-awesome/css/font-awesome.css';
import Vue from 'vue';

export default Vue.extend({
	name: 'available-training-variables',

	components: {
		VariableFacets
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		variables(): VariableSummary[] {
			return dataGetters.getAvailableVariableSummaries(this.$store);
		},
		html(): ( { key: string } ) => HTMLDivElement {
			return (group: { key: string }) => {
				const container = document.createElement('div');
				const trainingElem = document.createElement('button');
				trainingElem.className += 'btn btn-sm btn-outline-success mr-2 mb-2';
				trainingElem.innerHTML = 'Add to Training Set';
				trainingElem.addEventListener('click', () => {
					const training = routeGetters.getRouteTrainingVariables(this.$store);
					const trainingArray = training ? training.split(',') : [];
					const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
						training: trainingArray.concat([ group.key ]).join(',')
					});
					this.$router.push(entry);
				});
				const targetElem = document.createElement('button');
				targetElem.className += 'btn btn-sm btn-outline-success mr-2 mb-2';
				targetElem.innerHTML = 'Set as Target';
				targetElem.addEventListener('click', () => {
					const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
						target: group.key,
					});
					this.$router.push(entry);
				});
				container.appendChild(trainingElem);
				container.appendChild(targetElem);
				return container;
			};
		}
	}
});
</script>

<style>
.available-training-variables {
	display: flex;
	flex-direction: column;
}
</style>
