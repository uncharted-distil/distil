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
		<div class="suggestion-search">
			<b-input v-model="filterString" placeholder="Search Join Suggestions" />
		</div>
		<div class="suggestion-heading">
			<h6>Select a dataset to join with: </h6>
		</div>
		<div class="suggestion-list">
			<div v-if="filteredSuggestedItems.length === 0">
				No datasets are found
			</div>
			<b-list-group>
				<b-list-group-item
					v-for="item in filteredSuggestedItems"
					:key="item.key"
					href="#"
					v-bind:class="{ selected: item.selected }"
					:disabled="isImporting"
					@click="selectItem(item)"
				>
					<p> <b>{{item.dataset.name}}</b> </p>
					<div class="description" v-html="item.dataset.description">
						{{item.dataset.description}}
					</div>
					<div v-if="item.dataset.joinSuggestion" class="suggested-columns">
						<b>Suggested Join Columns: </b>{{item.dataset.joinSuggestion[0].joinColumns}}
					</div>
					<div>
						<span>
							<small v-if="!item.isAvailable" class="text-info">Requires import</small>
							<small v-if="item.isAvailable" class="text-success">Ready for join</small>
						</span>
						<span class="float-right">
							<small class="text-muted">{{formatNumber(item.dataset.numRows)}} rows</small>
							<small class="text-muted">{{formatBytes(item.dataset.numBytes)}}</small>
						</span>
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
import { createRouteEntry } from '../util/routes';
import { formatBytes } from '../util/bytes';
import { isDatamartProvenance } from '../util/data';
import { JOIN_DATASETS_ROUTE } from '../store/route/index';

interface JoinSuggestionItem {
	dataset: Dataset;
	key: string;
	isAvailable: boolean; // tell if dataset is available in the system for join. (note. undefined implies that check hasn't made yet)
	selected: boolean;
}

interface StatusPanelJoinState {
	suggestionItems: JoinSuggestionItem[];
	showStatusMessage: boolean;
	filterString: string;
}

export default Vue.extend({
	name: 'status-panel-join',
	data(): StatusPanelJoinState {
		return {
			showStatusMessage: true,
			suggestionItems: [],
			filterString: ''
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
		joinSuggestions(): Dataset[] {
			const joinSuggestions = this.joinSuggestionRequestData && this.joinSuggestionRequestData.suggestions;
			return joinSuggestions || [];
		},
		filteredSuggestedItems(): JoinSuggestionItem[] {
			const filteredItems = this.filterString.length > 0 && this.suggestionItems.length > 0
				? this.suggestionItems.filter(item => (
					item.dataset.name.indexOf(this.filterString) > -1
					|| item.dataset.description.indexOf(this.filterString) > -1
				))
				: this.suggestionItems;
			return filteredItems;
		},
		joinDatasetImportRequestData(): JoinDatasetImportPendingRequest {
			// get importing request for a dataset that is in the suggestion list.
			const request = datasetGetters.getPendingRequests(this.$store)
				.find(request => request.type === DatasetPendingRequestType.JOIN_DATASET_IMPORT);
			const isInSuggestionList = Boolean(this.joinSuggestions.find(item => item.id === (request && request.dataset)));
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
		},
		baseColumnSuggestions(): string[] {
			return this.selectedDataset && this.selectedDataset.joinSuggestion && this.selectedDataset.joinSuggestion[0]
				? this.selectedDataset.joinSuggestion[0].baseColumns
				: [];
		},
		joinColumnSuggestions(): string[] {
			return this.selectedDataset && this.selectedDataset.joinSuggestion && this.selectedDataset.joinSuggestion[0]
				? this.selectedDataset.joinSuggestion[0].joinColumns
				: [];
		},
	},
	methods: {
		initSuggestionItems() {
			const items = this.joinSuggestions || [];
			// resolve join availablity of the importing dataset
			const isImporting = this.isImporting || this.isImportRequestResolved;
			this.suggestionItems = items.map(suggestion => {
				const isImportingDataset = suggestion.id === (this.joinDatasetImportRequestData && this.joinDatasetImportRequestData.dataset);
				const isAvailable = isImportingDataset
					? this.isImportRequestResolved
					: !isDatamartProvenance(suggestion.provenance);
				const selected = isImporting && isImportingDataset;
				return {
					dataset: suggestion,
					// There could be multiple items with same dataset id with different join suggestions.
					// So item key must be a combination of id and the join suggestions to be unique
					key: suggestion.id
						+ (suggestion.joinSuggestion && suggestion.joinSuggestion[0] ? `${
							suggestion.joinSuggestion[0].baseColumns
							.concat(suggestion.joinSuggestion[0].joinColumns)
							.join('-')
						}` : ''),
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
		},
		join() {
			const selected = this.selectedItem;
			if (selected.isAvailable === false) {
				const importAskModal: any = this.$refs['import-ask-modal'];
				return importAskModal.show();
			}
			const replaceComma = str => str.replace(/, /g, '+');
			// navigate to join
			const entry = createRouteEntry(JOIN_DATASETS_ROUTE, {
				joinDatasets: `${this.dataset},${selected.dataset.id}`,
				target: this.target,
				baseColumnSuggestions: this.baseColumnSuggestions.map(replaceComma).join(','),
				joinColumnSuggestions: this.joinColumnSuggestions.map(replaceComma).join(','),
			});
			this.$router.push(entry);
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
.status-panel-join .suggestion-list {
	overflow: auto;
	overflow-wrap: break-word;
}

.status-panel-join .suggestion-list .suggested-columns {
	font-size: .75rem;
}

.status-panel-join .suggestion-search {
	height: 2em;
	margin-bottom: 20px;
	flex-shrink: 0;
}

.status-panel-join .list-group-item.selected{
	background-color: #00c5e114
}

.status-panel-join .list-group-item .description a:hover{
	color: #007bff;
	text-decoration: inherit;
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
