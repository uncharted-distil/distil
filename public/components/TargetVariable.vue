<template>
	<div class='target-variables'>
		<div class="bg-faded rounded-top">
			<h6 class="nav-link">Target Variable</h6>
		</div>
		<variable-facets
			:variables="variables"
			:dataset="dataset"
			:html="html"></variable-facets>
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
			const target = this.$store.getters.getTargetVariable();
			if (target) {
				return [ target ];
			}
			return [];
		},
		html() {
			return () => {
				const container = document.createElement('div');
				const remove = document.createElement('button');
				remove.className += 'btn btn-sm btn-outline-secondary mb-2';
				remove.innerHTML = 'Remove';
				remove.addEventListener('click', () => {
					const path = this.$store.getters.getRoutePath();
					const entry = createRouteEntryFromRoute(this.$store.getters.getRoute(), {
						target: null,
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
