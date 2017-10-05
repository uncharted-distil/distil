<template>
	<div>
		<div class="pipeline-preview" @click="onResult()">
			<div class="pipeline-header">
				<div>
					<strong>Dataset:</strong> {{result.dataset}}
				</div>
				<div>
					<strong>Date:</strong> {{formattedTime}}
				</div>
			</div>
			<div class="pipeline-body">
				<div>
					<strong>Feature:</strong> {{result.feature}}
				</div>
				<div>
					<b-badge v-if="isSubmitted()">
						{{status()}}
					</b-badge>
					<b-badge variant="info" v-if="isRunning()">
						{{status()}}
					</b-badge>
					<div v-if="isUpdated()">
						<b-badge variant="info" v-bind:key="score.metric" v-for="score in result.pipeline.scores">
							{{metricName(score.metric)}}: {{score.value}}
						</b-badge>
					</div>
					<div v-if="isCompleted()">
						<b-badge variant="info" v-bind:key="score.metric" v-for="score in result.pipeline.scores">
							{{metricName(score.metric)}}: {{score.value}}
						</b-badge>
					</div>
					</b-badge>
				</div>
			</div>
		</div>
		<div class="pipeline-progress">
			<b-progress v-if="isRunning()"
				:value="percentComplete"
				variant="secondary"
				striped
				:animated="true"></b-progress>
		</div>
	</div>
</template>

<script>
import moment from 'moment';
import { getMetricDisplayName } from '../util/pipelines';
import { createRouteEntry } from '../util/routes';

export default {
	name: 'pipeline-preview',

	props: [
		'result'
	],

	computed: {
		percentComplete() {
			return 100;
		},
		formattedTime() {
			const t = moment(this.result.timestamp);
			return t.format('MMM Do YYYY, h:mm:ss a');
		}
	},

	methods: {
		status() {
			if (this.result.progress === 'UPDATED') {
				const score = this.result.pipeline.scores[0];
				const metricName = getMetricDisplayName(score.metric);
				if (metricName) {
					return metricName + ': ' + score.value;
				}
				return score.value;
			}
			return this.result.progress;
		},
		metricName(metric) {
			return getMetricDisplayName(metric);
		},
		isSubmitted() {
			return this.result.progress==='SUBMITTED';
		},
		isRunning() {
			return this.result.progress==='RUNNING';
		},
		isUpdated() {
			return this.result.progress==='UPDATED';
		},
		isCompleted() {
			return this.result.progress !=='UPDATED' && this.result.pipeline !== undefined;
		},
		onResult() {
			 const entry = createRouteEntry('/results', {
 				terms: this.$store.getters.getRouteTerms(),
				dataset: this.result.dataset,
				filters: this.$store.getters.getRouteFilters(),
				target: this.$store.getters.getRouteTargetVariable(),
				training: this.$store.getters.getRouteTrainingVariables(),
				createRequestId: this.result.requestId,
				resultId: btoa(this.result.pipeline.resultUri)
			});
			this.$router.push(entry);
		}
	}
};
</script>

<style>
.pipeline-preview {
	display: flex;
	flex-direction: column;
}
.pipeline-header {
	display: flex;
	justify-content: space-between;
}
.pipeline-body {
	display: flex;
	justify-content: space-between;
}
.pipeline-preview .badge {
	display: block;
	margin: 4px 0;
}
.pipeline-progress {
	margin: 6px 0;
}
</style>
