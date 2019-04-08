<template>
	<div class="available-target-variables">
		<variable-facets
			enable-search
			enable-type-change
			enable-title
			ignore-highlights
			:instance-name="instanceName"
			:rows-per-page="numRowsPerPage"
			:groups="groups"
			:html="html">
		</variable-facets>
	</div>
</template>

<script lang="ts">

import 'jquery';
import { getters as datasetGetters, actions as datasetActions } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';
import { createRouteEntry } from '../util/routes';
import { filterSummariesByDataset } from '../util/data';
import VariableFacets from '../components/VariableFacets';
import { Grouping, Variable } from '../store/dataset/index';
import { AVAILABLE_TARGET_VARS_INSTANCE, SELECT_TRAINING_ROUTE } from '../store/route/index';
import { Group, createGroups } from '../util/facets';
import Vue from 'vue';

// 9 so it makes a nice clean grid
const NUM_TARGET_PER_PAGE = 9;

export default Vue.extend({
	name: 'available-target-variables',

	components: {
		VariableFacets
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		groups(): Group[] {
			const summaries = datasetGetters.getVariableSummaries(this.$store);
			const filtered = filterSummariesByDataset(summaries, this.dataset);
			return createGroups(filtered);
		},
		numRowsPerPage(): number {
			return NUM_TARGET_PER_PAGE;
		},
		instanceName(): string {
			return AVAILABLE_TARGET_VARS_INSTANCE;
		},
		variables(): Variable[] {
			return datasetGetters.getVariables(this.$store);
		},
		html(): (group: { key: string }) => HTMLDivElement {
			return (group: { key: string }) => {
				const container = document.createElement('div');
				const targetElem = document.createElement('button');
				targetElem.className += 'btn btn-sm btn-success ml-2 mr-2 mb-2';
				targetElem.innerHTML = 'Select Target';
				targetElem.addEventListener('click', () => {
					const target = group.key;
					// remove from training
					const trainingStr = routeGetters.getRouteTrainingVariables(this.$store);
					const training = trainingStr ? trainingStr.split(',') : [];
					const index = training.indexOf(target);
					if (index !== -1) {
						training.splice(index, 1);
					}
					const entry = createRouteEntry(SELECT_TRAINING_ROUTE, {
						target: group.key,
						dataset: routeGetters.getRouteDataset(this.$store),
						filters: routeGetters.getRouteFilters(this.$store),
						timeseriesAnalysis: routeGetters.getRouteTimeseriesAnalysis(this.$store),
						training: training.join(',')
					});
					this.$router.push(entry);
				});
				container.appendChild(targetElem);

				const v = this.variables.find(v => {
					return v.colName === group.key;
				});
				if (v.grouping) {
					const groupingElem = document.createElement('button');
					groupingElem.className += 'btn btn-sm btn-primary ml-2 mr-2 mb-2 float-right';
					groupingElem.innerHTML = 'Remove Grouping';
					groupingElem.addEventListener('click', () => {
						datasetActions.removeGrouping(this.$store, {
							dataset: this.dataset,
							grouping: v.grouping
						});
					});
					container.appendChild(groupingElem);
				}

				return container;
			};
		}
	}

});
</script>

<style>
.available-target-variables {
	height: 100%;
}

.available-target-variables .facets-group .facets-facet-horizontal .facet-range {
	cursor: pointer !important;
}

.available-target-variables .facets-group {
	margin: 5px;
}
.available-target-variables .facet-filters {
	padding: 2rem;
}
.available-target-variables .facets-root-container {
	display: flex;
	flex-wrap: wrap;
	justify-content: center;
}

.available-target-variables .facets-group-container {
	flex-grow: 1;
	width: 30%;
	max-width: 30%;
	margin: 5px;
}
</style>
