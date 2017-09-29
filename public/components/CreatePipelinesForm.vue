<template>
	<div class="create-pipelines-form">
		<b-button class="full-width" :variant="createVariant" @click="create" :disabled="disableCreate">
			Create Pipelines
		</b-button>
	</div>
</template>

<script>

import _ from 'lodash';
import { createRouteEntry } from '../util/routes';
import { getTask, getMetricDisplayNames, getOutputSchemaNames, getMetricSchemaName } from '../util/pipelines';

export default {
	name: 'create-pipelines-form',
	data() {
		return {
			descriptionText: '',
			feature: 'Feature',
			featureSet: false,
			metric: 'Metric',
			metricSet: false
		};
	},
	computed: {
		variables() {
			return this.$store.getters.getVariables();
		},
		// gets the metrics that are used to score predictions against the user selected variable
		metrics() {
			// get the variable entry from the store that matches the user selection
			if (!this.target || _.isEmpty(this.variables)) {
				return [];
			}
			// get the task info associated with that variable type
			const taskData = getTask(this.targetVariable.type);
			// grab the valid metrics from the task data to use as labels in the UI
			return getMetricDisplayNames(taskData);
		},
		target() {
			return this.$store.getters.getTargetVariable();
		},
		targetVariable() {
			return _.find(this.variables, v => {
				return _.toLower(v.name) === _.toLower(this.target);
			});
		},
		// determines create button status based on completeness of user input
		disableCreate() {
			return !this.target;
		},
		// determines  create button variant based on completeness of user input
		createVariant() {
			return !this.disableCreate ? 'primary' : 'secondary';
		}
	},
	methods: {
		// create button handler
		create() {
			// compute schema values for request
			const taskData = getTask(this.targetVariable.type);
			const task = taskData.schemaName;
			const output = _.values(getOutputSchemaNames(taskData))[0];
			const metrics = _.map(this.metrics, m => getMetricSchemaName(m));

			// dispatch action that triggers request send to server
			this.$store.dispatch('createPipelines', {
				feature: this.$store.getters.getRouteTargetVariable(),
				task: task,
				metric: metrics,
				output: output
			});

			// transition to build screen
			const entry = createRouteEntry('/pipelines', {
				dataset: this.$store.getters.getRouteDataset(),
				filters: this.$store.getters.getRouteFilters(),
				target: this.$store.getters.getRouteTargetVariable(),
				training: this.$store.getters.getRouteTrainingVariables()
			});
			this.$router.push(entry);
		}
	}
};
</script>

<style>
.create-pipelines-form {
	margin: 8px 16px;
}
.full-width {
	width: 100%;
}
</style>
