<template>
	<div>
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
			<b-btn class="mt-3 join-modal-button" variant="outline-danger" @click="onClose">
				<div class="row justify-content-center">
					<i class="fa fa-times-circle fa-2x mr-2 join-cancel-icon"></i>
					<b>Cancel</b>
				</div>
			</b-btn>
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
	name: 'join-datasets-preview',

	components: {
		JoinDataPreviewSlot
	},

	props: {
		datasetA: String as () => string,
		datasetB: String as () => string,
		previewTableData: Object as () => TableData
	},

	data() {
		return {
			pending: false
		};
	},

	computed: {
		terms(): string {
			return routeGetters.getRouteTerms(this.$store);
		},
		isPending(): boolean {
			return this.pending;
		},
		percentComplete(): number {
			return 100;
		},
		joinedDatasetID(): string {
			return `${this.datasetA}-${this.datasetB}`;
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
		commitJoin() {
			this.pending = true;

			datasetActions.importDataset(this.$store, {
				datasetID: this.joinedDatasetID,
				terms: this.terms,
				source: 'augmented'
			}).then(() => {
				this.$emit('success', this.joinedDatasetID);
				this.pending = false;
			}).catch(() => {
				this.$emit('failure');
				this.pending = false;
			});
		},
		onClose() {
			this.$emit('close');
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
</style>
