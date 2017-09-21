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
			return this.$store.getters.getAvailableVariables();
		},
		html() {
			return (group) => {
				const container = document.createElement('div');
				const trainingElem = document.createElement('button');
				trainingElem.className += 'btn btn-sm btn-outline-secondary mr-2 mb-2';
				trainingElem.innerHTML = 'Add to Training Set';
				trainingElem.addEventListener('click', () => {
					const path = this.$store.getters.getRoutePath();
					const training = this.$store.getters.getRouteTrainingVariables();
					const entry = createRouteEntryFromRoute(this.$store.getters.getRoute(), {
						training: training.concat([ group.key ])
					});
					this.$router.push(entry);
				});
				const targetElem = document.createElement('button');
				targetElem.className += 'btn btn-sm btn-outline-secondary mr-2 mb-2';
				targetElem.innerHTML = 'Set as Target';
				targetElem.addEventListener('click', () => {
					const path = this.$store.getters.getRoutePath();
					const entry = createRouteEntryFromRoute(this.$store.getters.getRoute(), {
						target: group.key,
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
