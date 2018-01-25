<template>
	<b-card header="Recent Models">
		<div v-if="recentPipelines.length === 0">None</div>
		<b-list-group v-bind:key="pipeline.timestamp" v-for="pipeline in recentPipelines">
			<b-list-group-item href="#" v-bind:key="pipeline.name">
				<pipeline-preview :result="pipeline"></pipeline-preview>
			</b-list-group-item>
		</b-list-group>
	</b-card>
</template>

<script lang="ts">

import PipelinePreview from '../components/PipelinePreview';
import { getters } from '../store/pipelines/module';
import { PipelineInfo } from '../store/pipelines/index';
import Vue from 'vue';

export default Vue.extend({
	name: 'recent-pipelines',

	components: {
		PipelinePreview
	},

	props: {
		maxPipelines: {
		default: 20,
			type: Number
		}
	},

	computed: {
		recentPipelines(): PipelineInfo[] {
			return getters.getPipelines(this.$store)
				.slice()
				.sort((a, b) => b.timestamp - a.timestamp)
				.slice(0, this.maxPipelines);
		}
	}
});
</script>

<style>
</style>
