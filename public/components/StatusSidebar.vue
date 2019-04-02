<template>
<div class="status-sidebar">
    <div class="status-icons">
        <div class="status-icon-wrapper" @click="onStatusIconClick(0)">
            <i class="status-icon fa fa-2x fa-info" aria-hidden="true"></i>
			<i v-if="variableRankingStatus === 'resolved'" class="new-update-notification fa fa-refresh fa-circle"></i>
			<i v-if="variableRankingStatus === 'pending'" class="new-update-notification fa fa-refresh fa-spin"></i>
        </div>
        <div class="status-icon-wrapper" @click="onStatusIconClick(1)">
            <i class="status-icon fa fa-2x fa-location-arrow" aria-hidden="true"></i>
			<i v-if="geocodingStatus === 'resolved'" class="new-update-notification fa fa-circle"></i>
			<i v-if="geocodingStatus === 'pending'" class="new-update-notification fa-refresh fa-spin"></i>
        </div>
        <div class="status-icon-wrapper" @click="onStatusIconClick(2)">
            <i class="status-icon fa fa-2x fa-long-arrow-right" aria-hidden="true"></i>
			<i v-if="joinSuggestionStatus === 'resolved'" class="new-update-notification fa fa-circle"></i>
			<i v-if="joinSuggestionStatus === 'pending'" class="new-update-notification fa-refresh fa-spin"></i>
        </div>
    </div>
</div>
    
</template>

<script lang="ts">

import Vue from 'vue';
import { actions as datasetActions, getters as datasetGetters } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';
import { DatasetPendingUpdateType, DatasetPendingUpdate, VariableRankingPendingUpdate } from '../store/dataset/index';

const STATUS_TYPES = [
	DatasetPendingUpdateType.VARIABLE_RANKING,
	DatasetPendingUpdateType.GEOCODING,
	DatasetPendingUpdateType.JOIN_SUGGESTION,
];

export default Vue.extend({
	name: 'status-sidebar',
	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		pendingUpdates: function () {
			const updates = datasetGetters.getPendingUpdates(this.$store).filter(update => update.dataset === this.dataset);
			return updates;
		},
		variableRankingUpdate: function () {
			return this.pendingUpdates.find(item =>  item.type === DatasetPendingUpdateType.VARIABLE_RANKING);
		},
		geocodingUpdate: function () {
			return this.pendingUpdates.find(item => item.type === DatasetPendingUpdateType.GEOCODING);
		},
		joinSuggestionUpdate: function () {
			return this.pendingUpdates.find(item => item.type === DatasetPendingUpdateType.JOIN_SUGGESTION);
		},
		variableRankingStatus: function () {
			return this.variableRankingUpdate && this.variableRankingUpdate.status;
		},
		geocodingStatus: function () {
			return this.geocodingUpdate && this.geocodingUpdate.status;
		},
		joinSuggestionStatus: function () {
			return this.joinSuggestionUpdate && this.joinSuggestionUpdate.status;
		},
	},
	mounted() {
		console.log('mounted with this dataset', this.dataset);
	},
	methods: {
		onStatusIconClick(iconIndex) {
			const update = this.pendingUpdates.find(item => item.type === STATUS_TYPES[iconIndex]);
			if (update) {
				datasetActions.updatePendingUpdateStatus(this.$store, {
					id: update.id,
					status: update.status === 'pending' ? update.status : 'reviewed',
				});
			}
			this.$emit('statusIconClick', STATUS_TYPES[iconIndex]);

		},
	},
});

</script>

<style>

.status-sidebar {
	background-color: #fff;
	width: 55px;
	border-left: 1px solid #ccc;
	height: 100%;
	display: flex;
	flex-direction: column;
}
.status-sidebar .status-icon-wrapper {
	padding-top: 15px;
	padding-bottom: 15px;
	text-align: center;
	position: relative;
}
.status-sidebar .status-icon {
	height: 30px;
	width: 30px;
	cursor: pointer;
}
.status-sidebar .new-update-notification {
	position: absolute;
    color: #dc3545;
    top: 10px;
    right: 10px;
}

</style>
