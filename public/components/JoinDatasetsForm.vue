<template>
	<div class="create-solutions-form">
		<b-modal id="export-success-modal" title="Join Succeeded"
			v-model="showJoinSuccess"
			cancel-disabled
			hide-header
			hide-footer>
			<div class="row justify-content-center">
				<div class="check-message-container">
					<i class="fa fa-check-circle fa-3x check-icon"></i>
					<div><b>Join Succeded</b></div>
				</div>
			</div>
			<div class="row justify-content-center">
				<b-btn class="mt-3 close-modal" block @click="showJoinSuccess = !showJoinSuccess">OK</b-btn>
			</div>
		</b-modal>
		<b-modal id="export-failure-modal" title="Join Failed"
			v-model="showJoinFailure"
			cancel-disabled
			hide-header
			hide-footer>
			<div class="row justify-content-center">
				<div class="check-message-container">
					<i class="fa fa-exclamation-triangle fa-3x fail-icon"></i>
					<div><b>Join Failed:</b> Internal server error</div>
				</div>
			</div>
			<div class="row justify-content-center">
				<b-btn class="mt-3 close-modal" block @click="showJoinFailure = !showJoinFailure">OK</b-btn>
			</div>
		</b-modal>
		<div v-if="columnTypesDoNotMatch" class="row justify-content-center mt-3 mb-3 warning-text">
			<i class="fa fa-exclamation-triangle warning-icon mr-2"></i>
			<span v-html="joinWarning"></span>
		</div>
		<div class="row justify-content-center">
			<b-button class="join-button" :variant="joinVariant" @click="join" :disabled="disableJoin">
				Join Datasets
			</b-button>
		</div>
		<div class="join-progress">
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
import { getters as routeGetters } from '../store/route/module';
import { TableColumn } from '../store/dataset/index';
import Vue from 'vue';

export default Vue.extend({
	name: 'join-datasets-form',

	props: {
		datasetAColumn: Object as () => TableColumn,
		datasetBColumn: Object as () => TableColumn
	},

	data() {
		return {
			pending: false,
			showJoin: false,
			showJoinSuccess: false,
			showJoinFailure: false,
			createErrorMessage: null
		};
	},

	computed: {
		columnsSelected(): boolean {
			return !!this.datasetAColumn && !!this.datasetBColumn;
		},
		columnTypesDoNotMatch(): boolean {
			return this.datasetAColumn && this.datasetBColumn &&
			this.datasetAColumn.type !== this.datasetBColumn.type;
		},
		isPending(): boolean {
			return this.pending;
		},
		joinWarning(): string {
			if (this.columnTypesDoNotMatch) {
				return `Unable to join column <b>${this.datasetAColumn.key}</b> of type <b>${this.datasetAColumn.type}</b> with <b>${this.datasetAColumn.key}</b> of type <b>${this.datasetAColumn.type}</b>`;
			}
		},
		disableJoin(): boolean {
			return this.isPending || !this.columnsSelected || this.columnTypesDoNotMatch;
		},
		joinVariant(): string {
			return !this.disableJoin ? 'success' : 'outline-secondary';
		},
		percentComplete(): number {
			return 100;
		}
	},

	methods: {
		join() {
			// flag as pending
			this.pending = true;
			// // dispatch action that triggers request send to server
			// solutionActions.createSolutionRequest(this.$store, {
			// 	dataset: this.dataset,
			// 	filters: this.filterParams,
			// 	target: routeGetters.getRouteTargetVariable(this.$store),
			// 	task: this.taskType,
			// 	subTask: this.taskSubType,
			// 	metrics: this.metrics,
			// 	maxSolutions: NUM_SOLUTIONS,
			// 	// intentionally nulled for now - should be made user settable in the future
			// 	maxTime: null,
			// }).then((res: Solution) => {
			// 	this.pending = false;
			// 	// transition to result screen
			// 	const entry = createRouteEntry(RESULTS_ROUTE, {
			// 		dataset: routeGetters.getRouteDataset(this.$store),
			// 		target: routeGetters.getRouteTargetVariable(this.$store),
			// 		solutionId: res.solutionId
			// 	});
			// 	this.$router.push(entry);
			// }).catch(err => {
			// 	// display error modal
			// 	this.pending = false;
			// 	this.createErrorMessage = err.message;
			// 	this.showCreateFailure = true;
			// });
		}
	}
});
</script>

<style>
.join-button {
	margin: 0 8px;
	width: 35%;
}

.close-modal {
	width: 35%;
}

.join-progress {
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

.warning-icon {
	color: #ee0701;
}

.warning-text {
	line-height: 16px;
	font-size: 16px;
}

.check-button {
	width: 60%;
	margin: 0 20%;
}
</style>
