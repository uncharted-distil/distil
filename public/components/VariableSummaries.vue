<template>
	<div id="variable-summaries">
		<div v-for="variable in variables">
			{{variable.name}}:{{variable.type}}
		</div>
	</div>
</template>

<script>

import _ from 'lodash';

export default {
	name: 'variable-summaries',	
	computed: {
		variables() {
			const activeDataset = this.$store.state.activeDataset;
			if (!_.isEmpty(activeDataset)) {
				return this.$store.getters.getVariables(activeDataset);
			} else {
				return [];
			}
		}
	}
};
</script>

<style>

</style>
