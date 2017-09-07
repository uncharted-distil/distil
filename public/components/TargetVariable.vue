<template>
	<div class='target-variables'>
		<div class="bg-faded rounded-top">
			<h6 class="nav-link">Target Variable</h6>
		</div>
		<variable-facets
			enable-filter="false"
			enable-toggle="false"
			:variables="variables"
			:dataset="dataset"
			:html="html"></variable-facets>
	</div>
</template>

<script>

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
			const target = this.$store.getters.getTargetVariable();
			if (target) {
				return [ target ];
			}
			return [];
		},
		html() {
			return (group) => {
				const container = document.createElement('div');
				const remove = document.createElement('button');
				remove.className += 'btn btn-sm btn-outline-secondary mb-2';
				remove.innerHTML = 'Remove';
				remove.addEventListener('click', () => {
					this.$store.commit('removeTargetVariable', group.key);
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
