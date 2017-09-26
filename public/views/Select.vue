<template>
	<div class="select">
		<available-variables class="select-available-variables"></available-variables>
		<training-variables class="select-training-variables"></training-variables>
		<div class="side-container">
			<target-variable class="select-target-variables"></target-variable>
			<select-data-table class="select-data-table"></select-data-table>
		</div>
	</div>
</template>

<script>
import SelectDataTable from '../components/SelectDataTable';
import AvailableVariables from '../components/AvailableVariables';
import TrainingVariables from '../components/TrainingVariables';
import TargetVariable from '../components/TargetVariable';

export default {
	name: 'select',

	components: {
		SelectDataTable,
		AvailableVariables,
		TrainingVariables,
		TargetVariable
	},

	mounted() {
		const dataset = this.$store.getters.getRouteDataset();
		this.$store.dispatch('getVariableSummaries', dataset);
	},

	watch: {
		'$route.query.dataset'() {
			const dataset = this.$store.getters.getRouteDataset();
			this.$store.dispatch('getVariableSummaries', dataset);
		}
	}
};
</script>

<style>
.select {
	display: flex;
	justify-content: space-around;
	padding: 8px;
}
.select-available-variables {
	width: 30%;
}
.select-training-variables {
	width: 30%;
}
.side-container {
	display: flex;
	flex-direction: column;
	width: 40%;
}
</style>
