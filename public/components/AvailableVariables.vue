<template>
	<div class='available-variables'>
		<div class="bg-faded rounded-top">
			<h6 class="nav-link">Available Set</h6>
		</div>
		<variable-facets
			:variables="variables"
			:dataset="dataset"
			:drag-end="onDragEnd"></variable-facets>
	</div>
</template>

<script>

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
		}
	},

	methods: {
		onDragEnd(key) {
			this.$store.commit('addTrainingVariable', key);
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
