<template>
	<div class="available-training-variables">
		<p class="nav-link font-weight-bold">Available Features</p>
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

import { overlayRouteEntry } from '../util/routes';
import { getters as dataGetters } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { VariableSummary } from '../store/data/index';
import { filterSummariesByDataset } from '../util/data';
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
			const summaries = dataGetters.getVariableSummaries(this.$store);
			return filterSummariesByDataset(summaries, this.dataset);
		},
		html(): ( { key: string } ) => HTMLDivElement {
			return (group: { key: string }) => {
				const container = document.createElement('div');
				const trainingElem = document.createElement('button');
				trainingElem.className += 'btn btn-sm btn-outline-secondary ml-2 mr-2 mb-2';
				trainingElem.innerHTML = 'Add';
				trainingElem.addEventListener('click', () => {
					const training = routeGetters.getRouteTrainingVariables(this.$store);
					const trainingArray = training ? training.split(',') : [];
					const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
						training: trainingArray.concat([ group.key ]).join(',')
					});
					this.$router.push(entry);
				});
				container.appendChild(trainingElem);
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
