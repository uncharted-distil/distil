<template>
	<div class="available-variables">
		<h6 class="nav-link">Features</h6>
		<variable-facets
			enable-search
			enable-sort
			enable-facet-filtering
			instance-name="availableVars"
			:variables="variables"
			:dataset="dataset"
			:html="html">
		</variable-facets>
	</div>
</template>

<script>

import { createRouteEntryFromRoute } from '../util/routes';
import VariableFacets from '../components/VariableFacets';
import 'font-awesome/css/font-awesome.css';

export default {
	name: 'available-variables',

	components: {
		VariableFacets
	},

	computed: {
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		variables() {
			return this.$store.getters.getAvailableVariableSummaries();
		},
		html() {
			return (group) => {
				const container = document.createElement('div');
				const trainingElem = document.createElement('button');
				trainingElem.className += 'btn btn-sm btn-outline-success mr-2 mb-2';
				trainingElem.innerHTML = 'Add to Training Set';
				trainingElem.addEventListener('click', () => {
					const training = this.$store.getters.getTrainingVariables();
					const entry = createRouteEntryFromRoute(this.$store.getters.getRoute(), {
						training: training.concat([group.key]).join(',')
					});
					this.$router.push(entry);
				});
				const targetElem = document.createElement('button');
				targetElem.className += 'btn btn-sm btn-outline-success mr-2 mb-2';
				targetElem.innerHTML = 'Set as Target';
				targetElem.addEventListener('click', () => {
					const entry = createRouteEntryFromRoute(this.$store.getters.getRoute(), {
						target: group.key
					});
					this.$router.push(entry);
				});
				container.appendChild(trainingElem);
				container.appendChild(targetElem);
				return container;
			};
		}
	}
};
</script>

<style>
.available-variables {
	display: flex;
	flex-direction: column;
}
</style>
