<template>
	<div class='available-variables'>
		<div class="bg-faded rounded-top">
			<h6 class="nav-link">Available Set</h6>
		</div>
		<variable-facets
			enable-filter="true"
			enable-toggle="true"
			instance-name="availableVars"
			:variables="variables"
			:dataset="dataset"
			:html="html">
		</variable-facets>
	</div>
</template>

<script>

import { createRouteEntry } from '../util/routes';
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
			return this.$store.getters.getAvailableVariables();
		},
		html() {
			return (group) => {
				const container = document.createElement('div');
				const training = document.createElement('button');
				training.className += 'btn btn-sm btn-outline-secondary mr-2 mb-2';
				training.innerHTML = 'Add to Training Set';
				training.addEventListener('click', () => {
					const path = this.$store.getters.getRoutePath();
					const entry = createRouteEntry(path, {
						dataset: this.$store.getters.getRouteDataset(),
						filters: this.$store.getters.getRouteFilters(),
						target: this.$store.getters.getRouteTargetVariable(),
						training: this.$store.getters.getRouteTrainingVariables().concat([ group.key ])
					});
					this.$router.push(entry);
					//this.$store.commit('addTrainingVariable', group.key);
				});
				const target = document.createElement('button');
				target.className += 'btn btn-sm btn-outline-secondary mr-2 mb-2';
				target.innerHTML = 'Set as Target';
				target.addEventListener('click', () => {
					const path = this.$store.getters.getRoutePath();
					const entry = createRouteEntry(path, {
						dataset: this.$store.getters.getRouteDataset(),
						filters: this.$store.getters.getRouteFilters(),
						training: this.$store.getters.getRouteTrainingVariables(),
						target: group.key,
					});
					this.$router.push(entry);
					//this.$store.commit('setTargetVariable', group.key);
				});
				container.appendChild(training);
				container.appendChild(target);
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
