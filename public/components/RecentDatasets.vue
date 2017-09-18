<template>
	<b-card header="Recent Datasets">
		<div v-if="recentDatasets === null">None</div>
		<b-list-group v-bind:key="dataset.name" v-for="dataset in recentDatasets">
			<dataset-preview
				:name="dataset.name"
				:description="dataset.description">
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

	computed: {
		recentDatasets() {
			// TODO: implement this correctly
			return null;
		}
	}

};
</script>

<style>
</style>
