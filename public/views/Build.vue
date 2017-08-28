<template>
	<div>
		<div class="row justify-content-center mt-2 mb-2">
			<create-pipelines-form class="col-md-12"></create-pipelines-form>
		</div>
		<div class="row justify-content-center mt-2 mb-2">
			<running-pipelines class="col-md-12"></running-pipelines>
		</div>
		<div class="row justify-content-center mt-2 mb-2">
			<completed-pipelines class="col-md-12"></completed-pipelines>
		</div>
	</div>
</template>

<script>
import CreatePipelinesForm from '../components/CreatePipelinesForm';
import RunningPipelines from '../components/RunningPipelines';
import CompletedPipelines from '../components/CompletedPipelines';

export default {
	name: 'pipelines',
	components: {
		CreatePipelinesForm,
		RunningPipelines,
		CompletedPipelines
	}
};
</script>

<style>

</style>
