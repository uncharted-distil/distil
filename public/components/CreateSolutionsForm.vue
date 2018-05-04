<template>
	<div class="create-solutions-form">
		<b-modal id="export-modal" title="Export Succeeded"
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
			<b-button class="export-button" :variant="exportVariant" @click="exportData" :disabled="disableExport">
				Task 1: Export Problem
			</b-button>
			<b-button class="create-button" :variant="createVariant" @click="create" :disabled="disableCreate">
				Create Models
			</b-button>
		</div>
		<div class="solution-progress">
			<b-progress v-if="isPending"
				:value="percentComplete"
				variant="outline-secondary"
				striped
				:animated="true"></b-progress>
		</div>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import { createRouteEntry } from '../util/routes';
import { getTask, getMetricDisplayNames, getMetricSchemaName } from '../util/solutions';
import { getters as dataGetters, actions as dataActions } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { RESULTS_ROUTE } from '../store/route/index';
import { actions as solutionActions } from '../store/solutions/module';
import { SolutionInfo } from '../store/solutions/index';
import { Variable } from '../store/data/index';
import { FilterParams } from '../util/filters';
import Vue from 'vue';

export default Vue.extend({
	name: 'create-solutions-form',
	data() {
		return {
			descriptionText: '',
			feature: 'Feature',
			featureSet: false,
			metric: 'Metric',
			metricSet: false,
			exportResults: null,
			pending: false
		};
	},
	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		variables(): Variable[] {
			return dataGetters.getVariables(this.$store);
		},
		filters(): FilterParams {
			return dataGetters.getSelectedFilterParams(this.$store);
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
		isPending(): boolean {
			return this.pending;
		},
		disableCreate(): boolean {
			return this.isPending || (!this.targetSelected || !this.trainingSelected);
		},
		disableExport(): boolean {
			return !this.targetSelected || !this.trainingSelected;
		},
		createVariant(): string {
			return !this.disableCreate ? 'success' : 'outline-secondary';
		},
		exportVariant(): string {
			return !this.disableExport ? 'primary' : 'outline-secondary';
		},
		percentComplete(): number {
			return 100;
		}
	},
	methods: {
		// create button handler
		create() {
			// compute schema values for request
			const taskData = getTask(this.targetVariable.type);
			const task = taskData.schemaName;
			const metrics = _.map(this.metrics as string[], m => getMetricSchemaName(m));
			this.pending = true;
			// dispatch action that triggers request send to server
			solutionActions.createSolutions(this.$store, {
				dataset: this.dataset,
				filters: this.filters,
				target: routeGetters.getRouteTargetVariable(this.$store),
				task: task,
				metrics: metrics,
				maxSolutions: 1
			}).then((res: SolutionInfo) => {
				this.pending = false;
				// transition to result screen
				const entry = createRouteEntry(RESULTS_ROUTE, {
					dataset: routeGetters.getRouteDataset(this.$store),
					target: routeGetters.getRouteTargetVariable(this.$store),
					solutionId: res.solutionId
				});
				this.$router.push(entry);
			});
		},

		// export button handler
		exportData() {
			dataActions.exportProblem(this.$store, {
				dataset: this.dataset,
				target: this.target,
				filters: this.filters,
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
.solution-progress {
	margin: 6px 10%;
}
</style>
