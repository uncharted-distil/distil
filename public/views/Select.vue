<template>
	<div class="select-view">
		<div class="left-container">
			<h4 class="header-label">Select the Training Features</h4>
			<div class="select-items">
				<available-variables class="select-available-variables"></available-variables>
				<training-variables class="select-training-variables"></training-variables>
			</div>
		</div>
		<div class="right-container">
			<h4 class="header-label">Select the Target Feature</h4>
			<target-variable class="select-target-variables"></target-variable>
			<select-data-table class="select-data-table"></select-data-table>
			<h4 class="header-label">Create the Pipelines</h4>
			<create-pipelines-form class="select-create-pipelines"></create-pipelines-form>
		</div>
	</div>
</template>

<script>
import FlowBar from '../components/FlowBar';
import CreatePipelinesForm from '../components/CreatePipelinesForm';
import SelectDataTable from '../components/SelectDataTable';
import AvailableVariables from '../components/AvailableVariables';
import TrainingVariables from '../components/TrainingVariables';
import TargetVariable from '../components/TargetVariable';

export default {
	name: 'select',

	components: {
		FlowBar,
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
.select-view {
	display: flex;
	justify-content: space-around;
	padding: 8px;
}
.header-label {
	color: #333;
	margin: 0.75rem 0;
}
.select-items {
	display: flex;
	justify-content: space-around;
	padding: 8px;
	width: 100%;
}
.select-available-variables {
	width: 50%;
}
.select-training-variables {
	width: 50%;
}
.left-container {
	display: flex;
	flex-direction: column;
	justify-content: space-around;
	padding: 8px;
	width: 50%;
}
.right-container {
	display: flex;
	flex-direction: column;
	width: 50%;
}
</style>
