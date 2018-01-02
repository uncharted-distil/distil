<template>
	<div class="training-variables">
		<h6 class="nav-link">Training features</h6>
		<variable-facets
			enable-search
			enable-facet-filtering
			type-change
			instance-name="trainingVars"
			:variables="variables"
			:dataset="dataset"
			:html="html">
		</variable-facets>
	</div>
</template>

<script lang="ts">

import { createRouteEntryFromRoute } from '../util/routes';
import VariableFacets from '../components/VariableFacets';
import 'font-awesome/css/font-awesome.css';
import Vue from 'vue';
import { getters as dataGetters} from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { VariableSummary } from '../store/data/index';
import { Group } from '../util/facets';

export default Vue.extend({
	name: 'training-variables',

	components: {
		VariableFacets
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		variables(): VariableSummary[] {
			return dataGetters.getTrainingVariableSummaries(this.$store);
		},
		html(): (Group) => HTMLDivElement {
			return (group: Group) => {
				const container = document.createElement('div');
				const remove = document.createElement('button');
				remove.className += 'btn btn-sm btn-outline-danger mb-2';
				remove.innerHTML = 'Remove';
				remove.addEventListener('click', () => {
					const training = routeGetters.getRouteTrainingVariables(this.$store).split(',');
					training.splice(training.indexOf(group.key), 1);
					const entry = createRouteEntryFromRoute(routeGetters.getRoute(this.$store), {
						training: training.join(',')
					});
					this.$router.push(entry);
				});
				container.appendChild(remove);
				return container;
			};
		}
	}
});
</script>

<style>
.training-variables {
	display: flex;
	flex-direction: column;
}
</style>
