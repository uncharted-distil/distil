<template>
	<b-card header="Pending Models">
		<div v-if="runningSolutions.length === 0">None</div>
		<b-list-group v-bind:key="solution.timestamp" v-for="solution in runningSolutions">
			<solution-preview :result="solution"></solution-preview>
		</b-list-group>
	</b-card>
</template>

<script lang="ts">

import SolutionPreview from '../components/SolutionPreview';
import { getters } from '../store/solutions/module';
import { SolutionInfo } from '../store/solutions/index';
import Vue from 'vue';

export default Vue.extend({
	name: 'running-solutions',

	props: {
		maxSolutions: {
			default: 20,
			type: Number
		}
	},

	components: {
		SolutionPreview
	},

	computed: {
		runningSolutions(): SolutionInfo[] {
			return getters.getRunningSolutions(this.$store)
				.slice()
				.sort((a, b) => b.timestamp - a.timestamp)
				.slice(0, this.maxSolutions);
		}
	}
});
</script>

<style>
</style>
