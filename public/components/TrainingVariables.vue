<template>
	<div class="training-variables">
		<p class="nav-link font-weight-bold">Features to Model</p>
		<variable-facets
			ref="facets"
			enable-search
			type-change
			instance-name="trainingVars"
			@click="onClick"
			:groups="groups"
			:dataset="dataset"
			:html="html">
			<div v-if="groups.length > 0" class="pb-2">
				<b-button size="sm" variant="outline-secondary" @click="removeAll">Remove All</b-button>
			</div>
			<div>
				{{subtitle}}
			</div>
			<div v-if="groups.length === 0">
				<i class="no-selections-icon fa fa-arrow-circle-left"></i>
			</div>
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
		},
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
.no-selections-icon {
	color: #32CD32;
	font-size: 46px;
}
</style>
