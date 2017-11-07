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
import _ from 'lodash';
import PipelinePreview from '../components/PipelinePreview';
import { getters } from '../store/pipelines/module';
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
		recentPipelines() {
			const pipelines = getters.getCompletedPipelines(this.$store);
			if (_.keys(pipelines).length > 0) {
				return _.values(pipelines).sort((a, b) => {
					return this.minResultTRimestamp(b) - this.minResultTRimestamp(a);
				}).slice(0, this.maxPipelines);
			}
			return null;
		}
	},

	methods: {
		minResultTRimestamp(pipeline) {
			let min = Infinity;
			_.values(pipeline).forEach(result => {
				if (result.createdTime < min) {
					min = result.createdTime;
				}
			});
			return min;
		}
	}
});
</script>

<style>
</style>
