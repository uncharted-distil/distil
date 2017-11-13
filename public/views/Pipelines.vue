<template>
	<div class="build-view">
		<h5 class="header-label">Monitor Pipeline Status</h5>
		<div class="build-results">
			<running-pipelines></running-pipelines>
			<completed-pipelines></completed-pipelines>
		</div>
	</div>
</template>

<script lange="ts">
import FlowBar from '../components/FlowBar';
import RunningPipelines from '../components/RunningPipelines';
import CompletedPipelines from '../components/CompletedPipelines';
import { getters } from '../store/app/module';
import { actions } from '../store/pipelines/module';
import Vue from 'vue';

export default Vue.extend({
	name: 'pipelines',
	components: {
		FlowBar,
		RunningPipelines,
		CompletedPipelines
	},

	computed: {
		sessionId() {
			return getters.getPipelineSessionID(this.$store);
		}
	},

	mounted() {
		actions.getSession(this.$store, {
			sessionId: this.sessionId
		});
	}
});
</script>

<style>
.header-label {
	color: #333;
	margin: 0.75rem 0;
}
.build-view {
	display: flex;
	flex-direction: column;
	align-items: center;
	margin: 8px;
}
.build-form {
	width: 50%;
}
.build-results {
	width: 80%;
	overflow: auto;
}
.build-results .card {
	margin-bottom: 4px;
}
</style>
