<template>
	<div class="create-pipelines-form">
		Feature:
		<b-dropdown :text="feature" class="m-md-2">
			<b-dropdown-item :key="variable" v-for="variable in variables" @click="featureSelect">{{variable}}</b-dropdown-item>
		</b-dropdown>
		Metric:
		<b-dropdown :text="metric" class="m-md-2">
			<b-dropdown-item :disabled="!featureSet" :key="metric" v-for="metric in metrics" @click="metricSelect">{{metric}}</b-dropdown-item>
		</b-dropdown>
		<b-button :variant="createVariant" @click="create" :disabled="disableCreate">
			Create
		</b-button>
	</div>
</template>

<script>

import _ from 'lodash';
import {getTask, getMetricDisplayNames, getOutputSchemaNames, getMetricSchemaName} from '../util/pipelines';

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
		// gets the variables associated with the currently selected dataset
		variables() {
			return _.map(this.$store.getters.getVariables(), v => v.name);
		},
		// gets the metrics that are used to score predictions against the user selected variable
		metrics() {
			// get the variable entry from the store that matches the user selection
			const variables = this.$store.getters.getVariables();
			if (_.isEmpty(variables)) {
				return [];
			}
			const variable = _.find(variables, v => _.toLower(v.name) === _.toLower(this.feature));
			if (_.isEmpty(variable)) {
				return [];
			}

			// get the task info associated with that variable type
			const taskData = getTask(variable.type);

			// grab the valid metrics from the task data to use as labels in the UI
			return getMetricDisplayNames(taskData);
		},
		// determines create button status based on completeness of user input
		disableCreate() {
			return !(this.featureSet && this.metricSet);
		},
		// determines  create button variant based on completeness of user input
		createVariant() {
			const allSet = this.featureSet && this.metricSet;
			return allSet ? 'success' : 'warning';
		}
	},
	methods: {
		// feature selection handler
		featureSelect(evt) {
			this.feature = evt.target.text;
			this.featureSet = true;
		},
		// metric selection handler
		metricSelect(evt) {
			this.metric = evt.target.text;
			this.metricSet = true;
		},
		// create button handler
		create() {
			// compute schema values for request
			const variables = this.$store.getters.getVariables();
			const variable = _.find(variables, v => _.toLower(v.name) === _.toLower(this.feature));

			const taskData = getTask(variable.type);
			const task = taskData.schemaName;
			const output = getOutputSchemaNames(taskData)[0];
			const metric = getMetricSchemaName(taskData, this.metric);

			// dispatch action that triggers request send to server
			this.$store.dispatch('createPipelines', {
				feature: this.feature,
				task: task,
				metric: metric,
				output: output
			});
		}
	}
};
</script>

<style>

</style>
