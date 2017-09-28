<template>
	<div class='training-variables'>
		<div class="bg-faded rounded-top">
			<h6 class="nav-link">Training Set</h6>
		</div>
		<variable-facets
			enable-search
			enable-sort
			enable-facet-filtering
			instance-name="trainingVars"
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
	name: 'training-variables',

	components: {
		VariableFacets
	},

	computed: {
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		variables() {
			return this.$store.getters.getTrainingVariableSummaries();
		},
		html() {
			return (group) => {
				const container = document.createElement('div');
				const remove = document.createElement('button');
				remove.className += 'btn btn-sm btn-outline-secondary mb-2';
				remove.innerHTML = 'Remove';
				remove.addEventListener('click', () => {
					const training = this.$store.getters.getTrainingVariables();
					training.splice(training.indexOf(group.key), 1);
					const entry = createRouteEntryFromRoute(this.$store.getters.getRoute(), {
						training: training.join(',')
					});
					this.$router.push(entry);
				});
				container.appendChild(remove);
				return container;
			};
		}
	}
};
</script>

<style>
.training-variables {
	display: flex;
	flex-direction: column;
}
</style>
