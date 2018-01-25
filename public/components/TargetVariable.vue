<template>
	<div>
		<p class="nav-link font-weight-bold">Target Feature</p>
		<div class="target-no-target" v-if="variables.length===0">
			<div class="text-danger">
				<i class="fa fa-times missing-icon"></i><strong>No Target Feature Selected</strong>
			</div>
		</div>
		<variable-facets v-if="variables.length>0"
			type-change
			:variables="variables"
			:dataset="dataset"
			:html="html"></variable-facets>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import 'font-awesome/css/font-awesome.css';
import VariableFacets from '../components/VariableFacets';
import { createRouteEntry } from '../util/routes';
import { getters as dataGetters } from '../store/data/module';
import { getters as routeGetters} from '../store/route/module';
import { SELECT_ROUTE } from '../store/route/index';
import { VariableSummary } from '../store/data/index';

export default Vue.extend({
	name: 'target-variables',

	components: {
		VariableFacets
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		variables(): VariableSummary[] {
			return dataGetters.getTargetVariableSummaries(this.$store);
		},
		html(): ( { key: string } ) => HTMLDivElement {
			return (group: { key: string }) => {
				const container = document.createElement('div');
				const targetElem = document.createElement('button');
				targetElem.className += 'btn btn-sm btn-outline-danger ml-2 mr-2 mb-2';
				targetElem.innerHTML = 'Select New Target Feature';
				targetElem.addEventListener('click', () => {
					const entry = createRouteEntry(SELECT_ROUTE, {
						target: group.key,
						dataset: routeGetters.getRouteDataset(this.$store),
						filters: routeGetters.getRouteFilters(this.$store),
						training: routeGetters.getRouteTrainingVariables(this.$store)
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
.target-no-target {
	width: 100%;
	background-color: #eee;
	padding: 8px;
	font-size: 1rem;
}
.missing-icon {
	padding-right: 4px;
}
</style>
