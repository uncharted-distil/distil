<template>
	<div class="create-pipelines-form">
		<b-modal id="export-modal" ref="exportModal" title="Export Succeeded"
			@hide="clearExportResults"
			:visible="!!exportResults"
			cancel-disabled
			hide-header
			hide-footer>
			<div class="row justify-content-center">Export Succeeded</div>
			<div class="row justify-content-center">
				<b-btn class="mt-3 close-modal" variant="success" block @click="clearExportResults">OK</b-btn>
			</div>
		</b-modal>
		<div class="row justify-content-center">
			<b-button class="create-button" :variant="createVariant" @click="create" :disabled="disableCreate">
				Create Models
			</b-button>
			<b-button class="export-button" :variant="exportVariant" @click="exportData" :disabled="disableCreate">
				Export Problem
			</b-button>
		</div>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import { createRouteEntry } from '../util/routes';
import { getTask, getMetricDisplayNames, getMetricSchemaName } from '../util/pipelines';
import { getters as dataGetters, actions as dataActions } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { RESULTS_ROUTE } from '../store/route/index';
import { actions as pipelineActions } from '../store/pipelines/module';
import { PipelineInfo } from '../store/pipelines/index';
import { getters as pipelineGetters } from '../store/pipelines/module';
import { Variable } from '../store/data/index';
import { FilterParams } from '../util/filters';
import Vue from 'vue';

export default Vue.extend({
	name: 'create-pipelines-form',
	data() {
		return {
			descriptionText: '',
			feature: 'Feature',
			featureSet: false,
			metric: 'Metric',
			metricSet: false,
			exportResults: null
		};
	},
	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		variables(): Variable[] {
			return dataGetters.getVariables(this.$store);
		},
		selectedFilters(): FilterParams {
			return {
				filters: dataGetters.getSelectedFilters(this.$store)
			};
		},
		// gets the metrics that are used to score predictions against the user selected variable
		metrics(): string[] {
			// get the variable entry from the store that matches the user selection
			if (!this.target || _.isEmpty(this.variables)) {
				return [];
			}
			// get the task info associated with that variable type
			const taskData = getTask(this.targetVariable.type);
			// grab the valid metrics from the task data to use as labels in the UI
			return getMetricDisplayNames(taskData);
		},
		trainingSelected(): boolean {
			return !_.isEmpty(this.training);
		},
		targetSelected(): boolean {
			return !_.isEmpty(this.target);
		},
		training(): string {
			return routeGetters.getRouteTrainingVariables(this.$store);
		},
		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},
		targetVariable(): Variable {
			return _.find(this.variables, v => {
				return _.toLower(v.name) === _.toLower(this.target);
			});
		},
		sessionId(): string {
			return pipelineGetters.getPipelineSessionID(this.$store);
		},
		// determines create button status based on completeness of user input
		disableCreate(): boolean {
			return !this.targetSelected || !this.trainingSelected;
		},
		// determines  create button variant based on completeness of user input
		createVariant(): string {
			return !this.disableCreate ? 'success' : 'outline-secondary';
		},
		// determines  create button variant based on completeness of user input
		exportVariant(): string {
			return !this.disableCreate ? 'outline-secondary' : 'outline-secondary';
		}
	},
	methods: {
		// create button handler
		create() {
			// compute schema values for request
			const taskData = getTask(this.targetVariable.type);
			const task = taskData.schemaName;
			const metrics = _.map(this.metrics as string[], m => getMetricSchemaName(m));
			// dispatch action that triggers request send to server
			pipelineActions.createPipelines(this.$store, {
				dataset: this.dataset,
				filters: this.selectedFilters,
				sessionId: this.sessionId,
				feature: routeGetters.getRouteTargetVariable(this.$store),
				task: task,
				metric: metrics,
				maxPipelines: 1
			}).then((res: PipelineInfo) => {
				// transition to result screen
				const entry = createRouteEntry(RESULTS_ROUTE, {
					dataset: routeGetters.getRouteDataset(this.$store),
					target: routeGetters.getRouteTargetVariable(this.$store),
					pipelineId: res.pipelineId
				});
				this.$router.push(entry);
			});
		},

		// export button handler
		exportData() {
			dataActions.exportProblem(this.$store, {
				dataset: this.dataset,
				target: this.target,
				filters: this.selectedFilters.filters,
			}).then(res => {
				this.exportResults = res;
			});
		},

		clearExportResults() {
			this.exportResults = null;
		}
	}
});
</script>

<style>
.create-button {
	margin: 0 8px;
	width: 35%;
}
.export-button {
	margin: 0 8px;
	width: 35%;
}
.close-modal {
	width: 50%;
}
.selected-icon {
	padding-right: 4px;
}
.requirement-met {
	padding: 0.5rem;
}
.dropdown-button-style {
	position: relative !important;
	width: 100%;
}
.dropdown-toggle {
	width: 100%;
}
</style>
