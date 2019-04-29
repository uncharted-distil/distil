<template>
    <div class="status-panel-join">
		<div class="status-message">
			<b-alert v-if="isImporting && importedDataset" :show="showStatusMessage" variant="info">
				Importing <b>{{ importedDataset.name }}</b>...
			</b-alert>
			<b-alert v-else-if="isImportRequestResolved" :show="showStatusMessage" variant="success" dismissible @dismissed="reviewImportingRequest">
				Successfully imported <b>{{ importedDataset.name }}</b>
			</b-alert>
			<b-alert v-else-if="isImportRequestError" :show="showStatusMessage" variant="danger" dismissible  @dismissed="reviewImportingRequest">
				Unexpected error has occured while importing <b>{{ importedDataset.name }}</b>
			</b-alert>
		</div>
		<div class="suggestion-heading">
			<h6>Select a dataset to join with: </h6>
		</div>
		<div class="suggstion-list">
			<div v-if="suggestionItems.length === 0">
				No datasets are found
			</div>
			<b-list-group>
				<b-list-group-item
					v-for="item in suggestionItems"
					:key="item.dataset.id" 
					href="#"
					v-bind:class="{ selected: item.selected }"
					:disabled="isImporting"
				>
					<div @click="selectItem(item)">
						<p> <b>{{item.dataset.name}}</b> </p>
						<p v-html="item.dataset.description">
							{{item.dataset.description}}
						</p>
						<div>
							<span>
								<small v-if="item.isAvailable === false" class="text-info">Requires import</small>
								<small v-if="item.isAvailable" class="text-success">Ready for join</small>
							</span>
							<span class="float-right">
								<small class="text-muted">{{formatNumber(item.dataset.numRows)}} rows</small>
								<small class="text-muted">{{formatBytes(item.dataset.numBytes)}}</small>
							</span>
						</div>
					</div>
				</b-list-group-item>
			</b-list-group>
		</div>
		<div class="join-button-container">
        	<b-button :disabled="!isJoinReady" variant="primary" @click="join">Join</b-button>
		</div>
		<b-modal
			v-if="selectedDataset"
			id="join-import-modal"
			ref="import-ask-modal"
			title="JoinSuggestionImport"
			@ok="importDataset"
		>
			<p class="">Dataset, <b>{{ selectedDataset.name }}</b> is not available in the system. Would you like to import the dataset?</p>
		</b-modal>
    </div>
</template>

<script lang="ts">

import Vue from 'vue';
import axios from 'axios';
import {
	Dataset,
	DatasetPendingRequestType,
	DatasetPendingRequestStatus,
	JoinSuggestionPendingRequest,
	JoinDatasetImportPendingRequest,
} from '../store/dataset/index';
import { actions as datasetActions, getters as datasetGetters } from '../store/dataset/module';
import { actions as appActions, getters as appGetters } from '../store/app/module';
import { getters as routeGetters } from '../store/route/module';
import { StatusPanelState, StatusPanelContentType } from '../store/app';
import { createRouteEntry } from '../util/routes'
import { formatBytes } from '../util/bytes';
import { JOIN_DATASETS_ROUTE } from '../store/route/index';

interface JoinSuggestionItem {
	dataset: Dataset;
	isAvailable: boolean; // tell if dataset is available in the system for join. (note. undefined implies that check hasn't made yet)
	selected: boolean;
}

interface StatusPanelJoinState {
	suggestionItems: JoinSuggestionItem[];
	showStatusMessage: boolean;
}

export default Vue.extend({
	name: 'status-panel-join',
	data(): StatusPanelJoinState {
		return {
			showStatusMessage: true,
			suggestionItems: [],
		};
	},
	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},
		joinSuggestionRequestData(): JoinSuggestionPendingRequest {
			const request = datasetGetters.getPendingRequests(this.$store)
				.find(request => request.dataset === this.dataset && request.type === DatasetPendingRequestType.JOIN_SUGGESTION);
			return <JoinSuggestionPendingRequest>request;
		},
		joinSuggestoins(): Dataset[] {
			const joinSuggestions = this.joinSuggestionRequestData && this.joinSuggestionRequestData.suggestions;
			return joinSuggestions || [];
		},
		joinDatasetImportRequestData(): JoinDatasetImportPendingRequest {
			// get importing request for a dataset that is in the suggestion list.
			const request = datasetGetters.getPendingRequests(this.$store)
				.find(request => request.type === DatasetPendingRequestType.JOIN_DATASET_IMPORT);
			const isInSuggestionList = Boolean(this.joinSuggestoins.find(item => item.id === (request && request.dataset)));
			return isInSuggestionList ? <JoinDatasetImportPendingRequest>request : undefined;
		},
		isImporting(): boolean {
			const requestStatus = this.joinDatasetImportRequestData && this.joinDatasetImportRequestData.status;
			return requestStatus === DatasetPendingRequestStatus.PENDING;
		},
		importedItem(): JoinSuggestionItem {
			return this.suggestionItems.find(item => item.dataset.id === this.joinDatasetImportRequestData.dataset);
		},
		importedDataset(): Dataset {
			return this.importedItem && this.importedItem.dataset;
		},
		isImportRequestResolved(): boolean {
			return this.joinDatasetImportRequestData && (this.joinDatasetImportRequestData.status === DatasetPendingRequestStatus.RESOLVED);
		},
		isImportRequestError(): boolean {
			return this.joinDatasetImportRequestData && (this.joinDatasetImportRequestData.status === DatasetPendingRequestStatus.ERROR);
		},
		selectedItem(): JoinSuggestionItem {
			return this.suggestionItems.find(item => item.selected);
		},
		selectedDataset(): Dataset {
			return this.selectedItem && this.selectedItem.dataset;
		},
		isJoinReady(): boolean {
			return this.selectedItem && this.selectedItem.isAvailable !== undefined && !this.isImporting;
		}
	},
	methods: {
		initSuggestionItems() {
			const items = this.joinSuggestoins || [];
			const isImporting = this.isImporting || this.isImportRequestResolved;
			this.suggestionItems = items.map(suggestion => {
				const isSameDataset = suggestion.id === (this.joinDatasetImportRequestData && this.joinDatasetImportRequestData.dataset);
				const isAvailable = this.isImportRequestResolved && isSameDataset ? true : undefined;
				const selected = isImporting && isSameDataset;
				return {
					dataset: suggestion,
					isAvailable,
					selected,
				};
			});
		},
		selectItem(item) {
			if (this.isImporting) { return; }
			if (this.selectedItem) {
				this.selectedItem.selected = false;
			}
			const selectedItem = item;
			selectedItem.selected = true;
			if (selectedItem.isAvailable === undefined) {
				this.checkDatasetExist(selectedItem.dataset.id).then(exist => selectedItem.isAvailable = exist);
			}
		},
		join() {
			const selected = this.selectedItem;
			if (selected.isAvailable === undefined) { return; }
			if (selected.isAvailable === false) {
				const importAskModal: any = this.$refs['import-ask-modal'];
				return importAskModal.show();
			}
			// navigate to join
			const entry = createRouteEntry(JOIN_DATASETS_ROUTE, {
				joinDatasets: `${this.dataset},${selected.dataset.id}`,
				target: this.target, 
			});
			this.$router.push(entry);
		},
		checkDatasetExist(datasetId) {
			return axios.get(`/distil/datasets/${datasetId}`).then(result => {
				return result ? true : false;
			}).catch(e => {
				return false;
			});
		},
		importDataset(args: {datasetID: string, source: string, provenance: string}) {
			const { id, provenance } = this.selectedDataset;
			this.showStatusMessage = true;
			if (!this.isImporting) {
				datasetActions.importJoinDataset(this.$store, {datasetID: id, source: 'contrib', provenance}).then(res => {
					if (res && (res.result === 'ingested')) {
						this.importedItem.isAvailable = true;
					}
				});
			}
		},
		formatBytes(n: number): string {
			return formatBytes(n);
		},
		formatNumber(num: number): string {
			if (num >= 1000000000) {
				return (num / 1000000000).toFixed(1).replace(/\.0$/, '') + 'B';
			}
			if (num >= 1000000) {
				return (num / 1000000).toFixed(1).replace(/\.0$/, '') + 'M';
			}
			if (num >= 1000) {
				return (num / 1000).toFixed(1).replace(/\.0$/, '') + 'K';
			}
			return String(num);
		},
		reviewImportingRequest() {
			const importRequest = this.joinDatasetImportRequestData;
			if (importRequest && importRequest.status !== DatasetPendingRequestStatus.PENDING) {
				datasetActions.updatePendingRequestStatus(this.$store, {
					id: importRequest.id,
					status: importRequest.status === DatasetPendingRequestStatus.ERROR
						? DatasetPendingRequestStatus.ERROR_REVIEWED
						: DatasetPendingRequestStatus.REVIEWED,
				});
			}
		}
	},
	created() {
		this.initSuggestionItems();
	},
	beforeDestroy() {
		this.reviewImportingRequest();
	},
});

</script>

<style>

.status-panel-join {
	height: 100%;
	display: flex;
	flex-direction: column;
}
.status-panel-join .suggestion-heading {
	height: 2em;
	flex-shrink: 0;
}
.status-panel-join .suggestion-heading h6{
	margin: 0;
}
.status-panel-join .suggstion-list {
	overflow: auto;
}
.status-panel-join .list-group-item.selected{
	background-color: #00c5e114
}
.status-panel-join .status-message {
	min-height: 0;
	flex-shrink: 0;
	margin-top: 5px;
}
.status-panel-join .join-button-container {
	min-height: 0;
	padding: 5px 0;
	flex-shrink: 0;
}
.status-panel-join .join-button-container button {
	width: 100%;
}

</style>
