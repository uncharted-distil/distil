<template>
	<b-card header="Completed">
		<div class="completed-results" v-if="pipelineResults === null">None</div>
		<b-list-group class="completed-results card-text" v-bind:key="results.constructor.name" v-for="results in pipelineResults">
			<b-list-group-item href="#" v-bind:key="result.name" v-for="result in results">
				<div class="completed-result" @click="onResult(result)">
					<div class="completed-result-name">{{result.name}}</div>
					<div class="completed-result-badge">
						<b-badge variant="info" v-bind:key="score.metric" v-for="score in result.pipeline.scores">
							{{metricName(score.metric)}}: {{score.value}}
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
				createRequestId: result.requestId
			});
			this.$router.push(entry);
		}
	}
};
</script>

<style scoped>

.completed-results {
	margin-top: 8px;
}

.completed-result {
	display: flex;
	justify-content: flex-start;
	flex-grow: 1;
	margin-top: 8px;
}

.completed-result-name {
	display: flex;
	align-items: center;
	margin-right: 4px;
}

.completed-result-badge {
	display: flex;
	align-items: center;
	margin-right: 4px;
}

</style>
