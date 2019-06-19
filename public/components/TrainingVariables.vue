<template>
	<div class="training-variables" v-bind:class='{"included": includedActive, "excluded": !includedActive }'>
		<p class="nav-link font-weight-bold">Features to Model
			<i class="float-right fa fa-angle-right fa-lg"></i>
		</p>
		<variable-facets
			ref="facets"
			enable-search
			enable-highlighting
			enable-type-change
			:instance-name="instanceName"
			:rows-per-page="numRowsPerPage"
			:summaries="trainingVariableSummaries"
			:html="html">
			<div class="available-variables-menu">
				<div>
					{{subtitle}}
				</div>
				<div v-if="trainingVariableSummaries.length > 0">
					<b-button size="sm" variant="outline-secondary" @click="removeAll">Remove All</b-button>
				</div>
			</div>
			<div v-if="trainingVariableSummaries.length === 0">
				<i class="no-selections-icon fa fa-arrow-circle-left"></i>
			</div>
		</variable-facets>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import VariableFacets from '../components/VariableFacets';
import { Variable, VariableSummary, Highlight } from '../store/dataset/index';
import { getters as routeGetters } from '../store/route/module';
import { getters as datasetGetters } from '../store/dataset/module';
import { TRAINING_VARS_INSTANCE } from '../store/route/index';
import { Group } from '../util/facets';
import { NUM_PER_PAGE } from '../util/data';
import { overlayRouteEntry } from '../util/routes';
import { removeFiltersByName } from '../util/filters';

export default Vue.extend({
	name: 'training-variables',

	components: {
		VariableFacets
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		numRowsPerPage(): number {
			return NUM_PER_PAGE;
		},
		includedActive(): boolean {
			return routeGetters.getRouteInclude(this.$store);
		},
		highlight(): Highlight {
			return routeGetters.getDecodedHighlight(this.$store);
		},
		trainingVariableSummaries(): VariableSummary[] {
			return routeGetters.getTrainingVariableSummaries(this.$store);
		},
		variables(): Variable[] {
			return datasetGetters.getVariables(this.$store);
		},
		subtitle(): string {
			return `${this.trainingVariableSummaries.length} features selected`;
		},
		instanceName(): string {
			return TRAINING_VARS_INSTANCE;
		},
		html(): (Group) => HTMLDivElement {
			return (group: Group) => {
				const container = document.createElement('div');
				const remove = document.createElement('button');
				remove.className += 'btn btn-sm btn-outline-secondary ml-2 mr-1 mb-2';
				remove.innerHTML = 'Remove';
				remove.addEventListener('click', () => {
					const training = routeGetters.getDecodedTrainingVariableNames(this.$store);
					training.splice(training.indexOf(group.colName), 1);
					const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
						training: training.join(',')
					});
					this.$router.push(entry);
					removeFiltersByName(this.$router, group.colName);
				});
				container.appendChild(remove);
				return container;
			};
		}
	},

	methods: {
		removeAll() {
			const facets = this.$refs.facets as any;
			const training = routeGetters.getDecodedTrainingVariableNames(this.$store);
			facets.availableVariables().forEach(variable => {
				training.splice(training.indexOf(variable), 1);
			});
			const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
				training: training.join(',')
			});
			this.$router.push(entry);
		}
	}
});
</script>

<style>
.training-variables {
	display: flex;
	flex-direction: column;
}
.no-selections-icon {
	color: #32CD32;
	font-size: 46px;
}
.training-variables-menu {
	display: flex;
	justify-content: space-between;
	padding: 4px 0;
	line-height: 30px;
}
</style>
