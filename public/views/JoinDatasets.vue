<template>
	<div class="container-fluid d-flex flex-column h-100 select-view">
		<div class="row flex-0-nav"></div>

		<div class="row flex-1 align-items-center justify-content-center bg-white">
			<div class="col-12 col-md-10">
				<h5 class="header-label">Select Features To Join {{joinDatasets[0].toUpperCase()}} with {{joinDatasets[1].toUpperCase()}}</h5>
			</div>
		</div>

		<div class="row flex-10 pb-3">
			<div class="col-12 col-md-3 d-flex flex-column">
				<div class="row flex-12">
					<variable-facets
						class="col-12 d-flex"
							enable-search
							enable-type-change
							:instance-name="instanceName"
							:rows-per-page="numRowsPerPage"
							:groups="groups">
					</variable-facets>
				</div>
			</div>
			<div class="col-12 col-md-9 d-flex flex-column">
				<div class="row flex-12">
					<div class="col-12 d-flex flex-column">
						<div class="row responsive-flex pb-3">
							<select-data-slot class="col-12 d-flex flex-column pt-2"></select-data-slot>
						</div>
						<div class="row responsive-flex pb-3">
							<select-data-slot class="col-12 d-flex flex-column pt-2"></select-data-slot>
						</div>
						<div class="row align-items-center">
							<div class="col-12 d-flex flex-column">
								<join-datasets-form class="select-create-solutions"></join-datasets-form>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>

	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import JoinDatasetsForm from '../components/JoinDatasetsForm.vue';
import SelectDataSlot from '../components/SelectDataSlot.vue';
import VariableFacets from '../components/VariableFacets.vue';
import TypeChangeMenu from '../components/TypeChangeMenu.vue';
import { filterSummariesByDataset, NUM_PER_PAGE } from '../util/data';
import { VariableSummary } from '../store/dataset/index';
import { createGroups, Group } from '../util/facets';
import { JOINED_VARS_INSTANCE } from '../store/route/index';
import { actions as viewActions } from '../store/view/module';
import { getters as routeGetters } from '../store/route/module';

export default Vue.extend({
	name: 'join-datasets',

	components: {
		JoinDatasetsForm,
		SelectDataSlot,
		VariableFacets
	},

	computed: {
		joinDatasets(): string[] {
			return routeGetters.getRouteJoinDatasets(this.$store);
		},
		getVariableSummaries(): VariableSummary[] {
			return routeGetters.getJoinVariableSummaries(this.$store);
		},
		groups(): Group[] {
			return createGroups(this.getVariableSummaries);
		},
		numRowsPerPage(): number {
			return NUM_PER_PAGE;
		},
		instanceName(): string {
			return JOINED_VARS_INSTANCE;
		},
		filtersStr(): string {
			return routeGetters.getRouteFilters(this.$store);
		},
		highlightRootStr(): string {
			return routeGetters.getRouteHighlightRoot(this.$store);
		},
		joinedVarsPage(): number {
			return routeGetters.getRouteJoinedVarsParge(this.$store);
		}
	},

	watch: {
		highlightRootStr() {
			viewActions.updateJoinDatasetsData(this.$store);
		},
		filtersStr() {
			if (this.filtersStr) {
				viewActions.updateJoinDatasetsData(this.$store);
			}
		},
		joinedVarsPage() {
			viewActions.updateJoinDatasetsData(this.$store);
		}
	},

	beforeMount() {
		viewActions.fetchJoinDatasetsData(this.$store);
	}
});

</script>

<style>
.select-view .nav-link {
	padding: 1rem 0 0.25rem 0;
	border-bottom: 1px solid #E0E0E0;
	color: rgba(0,0,0,.87);
}
.header-label {
	padding: 1rem 0 0.5rem 0;
	font-weight: bold;
}
.select-view .responsive-flex {
	flex:4;
}
@media (min-width: 1200px) {
	.select-view .responsive-flex {
		flex:6;
	}
}
</style>
