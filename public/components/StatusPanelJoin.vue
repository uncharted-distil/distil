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
        <b-button :disabled="!selectedItem || !selectedItem.isAvailable" variant="primary" @click="join">Join</b-button>
    </div>
</template>

<script lang="ts">

import Vue from 'vue';
import axios from 'axios';
import {
	Dataset,
	DatasetPendingRequestType,
	JoinSuggestionPendingRequest
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
			const joinSuggestions = this.joinSuggestionRequestData.suggestions;
			return joinSuggestions;
		},
		selectedItem(): JoinSuggestionItem {
			return this.suggestionItems[this.selectedIndex];
		},
	},
	methods: {
		updateSuggestionItems() {
			const items = this.joinSuggestoins || [];
			this.suggestionItems = items.map(suggestion => ({
				dataset: suggestion,
				isAvailable: undefined
			}));
			console.log(this.suggestionItems);
		},
		selectItem(index) {
			this.selectedIndex = index;
			const selected = this.suggestionItems[this.selectedIndex];
			if (selected.isAvailable === undefined) {
				this.checkDatasetExist(selected.dataset.id).then(exist => selected.isAvailable = exist);
			}
			setTimeout(() => {
				selected.isAvailable = true;
			}, 4000);
		},
		join() {
			const selected = this.suggestionItems[this.selectedIndex];
		},
		checkDatasetExist(datasetId) {
			return axios.get(`/distil/datasets/${datasetId}`).then(result => {
				return result ? true : false;
			}).catch(e => {
				return false;
			});
		},
		importDataset(args: {datasetID: string, source: string, provenance: string}) {
			// return axios.post(`/distil/import/${args.datasetID}/${args.source}/${args.provenance}`, {})
		}
	},
	watch: {
		joinSuggestions(suggestions) {
			this.updateSuggestionItems();
		},
	},
	created() {
		this.updateSuggestionItems();
	},
});

</script>

<style>
.status-panel-join .selected {
	background-color: bisque;
}

.status-panel-join .available {
	background-color: aqua;
}

.status-panel-join .noavail {
	background-color: brown;
}

</style>
