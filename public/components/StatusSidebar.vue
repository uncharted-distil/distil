<template>
<div class="status-sidebar">
    <div class="status-icons">
        <div class="status-icon-wrapper" @click="onStatusIconClick(0)">
            <i class="status-icon fa fa-2x fa-info" aria-hidden="true"></i>
			<i v-if="isNew(variableRankingStatus)" class="new-update-notification fa fa-refresh fa-circle"></i>
			<i v-if="isPending(variableRankingStatus)" class="new-update-notification fa fa-refresh fa-spin"></i>
        </div>
        <div class="status-icon-wrapper" @click="onStatusIconClick(1)">
            <i class="status-icon fa fa-2x fa-location-arrow" aria-hidden="true"></i>
			<i v-if="isNew(geocodingStatus)" class="new-update-notification fa fa-circle"></i>
			<i v-if="isPending(geocodingStatus)" class="new-update-notification fa fa-refresh fa-spin"></i>
        </div>
        <div class="status-icon-wrapper" @click="onStatusIconClick(2)">
            <i class="status-icon fa fa-2x fa-table" aria-hidden="true"></i>
			<i v-if="isNew(joinSuggestionStatus)" class="new-update-notification fa fa-circle"></i>
			<i v-if="isPending(joinSuggestionStatus)" class="new-update-notification fa fa-refresh fa-spin"></i>
        </div>
    </div>
</div>
    
</template>

<script lang="ts">

import Vue from 'vue';
import { actions as datasetActions, getters as datasetGetters } from '../store/dataset/module';
import { actions as appActions } from '../store/app/module';
import { getters as routeGetters } from '../store/route/module';
import { DatasetPendingRequestType, DatasetPendingRequest, VariableRankingPendingRequest, DatasetPendingRequestStatus } from '../store/dataset/index';

const STATUS_TYPES = [
	DatasetPendingRequestType.VARIABLE_RANKING,
	DatasetPendingRequestType.GEOCODING,
	DatasetPendingRequestType.JOIN_SUGGESTION,
];

export default Vue.extend({
	name: 'status-sidebar',
	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		pendingRequests: function () {
			const updates = datasetGetters.getPendingRequests(this.$store).filter(update => update.dataset === this.dataset);
			return updates;
		},
		variableRankingRequestData: function () {
			return this.pendingRequests.find(item =>  item.type === DatasetPendingRequestType.VARIABLE_RANKING);
		},
		geocodingRequestData: function () {
			return this.pendingRequests.find(item => item.type === DatasetPendingRequestType.GEOCODING);
		},
		joinSuggestionRequestData: function () {
			return this.pendingRequests.find(item => item.type === DatasetPendingRequestType.JOIN_SUGGESTION);
		},
		variableRankingStatus: function () {
			return this.variableRankingRequestData && this.variableRankingRequestData.status;
		},
		geocodingStatus: function () {
			return this.geocodingRequestData && this.geocodingRequestData.status;
		},
		joinSuggestionStatus: function () {
			return this.joinSuggestionRequestData && this.joinSuggestionRequestData.status;
		},
	},
	methods: {
		isNew(status) {
			return (status === DatasetPendingRequestStatus.RESOLVED) || (status === DatasetPendingRequestStatus.ERROR);
		},
		isPending(status) {
			return DatasetPendingRequestStatus.PENDING === status;
		},
		onStatusIconClick(iconIndex) {
			const statusType = STATUS_TYPES[iconIndex];
			appActions.openStatusPanelWithContentType(this.$store, statusType);
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
