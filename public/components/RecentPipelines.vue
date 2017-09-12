<template>
	<b-card header="Recent Pipelines">
		<div v-if="recentPipelines === null">None</div>
		<b-list-group v-bind:key="results.constructor.name" v-for="results in recentPipelines">
			<b-list-group-item href="#" v-bind:key="result.name" v-for="result in results">
				<pipeline-preview :result="result"></pipeline-preview>
			</b-list-group-item>
		</b-list-group>
	</b-card>
</template>

<script>

export default {
	name: 'recent-pipelines',

	computed: {
		recentPipelines() {
			// TODO: implement this correctly
			return null;
		}
	}

};
</script>

<style>
</style>
