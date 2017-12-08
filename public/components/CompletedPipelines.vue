<template>
	<b-card header="Completed Pipelines">
		<div v-if="pipelineResults === null">None</div>
		<b-list-group v-bind:key="results.constructor.name" v-for="results in pipelineResults">
			<b-list-group-item href="#" v-bind:key="result.name" v-for="result in results">
				<pipeline-preview :result="result"></pipeline-preview>
			</b-list-group-item>
		</b-list-group>
	</b-card>
</template>

<script lang="ts">

import _ from 'lodash';
import PipelinePreview from '../components/PipelinePreview.vue';
import { getters } from '../store/pipelines/module';
import { PipelineInfo } from '../store/pipelines/index';
import { Dictionary } from '../util/dict';
import Vue from 'vue';

export default Vue.extend({
	name: 'completed-pipelines',

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
		pipelineResults(): Dictionary<PipelineInfo>[] {
			const pipelines = getters.getCompletedPipelines(this.$store);
			if (_.keys(pipelines).length > 0) {
				return _.values(pipelines).sort((a, b) => {
					return this.minResultTimestamp(b) - this.minResultTimestamp(a);
				}).slice(0, this.maxPipelines);
			}
			return null;
		}
	},

	methods: {
		minResultTimestamp(pipeline: Dictionary<PipelineInfo>): number {
			let min = Infinity;
			_.values(pipeline).forEach(result => {
				if (result.timestamp < min) {
					min = result.timestamp;
				}
			});
			return min;
		}
	}
});
</script>

<style>
</style>
