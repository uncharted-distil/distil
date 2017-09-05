<template>
	<div class="completed-pipelines">
		<div class="row h6">
			<div class="col">Completed</div>
		</div>
		<div class="row mt-2 mb-2" v-if="pipelineResults === null">
			<div class="col">None</div>
		</div>
		<div class="row mt-2 mb-2" v-bind:key="result.name" v-for="result in pipelineResults">
			<div class="col-md-3 ">
				<a class="text-primary" @click="onResult(result)">{{result.name}}</a>
			</div>
			<div class="col-md-1">
				<b-badge variant="primary" v-bind:key="score.metric" v-for="score in result.pipeline.scores">
					{{metricName(score.metric)}}: {{score.value}}
				</b-badge>
			</div>
		</div>
	</div>
</template>

<script>

import _ from 'lodash';
import {getMetricDisplayName} from '../util/pipelines';
import { createRouteEntry } from '../util/routes';

export default {
	name: 'completed-pipelines',

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
			console.log(result);
			const entry = createRouteEntry('/results', {
				dataset: this.$store.getters.getRouteDataset(),
				filters: this.$store.getters.getRouteFilters(),
				results: encodeURIComponent(result.pipeline.resultUri)
			});
			this.$router.push(entry);
		}
	}
};
</script>

<style>

</style>
