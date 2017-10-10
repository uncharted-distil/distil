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

<script>
import DatasetPreview from '../components/DatasetPreview';

export default {
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
		recentDatasets() {
			const names = this.$store.getters.getRecentDatasets().slice(0, this.maxDatasets);
			return this.$store.getters.getDatasets(names);
		}
	}

};
</script>

<style>
</style>
