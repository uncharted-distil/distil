<template>
	<div class="select">
		<available-variables class="select-available-variables"></available-variables>
		<training-variables class="select-training-variables"></training-variables>
		<div class="side-container">
			<target-variable class="select-target-variables"></target-variable>
			<data-table class="select-data-table"></data-table>
		</div>
	</div>
</template>

<script>
import DataTable from '../components/DataTable';
import AvailableVariables from '../components/AvailableVariables';
import TrainingVariables from '../components/TrainingVariables';
import TargetVariable from '../components/TargetVariable';

export default {
	name: 'select',

	components: {
		DataTable,
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
		}/*,
		'$route.query.training'() {
			const training = this.$store.getters.getRouteTrainingVariables();
			this.$store.commit('setTrainingVariables', training);
			//this.$store.commit('addTrainingVariable', group.key);
		},
		'$route.query.target'() {
			const target = this.$store.getters.getRouteTrainingVariables();
			this.$store.commit('setTargetVariable', target);
			//this.$store.commit('addTrainingVariable', group.key);
		}*/
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
