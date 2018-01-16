<template>
	<b-card header="Pending Pipelines">
		<div v-if="runningPipelines.length === 0">None</div>
		<b-list-group v-bind:key="pipeline.timestamp" v-for="pipeline in runningPipelines">
			<pipeline-preview :result="pipeline"></pipeline-preview>
		</b-list-group>
	</b-card>
</template>

<script lang="ts">

import PipelinePreview from '../components/PipelinePreview';
import { getters } from '../store/pipelines/module';
import { PipelineInfo } from '../store/pipelines/index';
import Vue from 'vue';

export default Vue.extend({
	name: 'running-pipelines',

	props: {
		maxPipelines: {
			default: 20,
			type: Number
		}
	},

	components: {
		PipelinePreview
	},

	computed: {
		runningPipelines(): PipelineInfo[] {
			return getters.getRunningPipelines(this.$store)
				.slice()
				.sort((a, b) => b.timestamp - a.timestamp)
				.slice(0, this.maxPipelines);
		}
	}
});
</script>

<style>
</style>
