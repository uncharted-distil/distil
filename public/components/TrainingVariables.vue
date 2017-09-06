<template>
	<div class='training-variables'>
		<div class="bg-faded rounded-top">
			<h6 class="nav-link">Training Set</h6>
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
	name: 'training-variables',

	components: {
		VariableFacets
	},

	computed: {
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		variables() {
			return this.$store.getters.getTrainingVariables();
		}
	},

	methods: {
		onDragEnd(key) {
			this.$store.commit('removeTrainingVariable', key);
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
