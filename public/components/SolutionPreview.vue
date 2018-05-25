<template>
	<div>
		<div class="solution-preview" @click="onResult()">
			<div class="solution-header">
				<div>
					<strong>Dataset:</strong> {{result.dataset}}
				</div>
				<div>
					<strong>Date:</strong> {{formattedTime}}
				</div>
			</div>
			<div class="solution-body">
				<div>
					<strong>Feature:</strong> {{result.feature}}
				</div>
				<div>
					<b-badge v-if="isPending()">
						{{status()}}
					</b-badge>
					<b-badge variant="info" v-if="isRunning()">
						{{status()}}
					</b-badge>
					<div v-if="isCompleted()">
						<b-badge variant="info" v-bind:key="score.metric" v-for="score in result.scores">
							{{metricName(score.metric)}}: {{score.value}}
						</b-badge>
					</div>
					<div v-if="isErrored()">
						<b-badge variant="danger">
							ERROR
						</b-badge>
					</div>
					</b-badge>
				</div>
			</div>
		</div>
		<div class="solution-progress">
			<b-progress v-if="isRunning()"
				:value="percentComplete"
				variant="outline-secondary"
				striped
				:animated="true"></b-progress>
		</div>
	</div>
</template>

<script lang="ts">
import moment from 'moment';
import { getMetricDisplayName } from '../util/solutions';
import { createRouteEntry } from '../util/routes';
import { Solution, SOLUTION_PENDING, SOLUTION_RUNNING, SOLUTION_COMPLETED, SOLUTION_ERRORED } from '../store/solutions/index';
import { RESULTS_ROUTE } from '../store/route/index';
import Vue from 'vue';

export default Vue.extend({
	name: 'solution-preview',

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
			return this.result.progress;
		},
		metricName(metric): string {
			return getMetricDisplayName(metric);
		},
		isPending(): boolean {
			return (<Solution>this.result).progress === SOLUTION_PENDING;
		},
		isRunning(): boolean {
			return (<Solution>this.result).progress === SOLUTION_RUNNING;
		},
		isCompleted(): boolean {
			return (<Solution>this.result).progress === SOLUTION_COMPLETED;
		},
		isErrored(): boolean {
			return (<Solution>this.result).progress === SOLUTION_ERRORED;
		},
		onResult() {
			const result = <Solution>this.result;
			const entry = createRouteEntry(RESULTS_ROUTE, {
				dataset: result.dataset,
				target: result.feature,
				solutionId: result.solutionId
			});
			this.$router.push(entry);
		}
	}
});
</script>

<style>
.solution-preview {
	display: flex;
	flex-direction: column;
}
.solution-header {
	display: flex;
	justify-content: space-between;
}
.solution-body {
	display: flex;
	justify-content: space-between;
}
.solution-preview .badge {
	display: block;
	margin: 4px 0;
}
.solution-progress {
	margin: 6px 0;
}
</style>
