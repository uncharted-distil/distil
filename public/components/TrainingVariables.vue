<template>
	<div class="training-variables">
		<p class="nav-link font-weight-bold">Features to Model</p>
		<variable-facets
			enable-search
			type-change
			instance-name="trainingVars"
			@click="onClick"
			:groups="groups"
			:dataset="dataset"
			:html="html">
		</variable-facets>
	</div>
</template>

<script lang="ts">

import { overlayRouteEntry } from '../util/routes';
import VariableFacets from '../components/VariableFacets';
import 'font-awesome/css/font-awesome.css';
import Vue from 'vue';
import { getters as dataGetters} from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { Group, createGroups } from '../util/facets';
import { Highlights, getHighlights } from '../util/highlights';

export default Vue.extend({
	name: 'training-variables',

	components: {
		VariableFacets
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		highlightRoot(): Highlights {
			return getHighlights(this.$store);
		},
		groups(): Group[] {
			const summaries = dataGetters.getTrainingVariableSummaries(this.$store);
			const groups =  createGroups(summaries, false, false);
			if (this.highlightRoot.root) {
				groups.forEach(group => {
					if (group) {
						if (group.key === this.highlightRoot.root.key) {
							group.facets.forEach(facet => {
								facet.filterable = true;
							});
						}
					}
				});
			}
			return groups;
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
				});
				container.appendChild(remove);
				return container;
			};
		}
	},
	methods: {
		onClick(key: string) {
			console.log(key);
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
