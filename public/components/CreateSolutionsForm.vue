<template>
	<div class="create-solutions-form">
		<b-modal
			v-model="showExportSuccess"
			cancel-disabled
			hide-header
			hide-footer>
			<div class="row justify-content-center">
				<div class="check-message-container">
					<i class="fa fa-check-circle fa-3x check-icon"></i>
					<div><b>Export Succeded</b></div>
				</div>
			</div>
			<div class="row justify-content-center">
				<b-btn class="mt-3 close-modal" block @click="showExportSuccess = !showExportSuccess">OK</b-btn>
			</div>
		</b-modal>


		<error-modal
			:show="showExportFailure"
			title="Export Failed"
			@close="showExportFailure = !showExportFailure">
		</error-modal>

		<b-modal
			v-model="showExport"
			cancel-disabled
			hide-header
			hide-footer>
			<div class="row justify-content-center">
				<b-radio-group v-model="meaningful">
					<div class="meaningful-text">Is this a meaningful problem?</div>
					<b-radio value=true>Yes</b-radio>
					<b-radio value=false>No</b-radio>
				</b-radio-group>
			</div>
			<div class="row justify-content-center">
				<b-btn class="mt-3 close-modal" variant="success" block @click="exportData">Export</b-btn>
			</div>
		</b-modal>

		<error-modal
			:show="showCreateFailure"
			title="Model Failed"
			:error="createErrorMessage"
			@close="showCreateFailure = !showCreateFailure">
		</error-modal>

		<div class="row justify-content-center">
			<b-button class="export-button" :variant="exportVariant" @click="showExport = !showExport" v-if="isTask1">
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
import ErrorModal from '../components/ErrorModal.vue';
import { actions as appActions, getters as appGetters } from '../store/app/module';
import { getters as datasetGetters } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';
import { RESULTS_ROUTE } from '../store/route/index';
import { actions as solutionActions } from '../store/solutions/module';
import { Solution, NUM_SOLUTIONS } from '../store/solutions/index';
import { Variable } from '../store/dataset/index';
import { FilterParams } from '../util/filters';
import Vue from 'vue';

export default Vue.extend({
	name: 'create-solutions-form',

	components: {
		ErrorModal
	},

	data() {
		return {
			pending: false,
			meaningful: true,
			showExport: false,
			showExportSuccess: false,
			showExportFailure: false,
			showCreateFailure: false,
			createErrorMessage: null
		};
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		variables(): Variable[] {
			return datasetGetters.getVariables(this.$store);
		},
		filterParams(): FilterParams {
			return routeGetters.getDecodedFilterParams(this.$store);
		},
		metrics(): string[] {
			if (this.isTask2) {
				return appGetters.getProblemMetrics(this.$store);
			}
			return null;
		},
		taskType(): string {
			if (this.isTask2) {
				return appGetters.getProblemTaskType(this.$store);
			}
			return null;
		},
		taskSubType(): string {
			if (this.isTask2) {
				return appGetters.getProblemTaskSubType(this.$store);
			}
			return null;
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
				return _.toLower(v.colName) === _.toLower(this.target);
			});
		},
		isPending(): boolean {
			return this.pending;
		},
		isTask1(): boolean {
			return appGetters.isTask1(this.$store);
		},
		isTask2(): boolean {
			return appGetters.isTask2(this.$store);
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
			return !this.disableExport ? 'dark' : 'outline-secondary';
		},
		percentComplete(): number {
			return 100;
		}
	},
	methods: {
		// create button handler
		create() {
			// flag as pending
			this.pending = true;
			// dispatch action that triggers request send to server
			solutionActions.createSolutionRequest(this.$store, {
				dataset: this.dataset,
				filters: this.filterParams,
				target: routeGetters.getRouteTargetVariable(this.$store),
				task: this.taskType,
				subTask: this.taskSubType,
				metrics: this.metrics,
				maxSolutions: NUM_SOLUTIONS,
				// intentionally nulled for now - should be made user settable in the future
				maxTime: null,
			}).then((res: Solution) => {
				this.pending = false;
				// transition to result screen
				const entry = createRouteEntry(RESULTS_ROUTE, {
					dataset: routeGetters.getRouteDataset(this.$store),
					target: routeGetters.getRouteTargetVariable(this.$store),
					solutionId: res.solutionId
				});
				this.$router.push(entry);
			}).catch(err => {
				// display error modal
				this.pending = false;
				this.createErrorMessage = err.message;
				this.showCreateFailure = true;
			});
		},

		exportData() {
			appActions.exportProblem(this.$store, {
				dataset: this.dataset,
				target: this.target,
				filterParams: this.filterParams,
				meaningful: this.meaningful ? 'Yes' : 'No'
			}).then(res => {
				this.showExportSuccess = !this.showExportSuccess;
				this.meaningful = true;
			}).catch(err => {
				this.showExportFailure = !this.showExportFailure;
				this.meaningful = true;
			});
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
	width: 35% !important;
}

.solution-progress {
	margin: 6px 10%;
}

.check-message-container {
	display: flex;
	justify-content: flex-start;
	flex-direction: row;
	align-items: center;
}

.check-icon {
	display: flex;
	flex-shrink: 0;
	color:#00C851;
	padding-right: 15px;
}

.fail-icon {
	display: flex;
	flex-shrink: 0;
	color: #ee0701;
	padding-right: 15px;
}
</style>
