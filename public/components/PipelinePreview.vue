<template>
	<div>
		<div class="pipeline-preview" @click="onResult()">
			<div class="pipeline-header">
				<div>
					<strong>dataset:</strong> {{result.dataset}}
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

<script lang="ts">
import moment from 'moment';
import { getMetricDisplayName } from '../util/pipelines';
import { createRouteEntry } from '../util/routes';
import { getters } from '../store/route/module';
import { PipelineInfo } from '../store/pipelines/index';
import { pushRoute } from '../util/routes';
import Vue from 'vue';

export default Vue.extend({
	name: 'pipeline-preview',

	props: {
		'result': Object
	},

	computed: {
		percentComplete(): number {
			return 100;
		},
		formattedTime(): string {
			const t = moment(this.result.timestamp);
			return t.format('MMM Do YYYY, h:mm:ss a');
		}
	},

	methods: {
		status(): string {
			const result = <PipelineInfo>this.result;
			if (result.progress === 'UPDATED') {
				const score = result.pipeline.scores[0];
				const metricName = getMetricDisplayName(score.metric);
				if (metricName) {
					return metricName + ': ' + score.value;
				}
				return score.value.toString();
			}
			return result.progress;
		},
		metricName(metric): string {
			return getMetricDisplayName(metric);
		},
		isSubmitted(): boolean {
			return (<PipelineInfo>this.result).progress==='SUBMITTED';
		},
		isRunning(): boolean {
			return (<PipelineInfo>this.result).progress==='RUNNING';
		},
		isUpdated(): boolean {
			return (<PipelineInfo>this.result).progress==='UPDATED';
		},
		isCompleted(): boolean {
			return (<PipelineInfo>this.result).progress !=='UPDATED' && (<PipelineInfo>this.result).pipeline !== undefined;
		},
		onResult() {
			const result = <PipelineInfo>this.result;
			const entry = createRouteEntry('/results', {
 				terms: getters.getRouteTerms(this.$store),
				dataset: result.dataset,
				filters: getters.getRouteFilters(this.$store),
				target: result.feature,
				training: getters.getRouteTrainingVariables(this.$store),
				requestId: result.requestId,
				pipelineId: result.pipelineId
			});
			pushRoute(this.$store, this.$router, entry);
		}
	}
});
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
