<template>
	<div class="training-variables">
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
			:groups="groups"
			:dataset="dataset"
			:html="html">
			<div class="available-variables-menu">
				<div>
					{{subtitle}}
				</div>
				<div v-if="groups.length > 0">
					<b-button size="sm" variant="outline-secondary" @click="removeAll">Remove All</b-button>
				</div>
			</div>
			<div v-if="groups.length === 0">
				<i class="no-selections-icon fa fa-arrow-circle-left"></i>
			</div>
		</variable-facets>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import VariableFacets from '../components/VariableFacets';
import { VariableSummary } from '../store/dataset/index';
import { Highlight } from '../store/highlights/index';
import { getters as routeGetters } from '../store/route/module';
import { Group, createGroups } from '../util/facets';
import { getHighlights } from '../util/highlights';
import { NUM_PER_PAGE } from '../util/data';
import { overlayRouteEntry } from '../util/routes';
import { removeFiltersByName } from '../util/filters';

import 'font-awesome/css/font-awesome.css';

export default Vue.extend({
	name: 'training-variables',

	components: {
		VariableFacets
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		instanceName(): string {
			return 'trainingVars';
		},
		numRowsPerPage(): number {
			return NUM_PER_PAGE;
		},
		highlights(): Highlight {
			return getHighlights(this.$store);
		},
		trainingVariableSummaries(): VariableSummary[] {
			return routeGetters.getTrainingVariableSummaries(this.$store);
		},
		groups(): Group[] {
		 	return createGroups(this.trainingVariableSummaries);
		},
		subtitle(): string {
			return `${this.groups.length} features selected`;
		},
		html(): (Group) => HTMLDivElement {
			return (group: Group) => {
				const container = document.createElement('div');
				const remove = document.createElement('button');
				remove.className += 'btn btn-sm btn-outline-secondary ml-2 mr-2 mb-2';
				remove.innerHTML = 'Remove';
				remove.addEventListener('click', () => {
					const training = routeGetters.getRouteTrainingVariables(this.$store).split(',');
					training.splice(training.indexOf(group.key), 1);
					const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
						training: training.join(',')
					});
					this.$router.push(entry);
					removeFiltersByName(this.$router, group.key);
				});
				container.appendChild(remove);
				return container;
			};
		}
	},

	methods: {
		removeAll() {
			const facets = this.$refs.facets as any;
			const training = routeGetters.getRouteTrainingVariables(this.$store);
			const trainingArray = training ? training.split(',') : [];
			facets.availableVariables().forEach(variable => {
				trainingArray.splice(trainingArray.indexOf(variable), 1);
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
