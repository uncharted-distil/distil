<template>
	<b-card header="Recent Pipelines">
		<div v-if="recentPipelines === null">None</div>
		<b-list-group v-bind:key="results.constructor.name" v-for="pipeline in recentPipelines">
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
			const pipelines = getters.getPipelines(this.$store);
			if (pipelines.length > 0) {
				return pipelines.sort((a, b) => {
					return b.timestamp - a.timestamp;
				}).slice(0, this.maxPipelines);
			}
			return null;
		}
	}
});
</script>

<style>
</style>
