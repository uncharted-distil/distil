<template>
	<div class="available-training-variables">
		<p class="nav-link font-weight-bold">Available Features
			<i class="float-right fa fa-angle-right fa-lg"></i>
		</p>
		<variable-facets
			ref="facets"
			enable-search
			enable-type-change
			instance-name="availableTrainingVars"
			:groups="groups"
			:dataset="dataset"
			:html="html">
			<div v-if="groups.length > 0" class="pb-2">
				<b-button size="sm" variant="outline-secondary" @click="addAll">Add All</b-button>
			</div>
			<div>
				{{subtitle}}
			</div>
		</variable-facets>
	</div>
</template>

<script lang="ts">

import { overlayRouteEntry } from '../util/routes';
import { getters as dataGetters } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { filterSummariesByDataset } from '../util/data';
import { Group, createGroups } from '../util/facets';
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
		groups(): Group[] {
			const summaries = dataGetters.getAvailableVariableSummaries(this.$store);
			const filtered = filterSummariesByDataset(summaries, this.dataset);
			return createGroups(filtered);
		},
		subtitle(): string {
			return `${this.groups.length} features available (sorted by interestingness)`;
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
	},

	methods: {
		addAll() {
			const facets = this.$refs.facets as any;
			const training = routeGetters.getRouteTrainingVariables(this.$store);
			const trainingArray = training ? training.split(',') : [];
			facets.availableVariables().forEach(variable => {
				trainingArray.push(variable);
			});
			const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
				training: trainingArray.join(',')
			});
			this.$router.push(entry);
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
