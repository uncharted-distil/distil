<template>
	<div class="search-results">
		<div class="bg-faded rounded mb-3" :key="dataset.name" v-for="dataset in datasets">
			<dataset-preview
				:name="dataset.name"
				:description="dataset.description"
				:summary="dataset.summary"
				:summaryML="dataset.summaryML"
				:variables="dataset.variables"
				:numBytes="dataset.numBytes"
				:numRows="dataset.numRows">
			</dataset-preview>
		</div>
	</div>
</template>

<script lang="ts">

import DatasetPreview from '../components/DatasetPreview';
import Vue from 'vue';
import { getters as datasetGetters } from '../store/dataset/module';
import { Dataset } from '../store/dataset/index';

export default Vue.extend({
	name: 'search-results',

	components: {
		DatasetPreview
	},

	computed: {
		datasets(): Dataset[] {
			return datasetGetters.getDatasets(this.$store);
		}
	},

});
</script>

<style>
.search-results {
	width: 100%;
	overflow: auto;
}
</style>
