<template>
	<div class="search-results">
		<div class="row justify-content-center" v-if="isPending">
			<div v-html="spinnerHTML"></div>
		</div>
		<div class="search-results-container" ref="datasetResults">
			<div class="mb-3" :key="dataset.id" v-for="dataset in datasets">
				<dataset-preview
					:dataset="dataset"
					allow-join
					allow-import
					v-on:join-dataset="onJoin">
				</dataset-preview>
			</div>
			<div class="row justify-content-center" v-if="!isPending && (!dataset || datasets.length === 0)">
				<h5>No datasets found for search</h5>
			</div>
		</div>
	</div>
</template>

<script lang="ts">

import $ from 'jquery';
import DatasetPreview from '../components/DatasetPreview';
import Vue from 'vue';
import { spinnerHTML } from '../util/spinner';
import { getters as datasetGetters } from '../store/dataset/module';
import { Dataset } from '../store/dataset/index';

export default Vue.extend({
	name: 'search-results',

	components: {
		DatasetPreview
	},

	props: {
		isPending: Boolean as () => boolean
	},

	computed: {
		datasets(): Dataset[] {
			return datasetGetters.getDatasets(this.$store);
		},
		spinnerHTML(): string {
			return spinnerHTML();
		}
	},

	methods: {
		onJoin(arg) {
			this.$emit('join-dataset', arg);
		}
	},

	watch: {
		datasets() {
			// reset back to top on dataset change
			const $results = this.$refs.datasetResults as Element;
			$results.scrollTop = 0;
		}
	}

});
</script>

<style>
.search-results {
	width: 100%;
}
.search-results-container {
	width: 100%;
	overflow-x: hidden;
	overflow-y: auto;
}
</style>
