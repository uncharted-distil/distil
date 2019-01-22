<template>
	<div class="create-solutions-form">
		<b-modal title="Join Preview"
			class="join-preview-modal"
			v-model="showJoinSuccess"
			cancel-disabled
			hide-header
			hide-footer>

			<div class="row justify-content-center">
				<div class="check-message-container">
					<h5 class="mt-4 mb-4"><b>Join Preview</b></h5>
				</div>
			</div>

			<join-data-preview-slot
				:items="joinDataPreviewItems"
				:fields="joinDataPreviewFields"
				:numRows="joinDataPreviewNumRows"
				:hasData="joinDataPreviewHasData"
				instance-name="join-dataset-bottom"></join-data-preview-slot>

			<div class="row justify-content-center">
				<b-btn class="mt-3 join-modal-button" variant="outline-success" @click="commitJoin">
					<div class="row justify-content-center">
						<i class="fa fa-check-circle fa-2x mr-2 join-confirm-icon"></i>
						<b>Commit join</b>
					</div>
				</b-btn>
				<b-btn class="mt-3 join-modal-button" variant="outline-danger" @click="showJoinSuccess = !showJoinSuccess">
					<div class="row justify-content-center">
						<i class="fa fa-times-circle fa-2x mr-2 join-cancel-icon"></i>
						<b>Cancel</b>
					</div>
				</b-btn>
			</div>

		</b-modal>
		<b-modal title="Join Failed"
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
				<b-btn class="mt-3 join-modal-button" block @click="showJoinFailure = !showJoinFailure">OK</b-btn>
			</div>
		</b-modal>
		<div v-if="columnTypesDoNotMatch" class="row justify-content-center mt-3 mb-3 warning-text">
			<i class="fa fa-exclamation-triangle warning-icon mr-2"></i>
			<span v-html="joinWarning"></span>
		</div>
		<div class="row justify-content-center">
			<b-button class="join-button" :variant="joinVariant" @click="previewJoin" :disabled="disableJoin">
				<div class="row justify-content-center">
					<i class="fa fa-check-circle fa-2x mr-2"></i>
					<b>Join Datasets</b>
				</div>
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
import Vue from 'vue';
import JoinDataPreviewSlot from '../components/JoinDataPreviewSlot.vue';
import { createRouteEntry } from '../util/routes';
import { Dictionary } from '../util/dict';
import { getters as routeGetters } from '../store/route/module';
import { Dataset, TableData, TableColumn, TableRow } from '../store/dataset/index';
import { getters as datasetGetters, actions as datasetActions } from '../store/dataset/module';
import { getTableDataItems, getTableDataFields } from '../util/data';

export default Vue.extend({
	name: 'join-datasets-form',

	components: {
		JoinDataPreviewSlot
	},

	props: {
		datasetA: String as () => string,
		datasetB: String as () => string,
		datasetAColumn: Object as () => TableColumn,
		datasetBColumn: Object as () => TableColumn
	},

	data() {
		return {
			pending: false,
			showJoin: false,
			showJoinSuccess: false,
			showJoinFailure: false,
			joinErrorMessage: null,
			previewTableData: null
		};
	},

	computed: {
		datasets(): Dataset[] {
			return datasetGetters.getDatasets(this.$store);
		},
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
		},
		joinDataPreviewItems(): TableRow[] {
			return getTableDataItems(this.previewTableData);
		},
		joinDataPreviewFields(): Dictionary<TableColumn> {
			return getTableDataFields(this.previewTableData);
		},
		joinDataPreviewNumRows(): number {
			return this.previewTableData ? this.previewTableData.numRows : 0;
		},
		joinDataPreviewHasData(): boolean {
			return !!this.previewTableData;
		}
	},

	methods: {
		previewJoin() {
			// flag as pending
			this.pending = true;

			const a = _.find(this.datasets, d => {
				return d.id === this.datasetA;
			});

			const b = _.find(this.datasets, d => {
				return d.id === this.datasetB;
			});

			// dispatch action that triggers request send to server
			datasetActions.joinDatasetsPreview(this.$store, {
				datasetA: a,
				datasetB: b,
				datasetAColumn: this.datasetAColumn.key,
				datasetBColumn: this.datasetBColumn.key
			}).then(tableData => {
				this.pending = false;
				this.showJoinSuccess = true;
				this.previewTableData = tableData;
			}).catch(err => {
				// display error modal
				this.pending = false;
				this.showJoinFailure = true;
				this.joinErrorMessage = err.message;
				this.previewTableData = null;
			});
		},
		commitJoin() {
			console.log('commit join');
		}
	}
});
</script>

<style>
.join-button {
	margin: 0 8px;
	width: 35%;
	line-height: 32px !important;
}

.join-modal-button {
	margin: 0 8px;
	width: 25% !important;
	line-height: 32px !important;
}

.join-preview-modal .modal-dialog {
	position: relative;
	max-width: 80% !important;
	max-height: 90%;
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

.join-confirm-icon {
	color: #00C851;
}

.join-cancel-icon {
	color: #ee0701;
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
