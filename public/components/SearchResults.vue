<template>
	<div id="search-results">
		<div v-for="dataset in datasets">
			<div v-on:click="setActiveDataset(dataset.name)">
				<h4>{{dataset.name}}</h4>
			</div>
			<div class="description-text">{{dataset.description}}</div>
		</div>
	</div>
</template>

<script>

import _ from 'lodash';

export default {
	name: 'search-results',

	//data change handlers
	computed: {
		datasets() {
			return this.$store.getters.getDatasets();
		}
	},

	methods: {
		setActiveDataset(datasetName) {
			if (datasetName !== this.$store.state.activeDataset) {
				this.$store.commit('setActiveDataset', datasetName);
				const dataset = _.find(this.$store.state.datasets, d => {
					return d.name === datasetName;
				});
				this.$store.dispatch('getVariableSummaries', dataset);
			}
		}
	}
};
</script>

<style>
.description-text {
	height: 150px;
	overflow: auto;
}
</style>
