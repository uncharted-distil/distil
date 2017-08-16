<template>
	<div class="create-pipelines-form">
		Feature:
		<b-dropdown :text="feature" class="m-md-2">
			<b-dropdown-item :key="variable.name" v-for="variable in variables" @click="featureSelect">{{variable.name}}</b-dropdown-item>
		</b-dropdown>
		Task:
		<b-dropdown :text="task" class="m-md-2">
			<b-dropdown-item :key="task" v-for="task in tasks" @click="taskSelect">{{task}}</b-dropdown-item>
		</b-dropdown>
		Metric:
		<b-dropdown :text="metric" class="m-md-2">
			<b-dropdown-item :key="metric" v-for="metric in metrics" @click="metricSelect">{{metric}}</b-dropdown-item>
		</b-dropdown>
		Output:
		<b-dropdown :text="output" class="m-md-2">
			<b-dropdown-item :key="output" v-for="output in outputs" @click="outputSelect">{{output}}</b-dropdown-item>
		</b-dropdown>
		<b-button :variant="createVariant" @click="create" :disabled="disableCreate">
			Create
		</b-button>
	</div>
</template>

<script>

export default {
	name: 'create-pipelines-form',
	data() {
		return {
			descriptionText: '',
			feature: 'Feature',
			featureSet: false,
			task: 'Task',
			taskSet: false,
			tasks: [
				'regression',
				'classification'
			],
			metric: 'Metric',
			metricSet: false,
			metrics: [
				'accuracy',
				'precision',
				'recall',
				'f1_micro',
				'f1_macro',
				'roc-auc',
				'log_loss',
				'mean_squared_err',
				'mean_absolute_err',
				'median_absolute_err',
				'r2'
			],
			output: 'Output',
			outputSet: false,
			outputs: [
				'class_label',
				'probability',
				'general_score',
				'multilabel',
				'regression_value'
			]
		};
	},
	mounted() {
		// make sure variables are immediately available so they can be added to the
		// dropdown
		this.$store.dispatch('getVariables', this.$store.getters.getRouteDataset());
	},
	watch: {
		'$route.query'() {
			this.$store.dispatch('getVariables', this.$store.getters.getRouteDataset());
		}
	},
	computed: {
		variables() {
			return this.$store.getters.getVariables();
		},
		disableCreate() {
			return !(this.featureSet && this.taskSet && this.metricSet && this.outputSet);
		},
		createVariant() {
			const allSet = this.featureSet && this.taskSet && this.metricSet && this.outputSet;
			return allSet ? 'success' : 'warning';
		}
	},
	methods: {
		featureSelect(evt) {
			this.feature = evt.target.text;
			this.featureSet = true;
		},
		taskSelect(evt) {
			this.task = evt.target.text;
			this.taskSet = true;
		},
		outputSelect(evt) {
			this.output = evt.target.text;
			this.outputSet = true;
		},
		metricSelect(evt) {
			this.metric = evt.target.text;
			this.metricSet = true;
		},
		create() {
			this.$store.dispatch('createPipelines', {
				feature: this.feature,
				task: this.task,
				metric: this.metric,
				output: this.output
			});
		}
	}
};
</script>

<style>

</style>
