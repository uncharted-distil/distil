<template>
	<b-card header="Pending">
		<div class="results" v-if="pipelineResults === null">None</div>
		<b-list-group class="results card-text" v-bind:key="results.constructor.name" v-for="results in pipelineResults">
			<b-list-group-item href="#" v-bind:key="result.name" v-for="result in results">
				<div class="result" @click="onResult(result)">
					<div class="result-name">
						{{result.name}}
					</div>
					<div class="result-badge">
						<b-badge variant="default" v-if="result.progress==='SUBMITTED'">
							{{status(result)}}
						</b-badge>
					</div>
					<div class="result-badge">
						<b-badge variant="info" v-if="result.progress!=='SUBMITTED'">
							{{status(result)}}
						</b-badge>
					</div>
				</div>
			</b-list-group-item>
		</b-list-group>
	</b-card>
</template>

<script>
import _ from 'lodash';
import {getMetricDisplayName} from '../util/pipelines';
import { createRouteEntry } from '../util/routes';

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
			if (result.progress === 'UPDATED') {
				const score = result.pipeline.scores[0];
				const metricName = getMetricDisplayName(score.metric);
				return metricName + ': ' + score.value;
			}
			return result.progress;
			return 'score';
		},
		onResult(result) {
			console.log(result);
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

<style scoped>

.results {
	margin-top: 8px;
}

.result {
	display: flex;
	justify-content: flex-start;
	flex-grow: 1;
}

.result-name {
	display: flex;
	flex-basis: 20%;
	align-items: center;
}

.result-badge {
	display: flex;
	align-items: center;
}

</style>
