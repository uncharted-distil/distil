<template>
	<b-card header="Recent Models">
		<div v-if="recentSolutions.length === 0">None</div>
		<b-list-group v-bind:key="solution.timestamp" v-for="solution in recentSolutions">
			<b-list-group-item href="#" v-bind:key="solution.name">
				<solution-preview :result="solution"></solution-preview>
			</b-list-group-item>
		</b-list-group>
	</b-card>
</template>

<script lang="ts">

import SolutionPreview from '../components/SolutionPreview';
import { getters } from '../store/solutions/module';
import { SolutionInfo } from '../store/solutions/index';
import Vue from 'vue';

export default Vue.extend({
	name: 'recent-solutions',

	components: {
		SolutionPreview
	},

	props: {
		maxSolutions: {
		default: 20,
			type: Number
		}
	},

	computed: {
		recentSolutions(): SolutionInfo[] {
			return getters.getSolutions(this.$store)
				.slice()
				.sort((a, b) => b.timestamp - a.timestamp)
				.slice(0, this.maxSolutions);
		}
	}
});
</script>

<style>
</style>
