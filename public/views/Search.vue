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
					:is-pending="isPending"
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
				<div class="col-4 mb-3" v-for="dataset in joinDatasets">
					<dataset-preview-card
						:dataset="dataset"
						@remove-from-join="onRemoveFromJoin">
					</dataset-preview-card>
				</div>
				<div class="join-button col-2 mb-3">
					<b-button size="lg" large variant="primary" :disabled="numJoiningDatasets !== 2" @click="onJoinDatasets">
						<div class="row justify-content-center join-datasets-button pl-4 pr-4">
							<i class="fa fa-compress mr-2"></i>
							<b>Join Datasets</b>
						</div>
					</b-button>
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
import { createRouteEntry, overlayRouteEntry } from '../util/routes';
import { getters as routeGetters } from '../store/route/module';
import { getters as datasetGetters, actions as datasetActions } from '../store/dataset/module';
import { SEARCH_ROUTE, JOIN_DATASETS_ROUTE } from '../store/route/index';

export default Vue.extend({
	name: 'search-view',

	components: {
		SearchBar,
		SearchResults,
		DatasetPreviewCard
	},

	data() {
		return {
			isPending: false
		};
	},

	computed: {
		terms(): string {
			return routeGetters.getRouteTerms(this.$store);
		},
		datasets(): Dataset[] {
			return datasetGetters.getDatasets(this.$store);
		},
		joinDatasetIDs(): string[] {
			return routeGetters.getRouteJoinDatasets(this.$store);
		},
		joinDatasets(): Dataset[] {
			const lookup = {};
			this.joinDatasetIDs.forEach(id => {
				lookup[id] = true;
			});
			return this.datasets.filter(d => {
				return lookup[d.id];
			});
		},
		numJoiningDatasets(): number {
			return this.joinDatasetIDs.length;
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
			this.isPending = true;
			datasetActions.searchDatasets(this.$store, this.terms)
				.then(() => {
					this.isPending = false;
				});
		},
		onJoin(id) {
			// check if already exists
			const exists = _.find(this.joinDatasetIDs, datasetID => {
				return datasetID === id;
			});
			if (exists) {
				return;
			}

			// otherwise add
			const joinDatasetIDs = this.joinDatasetIDs;
			if (joinDatasetIDs.length !== 2) {
				joinDatasetIDs.push(id);
				const entry = overlayRouteEntry(this.$route, {
					joinDatasets: joinDatasetIDs.join(','),
				});
				this.$router.push(entry);
			}
		},
		closeJoin() {
			const entry = createRouteEntry(SEARCH_ROUTE, {
				terms: this.terms
			});
			this.$router.push(entry);
		},
		onJoinDatasets() {
			if (this.numJoiningDatasets === 2) {
				const entry = createRouteEntry(JOIN_DATASETS_ROUTE, {
					joinDatasets: this.joinDatasetIDs.join(',')
				});
				this.$router.push(entry);
			}
		},
		onRemoveFromJoin(datasetID) {
			const joinDatasetIDs = this.joinDatasetIDs.filter(id => {
				return id !== datasetID;
			});
			const entry = overlayRouteEntry(this.$route, {
				joinDatasets: joinDatasetIDs.join(','),
			});
			this.$router.push(entry);
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
.join-datasets-button,
.join-datasets-button i {
	line-height: 32px !important;
	text-align: center;
}
</style>
