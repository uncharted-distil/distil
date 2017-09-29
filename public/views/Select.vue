<template>
	<div class="select">
		<available-variables class="select-available-variables"></available-variables>
		<training-variables class="select-training-variables"></training-variables>
		<div class="side-container">
			<target-variable class="select-target-variables"></target-variable>
			<create-pipelines-form class="select-create-pipelines"></create-pipelines-form>
			<select-data-table class="select-data-table"></select-data-table>
		</div>
	</div>
</template>

<script>
import CreatePipelinesForm from '../components/CreatePipelinesForm';
import SelectDataTable from '../components/SelectDataTable';
import AvailableVariables from '../components/AvailableVariables';
import TrainingVariables from '../components/TrainingVariables';
import TargetVariable from '../components/TargetVariable';

export default {
	name: 'select',

	components: {
		CreatePipelinesForm,
		SelectDataTable,
		AvailableVariables,
		TrainingVariables,
		TargetVariable
	},

	computed: {
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		variables() {
			return this.$store.getters.getVariables();
		}
	},

	mounted() {
		this.fetch();
	},

	watch: {
		'$route.query.dataset'() {
			this.fetch();
		},
		'$route.query.training'() {
		},
	},

	methods: {
		fetch() {
			this.$store.dispatch('getVariables', this.dataset)
				.then(() => {
					this.$store.dispatch('getVariableSummaries', {
						dataset: this.dataset,
						variables: this.variables
					});
				});
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
