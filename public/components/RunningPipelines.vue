<template>
	<div class="running-pipelines">
		<div class="row">
			<div class="h6 col">Pending</div>
		</div>
		<div class="row mt-2 mb-2" v-if="pipelineResults === null">
			<div class="col">None</div>
		</div>
		<div class="row mt-2 mb-2" v-bind:key="result.name" v-for="result in pipelineResults">
			<div class="col-md-3">
				{{result.name}}
			</div>
			<div class="col-md-1">
				<b-badge variant="default" v-if="result.progress==='SUBMITTED'">{{status(result)}}</b-badge>
				<b-badge variant="info" v-if="result.progress!=='SUBMITTED'">{{status(result)}}</b-badge>
			</div>
		</div>
	</div>
</template>

<script>
import _ from 'lodash';
import {getMetricDisplayName} from '../util/pipelines';

export default {
	name: 'running-pipelines',

	//data change handlers
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
			console.log(result.progress);
			if (result.progress === 'UPDATED') {
				const score = result.pipeline.scores[0];
				const metricName = getMetricDisplayName(score.metric);
				return metricName + ': ' + score.value;
			}
			return result.progress;
		}
	}
};
</script>

<style>

</style>
