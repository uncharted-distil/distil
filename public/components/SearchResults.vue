<template>
	<div class="search-results" ref="datasetResults">
		<div class="mb-3" :key="dataset.id" v-for="dataset in datasets">
			<dataset-preview
				:dataset="dataset"
				allow-join
				allow-import
				v-on:join-dataset="onJoin">
			</dataset-preview>
		</div>
	</div>
</template>

<script lang="ts">

import $ from 'jquery';
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
	overflow: auto;
}
</style>
