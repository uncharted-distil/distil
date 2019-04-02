<template>

<div class="status-panel" v-if="isOpen">
	<div>
		<div class="heading">
			<h4>{{ contentData.title }}</h4>
			<div class="close-button" @click="close()">
				<i class="fa fa-2x fa-times" aria-hidden="true"></i>
			</div>
		</div>
		{{title}}
		{{ isOpen }}
		{{ statusType }}
		{{ requestData }}
		{{ statusPanelState }}
	</div>
</div>
    
</template>

<script lang="ts">

import Vue from 'vue';
import { DatasetPendingRequest, DatasetPendingRequestType } from '../store/dataset/index';
import { actions as datasetActions, getters as datasetGetters } from '../store/dataset/module';
import { actions as appActions, getters as appGetters } from '../store/app/module';
import { getters as routeGetters } from '../store/route/module';
import { StatusPanelState, StatusPanelContentType } from '../store/app';

export default Vue.extend({
	name: 'status-panel',
	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		statusPanelState(): StatusPanelState {
			return appGetters.getStatusPanelState(this.$store);
		},
		isOpen(): boolean {
			return this.statusPanelState.isOpen;
		},
		statusType(): StatusPanelContentType {
			return this.statusPanelState.contentType;
		},
		requestData: function () {
			const request = datasetGetters
				.getPendingRequests(this.$store)
				.find(request => request.dataset === this.dataset && request.type === this.statusType);
			return request;
		},
		contentData(): {
			title: string,
			pendingMsg?: string,
			resolvedMsg?: string,
			defaultMsg?: string,
		} {
			switch (this.statusType) {
				case DatasetPendingRequestType.VARIABLE_RANKING:
					return {
						title: 'Variable Ranking',
						pendingMsg: '',
						resolvedMsg: '',
						defaultMsg: '',
					};
				case DatasetPendingRequestType.GEOCODING:
					return {
						title: 'Geo Coding',
						pendingMsg: '',
						resolvedMsg: '',
						defaultMsg: '',
					};
				case DatasetPendingRequestType.JOIN_SUGGESTION:
					return {
						title: 'Join Suggestion',
						pendingMsg: '',
						resolvedMsg: '',
						defaultMsg: '',
					};
				default:
					return {
						title: ''
					};
			}
		},
	},
	watch: {
		requestData: function (data) {
			// when pending get resolved, change the status to reviewed
			if (data && (data.status === 'resolved' || data.status === 'error')) {
				const { status } = data;
				datasetActions.updatePendingRequestStatus(this.$store, {
					id: data.id,
					status: 'reviewed',
				});
			}
		}
	},
	methods: {
		close() {
			appActions.closeStatusPanel(this.$store);
		}
	}
});

</script>

<style>

.status-panel {
	box-shadow: 0 4px 5px 0 rgba(0,0,0,0.14), 0 1px 10px 0 rgba(0,0,0,0.12), 0 2px 4px -1px rgba(0,0,0,0.2);
	display: block;
	position: fixed;
	right: 0;
	top: 0;
	bottom: 0;
	z-index: 100;
	width: 300px;
	height: 100%;
	background: #fff;
}

</style>
