<template>
	<div class="available-training-variables">
		<p class="nav-link font-weight-bold">Available Features
			<i class="float-right fa fa-angle-right fa-lg"></i>
		</p>
		<variable-facets
			ref="facets"
			enable-search
			enable-type-change
			:instance-name="instanceName"
			:rows-per-page="numRowsPerPage"
			:groups="groups"
			:html="html">
			<div class="available-variables-menu">
				<div>
					{{subtitle}}
				</div>
				<div v-if="groups.length > 0">
					<b-button size="sm" variant="outline-secondary" @click="addAll">Add All</b-button>
				</div>
			</div>
		</variable-facets>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import { overlayRouteEntry } from '../util/routes';
import { Variable, VariableSummary } from '../store/dataset/index';
import { getters as datasetGetters } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';
import { filterSummariesByDataset, NUM_PER_PAGE } from '../util/data';
import { AVAILABLE_TRAINING_VARS_INSTANCE } from '../store/route/index';
import { Group, createGroups, updateImportance } from '../util/facets';
import VariableFacets from '../components/VariableFacets';
import { Dictionary } from 'vue-router/types/router';

export default Vue.extend({
	name: 'available-training-variables',

	components: {
		VariableFacets
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		availableVariableSummaries(): VariableSummary[] {
			return routeGetters.getAvailableVariableSummaries(this.$store);
		},
		variables(): Variable[] {
			return datasetGetters.getVariables(this.$store);
		},
		groups(): Group[] {
			const filtered = filterSummariesByDataset(this.availableVariableSummaries, this.dataset);
			const groups = createGroups(filtered);
			return updateImportance(groups, this.variables);
		},
		subtitle(): string {
			return `${this.groups.length} features available`;
		},
		numRowsPerPage(): number {
			return NUM_PER_PAGE;
		},
		instanceName(): string {
			return AVAILABLE_TRAINING_VARS_INSTANCE;
		},
		html(): (group: Group) => HTMLDivElement {
			return (group: Group) => {
				const container = document.createElement('div');
				const trainingElem = document.createElement('button');
				trainingElem.className += 'btn btn-sm btn-outline-secondary ml-2 mr-1 mb-2';
				trainingElem.innerHTML = 'Add';
				trainingElem.addEventListener('click', () => {
					const training = routeGetters.getRouteTrainingVariables(this.$store);
					const trainingArray = training ? training.split(',') : [];
					const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
						training: trainingArray.concat([ group.colName ]).join(',')
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
				console.log(variable);
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
.available-variables-menu {
	display: flex;
	justify-content: space-between;
	padding: 4px 0;
	line-height: 30px;
}
</style>
