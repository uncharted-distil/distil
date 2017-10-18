<template>
	<div class="search-results">
		<div class="bg-faded rounded mb-1" :key="dataset.name" v-for="dataset in datasets">
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

<script>

import DatasetPreview from '../components/DatasetPreview';

export default {
	name: 'search-results',

	components: {
		DatasetPreview
	},

	computed: {
		datasets() {
			return this.$store.getters.getDatasets();
		}
	}

};
</script>

<style>
.search-results {
	overflow: auto;
}
</style>
