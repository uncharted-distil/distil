<template>
	<div class="search-results">
		<div class="bg-faded rounded mb-3" :key="dataset.name" v-for="dataset in datasets">
			<dataset-preview
				:name="dataset.name"
				:description="dataset.description"
				:summary="dataset.summary"
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
import { getters } from '../store/data/module';
import { Datasets } from '../store/data/index';

export default Vue.extend({
	name: 'search-results',

	components: {
		DatasetPreview
	},

	computed: {
		datasets(): Datasets[] {
			return getters.getDatasets(this.$store);
		}
	},

});
</script>

<style>
.search-results {
	overflow: auto;
}
</style>
