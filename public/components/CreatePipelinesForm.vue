<template>
	<div class="create-pipelines-form">
		<div class="requirements">
			<div class="requirement-met text-success" v-if="trainingSelected">
				<i class="fa fa-check selected-icon"></i><strong>Training Features Selected</strong>
			</div>
			<div class="requirement-met text-success" v-if="targetSelected">
				<i class="fa fa-check selected-icon"></i><strong>Target Feature Selected</strong>
			</div>
		</div>
		<b-button class="create-button" :variant="createVariant" @click="create" :disabled="disableCreate">
			Create Pipelines
		</b-button>
	</div>
</template>

<script lang="ts">

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
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		variables() {
			return this.$store.getters.getVariables();
		},
		selectedFilters() {
			return this.$store.getters.getSelectedFilters();
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
		trainingSelected() {
			return !_.isEmpty(this.training);
		},
		targetSelected() {
			return !!this.target;
		},
		training() {
			return this.$store.getters.getTrainingVariables();
		},
		target() {
			return this.$store.getters.getTargetVariable();
		},
		targetVariable() {
			return _.find(this.variables, v => {
				return _.toLower(v.name) === _.toLower(this.target);
			});
		},
		sessionId() {
			return this.$store.getters.getPipelineSessionID();
		},
		// determines create button status based on completeness of user input
		disableCreate() {
			return !this.targetSelected || !this.trainingSelected;
		},
		// determines  create button variant based on completeness of user input
		createVariant() {
			return !this.disableCreate ? 'outline-success' : 'outline-secondary';
		}
	},
	methods: {
		// create button handler
		create() {
			// compute schema values for request
			const taskData = getTask(this.targetVariable.type);
			const task = taskData.schemaName;
			const output = _.values(getOutputSchemaNames(taskData))[0];
			const metrics = _.map(this.computed.metrics() as string[], m => getMetricSchemaName(m));

			// dispatch action that triggers request send to server
			this.$store.dispatch('createPipelines', {
				dataset: this.dataset,
				filters: this.selectedFilters,
				sessionId: this.sessionId,
				feature: this.$store.getters.getRouteTargetVariable(),
				task: task,
				metric: metrics,
				output: output
			});

			// transition to build screen
			const entry = createRouteEntry('/pipelines', {
				terms: this.$store.getters.getRouteTerms(),
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
.create-button {
	width: 60%;
	margin: 0 20%;
}
.selected-icon {
	padding-right: 4px;
}
.requirement-met {
	padding: 4px 8px;
}
.requirements {
	margin-bottom: 8px;
}
</style>
