<template>

<div class="status-panel">
	<div @click="$emit('close')">
		{{ statusType }}
		{{ requestData }}
	</div>
</div>
    
</template>

<script lang="ts">

import Vue from 'vue';
import { DatasetPendingRequest } from '../store/dataset/index';
import { actions as datasetActions, getters as datasetGetters } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';

export default Vue.extend({
	name: 'status-panel',
	props: {
		statusType: {
			type: String,
			required: true,
		},
	},
	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		requestData: function () {
			const request = datasetGetters
				.getPendingRequests(this.$store)
				.find(request => request.dataset === this.dataset && request.type === this.statusType);
			return request;
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
				})
			}
		}
	},
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
