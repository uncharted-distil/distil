<template>
<div class="status-sidebar">
    <div class="status-icons">
        <div class="event-icon-wrapper">
            <i class="event-icon fa fa-2x fa-info" aria-hidden="true"></i>
			<i v-if="variableRankingStatus === 'done'" class="new-update-notification fa fa-refresh fa-circle"></i>
			<i v-if="variableRankingStatus === 'pending'" class="new-update-notification fa fa-refresh fa-spin"></i>
        </div>
        <div class="event-icon-wrapper">
            <i class="event-icon fa fa-2x fa-location-arrow" aria-hidden="true"></i>
			<i v-if="hasNewGeocodingUpdate" class="new-update-notification fa fa-circle"></i>
        </div>
        <div class="event-icon-wrapper">
            <i class="event-icon fa fa-2x fa-long-arrow-right" aria-hidden="true"></i>
			<i v-if="hasNewJoinSuggestionUpdate" class="new-update-notification fa fa-circle"></i>
        </div>
    </div>
</div>
    
</template>

<script lang="ts">

import Vue from 'vue';
import { actions as datasetActions, getters as datasetGetters } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';
import { DatasetPendingUpdateType, DatasetPendingUpdate, VariableRankingPendingUpdate } from '../store/dataset/index';

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
			const up = this.pendingUpdates.find(item =>  item.type === DatasetPendingUpdateType.VARIABLE_RANKING); 
			console.log('var ranking update');
			console.log(up);
			return <VariableRankingPendingUpdate>up;
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
		hasNewRankingUpdate: function () {
			return Boolean(this.variableRankingUpdate);
		},
		isPending: function () {
			console.log('status update');
			return this.variableRankingUpdate && (this.variableRankingUpdate.status === 'pending');
		},
		hasNewGeocodingUpdate: function () {
			return Boolean(this.geocodingUpdate);
		},
		hasNewJoinSuggestionUpdate: function () {
			return Boolean(this.joinSuggestionUpdate);
		},
	},
	mounted() {
		console.log('mounted with this dataset', this.dataset);
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
.status-sidebar .event-icon-wrapper {
	padding-top: 15px;
	padding-bottom: 15px;
	text-align: center;
	position: relative;
}
.status-sidebar .event-icon {
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
