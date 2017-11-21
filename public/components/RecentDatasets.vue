<template>
	<b-card header="Recent Datasets">
		<div v-if="recentDatasets === null">None</div>
		<b-list-group v-bind:key="dataset.name" v-for="dataset in recentDatasets">
			<dataset-preview
				:name="dataset.name"
				:description="dataset.description"
				:summary="dataset.summary"
				:variables="dataset.variables"
				:numBytes="dataset.numBytes"
				:numRows="dataset.numRows">
			</dataset-preview>
		</b-list-group>
	</b-card>
</template>

<script lang="ts">
import DatasetPreview from '../components/DatasetPreview';
import { getters } from '../store/data/module';
import { filterDatasets, getRecentDatasets } from '../util/data';
import { Datasets } from '../store/data/index';
import Vue from 'vue';

export default Vue.extend({
	name: 'recent-datasets',

	components: {
		DatasetPreview
	},

	props: {
		maxDatasets: {
			default: 5,
			type: Number
		}
	},

	computed: {
		recentDatasets(): Datasets[] {
			const names = getRecentDatasets().slice(0, this.maxDatasets);
			return filterDatasets(names, getters.getDatasets(this.$store));
		}
	}

});
</script>

<style>
</style>
