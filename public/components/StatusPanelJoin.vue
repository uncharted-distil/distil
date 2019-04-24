<template>
    <div class="status-panel-join">
		<p>Select a dataset to join with</p>
        <div v-for="(item, index) in suggestionItems" :key="item.id">
			<div @click="selectItem(index)" v-bind:class="{ selected: index === selectedIndex, available: item.isAvailable, noavail: item.isAvailable === false }">
				<p>
					{{item.dataset.id}}
				</p>
				<p>
					name: {{item.dataset.name}}
				</p>
				<p>
					{{item.dataset.description}}
				</p>
				<p>
					rows: {{item.dataset.numRows}}
				</p>
				<p>
					size: {{item.dataset.numBytes}}
				</p>
			</div>
        </div>
		<div class="status-message">
			<div v-if="isImporting && importedDataset">
				Importing dataset, <b>{{ importedDataset.name }}</b> ...
			</div>
			<div v-else-if="isImportRequestResolved">
				dataset, <b>{{ importedDataset.name }}</b> imported successfully
			</div>
			<div v-else-if="isImportRequestError">
				Error has occured while importing dataset, <b>{{ importedDataset.name }}</b>
			</div>
		</div>
        <b-button :disabled="!isJoinReady" variant="primary" @click="join">Join</b-button>
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

interface JoinSuggestionItem {
	dataset: Dataset;
	isAvailable: boolean; // tell if dataset is available in the system for join
}

interface StatusPanelJoinState {
	selectedIndex: number;
	suggestionItems: JoinSuggestionItem[];
}

export default Vue.extend({
	name: 'status-panel-join',
	data(): StatusPanelJoinState {
		return {
			selectedIndex: -1,
			suggestionItems: [],
		};
	},
	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
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
			return this.suggestionItems[this.selectedIndex];
		},
		selectedDataset(): Dataset {
			return this.selectedItem && this.selectedItem.dataset;
		},
		isJoinReady(): boolean {
			return this.selectedItem && this.selectedItem.isAvailable !== undefined && !this.isImporting;
		}
	},
	methods: {
		updateSuggestionItems() {
			const items = this.joinSuggestoins || [];
			this.suggestionItems = items.map(suggestion => ({
				dataset: suggestion,
				isAvailable: undefined
			}));
		},
		selectItem(index) {
			if (this.isImporting) { return; }
			this.selectedIndex = index;
			const selected = this.suggestionItems[this.selectedIndex];
			if (selected.isAvailable === undefined) {
				this.checkDatasetExist(selected.dataset.id).then(exist => selected.isAvailable = exist);
			}
		},
		join() {
			const selected = this.suggestionItems[this.selectedIndex];
			if (selected.isAvailable === undefined) { return; }
			if (selected.isAvailable === false) {
				const importAskModal: any = this.$refs['import-ask-modal'];
				importAskModal.show();
			}
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
			if (!this.isImporting) {
				datasetActions.importJoinDataset(this.$store, {datasetID: id, source: 'contrib', provenance, time: 20000}).then(() => {
					this.importedItem.isAvailable = true;
				});
			}
		},
	},
	watch: {
		joinSuggestions(suggestions) {
			this.updateSuggestionItems();
		},
	},
	created() {
		this.updateSuggestionItems();
	},
	beforeDestroy() {
		const importRequest = this.joinDatasetImportRequestData;
		if (importRequest && importRequest.status !== DatasetPendingRequestStatus.PENDING) {
			datasetActions.updatePendingRequestStatus(this.$store, {
				id: importRequest.id,
				status: importRequest.status === DatasetPendingRequestStatus.ERROR
					? DatasetPendingRequestStatus.ERROR_REVIEWED
					: DatasetPendingRequestStatus.REVIEWED,
			});
		}
	},
});

</script>

<style>

.status-panel-join .available {
	background-color: aqua;
}

.status-panel-join .noavail {
	background-color: brown;
}

.status-panel-join .selected {
	background-color: bisque;
}

</style>
