<template>
	<div class="build-view">
		<div class="build-results">
			<running-pipelines></running-pipelines>
			<completed-pipelines></completed-pipelines>
		</div>
	</div>
</template>

<script>
import RunningPipelines from '../components/RunningPipelines';
import CompletedPipelines from '../components/CompletedPipelines';

export default {
	name: 'pipelines',
	components: {
		RunningPipelines,
		CompletedPipelines
	},

	mounted() {
		this.$store.dispatch('getSession');
	}
};
</script>

<style>
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
	width: 50%;
	overflow: auto;
}
.build-results .card {
	margin-bottom: 4px;
}
</style>
