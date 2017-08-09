<template>
	<div class="search-results">
		<div class="bg-faded rounded mb-1" v-for="dataset in datasets" v-bind:key="dataset.name">
			<div class="dataset-header nav rounded-top">
				<a class="nav-link hover" v-on:click="setActiveDataset(dataset.name)">
					{{dataset.name}}
				</a>
				<a class="nav-link hover" v-on:click="toggleExpansion(dataset.name)">
					<i v-if="isExpanded(dataset.name)" class="fa fa-minus"></i>
					<i v-if="!isExpanded(dataset.name)" class="fa fa-plus"></i>
				</a>
			</div>
			<div class="dataset-body" v-if="isExpanded(dataset.name)">
				<p class="p-2" v-html="dataset.description">
				</p>
			</div>
		</div>
	</div>
</template>

<script>

import Vue from 'vue';
import {createRouteEntry} from '../util/routes';

export default {
	name: 'search-results',

	//data change handlers
	computed: {
		datasets() {
			return this.$store.getters.getDatasets();
		}
	},

	data() {
		return {
			// we don't know dataset names here, so use `Vue.set` to update them
			expanded: {}
		};
	},

	methods: {
		setActiveDataset(datasetName) {
			// clear filters when we select a new dataset
			const filters = datasetName === this.$store.getters.getRouteDataset() ? this.$store.getters.getRouteFilters() : null;
			const entry = createRouteEntry(
				'/dataset',
				datasetName,
				this.$store.getters.getRouteTerms(),
				filters
			);
			this.$router.push(entry);
		},
		toggleExpansion(datasetName) {
			Vue.set(this.expanded, datasetName, !this.expanded[datasetName]);
		},
		isExpanded(datasetName) {
			return this.expanded[datasetName];
		}
	}
};
</script>

<style>
.dataset-header {
	border: 1px solid #ccc;
	justify-content: space-between
}
.dataset-body {
	border-left: 1px solid #ccc;
	border-right: 1px solid #ccc;
	border-bottom: 1px solid #ccc;
}
</style>
