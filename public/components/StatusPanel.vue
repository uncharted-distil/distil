<template>

<div class="status-panel">
	<div @click="$emit('close')">
		{{ statusType }}
		{{ updateData }}
	</div>
</div>
    
</template>

<script lang="ts">

import Vue from 'vue';
import { DatasetPendingUpdate } from '../store/dataset/index';
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
		updateData: function () {
			const update = datasetGetters
				.getPendingUpdates(this.$store)
				.find(update => update.dataset === this.dataset && update.type === this.statusType);
			return update;
		},

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
