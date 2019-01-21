<template>
	<div class="container-fluid d-flex flex-column h-100 search-view">
		<div class="row flex-0-nav"></div>

		<div class="row flex-1 align-items-center justify-content-center bg-white">
			<div class="col-12 col-md-10">
				<h5 class="header-label">Select a Dataset</h5>
			</div>
		</div>
		<div class="row flex-2 align-items-center justify-content-center">
			<div class="col-12 col-md-6">
				<search-bar class="search-search-bar"></search-bar>
			</div>
		</div>
		<div class="row flex-10 justify-content-center pb-3">
			<div class="col-12 col-md-10 d-flex">
				<search-results class="search-search-results"
					v-on:join-dataset="onJoin">
				</search-results>
			</div>
		</div>
		<div v-if="numJoiningDatasets !== 0">
			<div class="row flex-1 align-items-center justify-content-center bg-white">
				<div class="col-12 col-md-10">
					<h5 class="header-label">Join Datasets</h5>
					<b-button size="sm" variant="secondary" class="close-join-button" @click="closeJoin"><i class="fa fa-times"></i></b-button>
				</div>
			</div>
			<div class="row flex-1 align-items-center justify-content-center">
				<div class="col-4 mb-3" v-for="dataset in joiningDatasets">
					<dataset-preview-card
						:dataset="dataset">
					</dataset-preview-card>
				</div>
				<div class="join-button col-2 mb-3" v-if="numJoiningDatasets === 2">
					<b-button size="lg" large variant="primary" @click="onJoinDatasets">Join Datasets</b-button>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import DatasetPreviewCard from '../components/DatasetPreviewCard.vue';
import SearchBar from '../components/SearchBar.vue';
import SearchResults from '../components/SearchResults.vue';
import { Dataset } from '../store/dataset/index';
import { createRouteEntry } from '../util/routes';
import { getters as routeGetters } from '../store/route/module';
import { getters as datasetGetters, actions as datasetActions } from '../store/dataset/module';
import { JOIN_DATASETS_ROUTE } from '../store/route/index';

export default Vue.extend({
	name: 'search-view',

	components: {
		SearchBar,
		SearchResults,
		DatasetPreviewCard
	},

	data() {
		return {
			joiningDatasets: {}
		};
	},

	computed: {
		terms(): string {
			return routeGetters.getRouteTerms(this.$store);
		},
		datasets(): Dataset[] {
			return datasetGetters.getDatasets(this.$store);
		},
		numJoiningDatasets(): number {
			return _.size(this.joiningDatasets);
		}
	},

	beforeMount() {
		this.fetch();
	},

	watch: {
		terms() {
			this.fetch();
		}
	},

	methods: {
		fetch() {
			datasetActions.searchDatasets(this.$store, this.terms);
		},
		onJoin(id) {
			if (this.numJoiningDatasets !== 2) {
				const dataset = _.find(this.datasets, d => {
					return d.id === id;
				});
				Vue.set(this.joiningDatasets, id, dataset);
			}
		},
		closeJoin() {
			this.joiningDatasets = {};
		},
		onJoinDatasets() {
			if (this.numJoiningDatasets === 2) {
				const datasets = _.keys(this.joiningDatasets);
				const entry = createRouteEntry(JOIN_DATASETS_ROUTE, {
					joinDatasets: datasets.join(',')
				});
				this.$router.push(entry);
			}
		}

	}
});
</script>

<style>
.header-label {
	padding: 1rem 0 0.5rem 0;
	font-weight: bold;
}
.search-search-bar {
	width: 100%;
	box-shadow: 0 1px 2px 0 rgba(0,0,0,0.10);
}
.close-join-button {
	position: absolute;
	top: 4px;
	right: 4px;
	cursor: pointer;
}
.join-button {
	text-align: center;
}
</style>
