<template>
	<div class="select">
		<available-variables class="select-available-variables"></available-variables>
		<training-variables class="select-training-variables"></training-variables>
		<data-table class="select-data-table"></data-table>
	</div>
</template>

<script>
import DataTable from '../components/DataTable';
import AvailableVariables from '../components/AvailableVariables';
import TrainingVariables from '../components/TrainingVariables';

export default {
	name: 'select',

	components: {
		DataTable,
		AvailableVariables,
		TrainingVariables
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
.select-data-table {
	width: 40%;
}
</style>
