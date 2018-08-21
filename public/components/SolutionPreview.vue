<template>
	<div>
		<div class="solution-preview" @click="onResult()">
			<div class="solution-header">
				<div>
					<strong>Dataset:</strong> {{solution.dataset}}
				</div>
				<div>
					<strong>Date:</strong> {{formattedTime}}
				</div>
			</div>
			<div class="solution-body">
				<div>
					<strong>Feature:</strong> {{solution.feature}}
				</div>
				<div>
					<b-badge v-if="isPending">
						{{status}}
					</b-badge>
					<b-badge variant="info" v-if="isRunning">
						{{status}}
					</b-badge>
					<div v-if="isCompleted">
						<b-badge variant="info" v-bind:key="score.metric" v-for="score in solution.scores">
							{{score.label}}: {{score.value.toFixed(2)}}
						</b-badge>
					</div>
					<div v-if="isErrored">
						<b-badge variant="danger">
							ERROR
						</b-badge>
					</div>
				</div>
			</div>
		</div>
		<div class="solution-progress">
			<b-progress v-if="isRunning"
				:value="percentComplete"
				variant="outline-secondary"
				striped
				:animated="true"></b-progress>
		</div>
	</div>
</template>

<script lang="ts">
import moment from 'moment';
import { createRouteEntry } from '../util/routes';
import { SOLUTION_PENDING, SOLUTION_RUNNING, SOLUTION_COMPLETED, SOLUTION_ERRORED, Solution } from '../store/solutions/index';
import { RESULTS_ROUTE } from '../store/route/index';
import Vue from 'vue';

export default Vue.extend({
	name: 'solution-preview',

	props: {
		solution: Object as () => Solution
	},

	computed: {
		percentComplete(): number {
			return 100;
		},
		formattedTime(): string {
			const t = moment(this.solution.timestamp);
			return t.format('MMM Do YYYY, h:mm:ss a');
		},
		status(): string {
			return this.solution.progress;
		},
		isPending(): boolean {
			return this.solution.progress === SOLUTION_PENDING;
		},
		isRunning(): boolean {
			return this.solution.progress === SOLUTION_RUNNING;
		},
		isCompleted(): boolean {
			return this.solution.progress === SOLUTION_COMPLETED;
		},
		isErrored(): boolean {
			return this.solution.progress === SOLUTION_ERRORED;
		},
		isBad(): boolean {
			return this.solution.isBad;
		}
	},

	methods: {
		onResult() {
			const entry = createRouteEntry(RESULTS_ROUTE, {
				dataset: this.solution.dataset,
				target: this.solution.feature,
				solutionId: this.solution.solutionId
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
