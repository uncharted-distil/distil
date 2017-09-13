<template>
	<b-card header="Pending Pipelines">
		<div v-if="pipelineResults === null">None</div>
		<b-list-group v-bind:key="results.constructor.name" v-for="results in pipelineResults">
			<b-list-group-item href="#" v-bind:key="result.name" v-for="result in results">
				<pipeline-preview :result="result"></pipeline-preview>
			</b-list-group-item>
		</b-list-group>
	</b-card>
</template>

<script>
import _ from 'lodash';
import PipelinePreview from '../components/PipelinePreview';
import { getMetricDisplayName } from '../util/pipelines';
import { createRouteEntry } from '../util/routes';

export default {
	name: 'running-pipelines',

	components: {
		PipelinePreview
	},

	computed: {
		pipelineResults() {
			if (_.keys(this.$store.state.runningPipelines).length > 0) {
				return this.$store.state.runningPipelines;
			} else {
				return null;
			}
		}
	},
	methods: {
		status(result) {
			if (result.progress === 'UPDATED') {
				const score = result.pipeline.scores[0];
				const metricName = getMetricDisplayName(score.metric);
				return metricName + ': ' + score.value;
			}
			return result.progress;
		},
		onResult(result) {
			const entry = createRouteEntry('/results', {
				dataset: this.$store.getters.getRouteDataset(),
				filters: this.$store.getters.getRouteFilters(),
				createRequestId: result.requestId
			});
			this.$router.push(entry);
		}
	}
};
</script>

<style>
</style>
