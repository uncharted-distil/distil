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

<script>

import _ from 'lodash';
import PipelinePreview from '../components/PipelinePreview';
import {getMetricDisplayName} from '../util/pipelines';

export default {
	name: 'completed-pipelines',

	components: {
		PipelinePreview
	},

	computed: {
		pipelineResults() {
			if (_.keys(this.$store.state.completedPipelines).length > 0) {
				return this.$store.state.completedPipelines;
			}
			return null;
		},
	},
	methods: {
		metricName(metric) {
			return getMetricDisplayName(metric);
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
