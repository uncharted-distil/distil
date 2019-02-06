<template>
	<div class="container-fluid d-flex flex-column h-100 join-view">
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
							enable-highlighting
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
							<join-data-slot class="col-12 d-flex flex-column pt-2"
								:dataset="topDataset"
								:items="topDatasetItems"
								:fields="topDatasetFields"
								:numRows="topDatasetNumRows"
								:hasData="topDatasetHasData"
								:selected-column="topColumn"
								:other-selected-column="bottomColumn"
								instance-name="join-dataset-top"
								@col-clicked="onTopColumnClicked"></join-data-slot>
						</div>
						<div class="row responsive-flex pb-3">
							<join-data-slot class="col-12 d-flex flex-column pt-2"
								:dataset="bottomDataset"
								:items="bottomDatasetItems"
								:fields="bottomDatasetFields"
								:numRows="bottomDatasetNumRows"
								:hasData="bottomDatasetHasData"
								:selected-column="bottomColumn"
								:other-selected-column="topColumn"
								instance-name="join-dataset-bottom"
								@col-clicked="onBottomColumnClicked"></join-data-slot>
						</div>
						<div class="row align-items-center">
							<div class="col-12 d-flex flex-column">
								<join-datasets-form class="select-create-solutions"
									:dataset-a="topDataset"
									:dataset-b="bottomDataset"
									:dataset-a-column="topColumn"
									:dataset-b-column="bottomColumn"></join-datasets-form>
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
import JoinDataSlot from '../components/JoinDataSlot.vue';
import VariableFacets from '../components/VariableFacets.vue';
import TypeChangeMenu from '../components/TypeChangeMenu.vue';
import { overlayRouteEntry } from '../util/routes';
import { Dictionary } from '../util/dict';
import { VariableSummary, TableData, TableColumn, TableRow } from '../store/dataset/index';
import { filterSummariesByDataset, NUM_PER_PAGE,
	getTableDataItems, getTableDataFields } from '../util/data';
import { createGroups, Group } from '../util/facets';
import { JOINED_VARS_INSTANCE } from '../store/route/index';
import { actions as viewActions } from '../store/view/module';
import { getters as routeGetters } from '../store/route/module';
import { getters as datasetGetters } from '../store/dataset/module';

export default Vue.extend({
	name: 'join-datasets',

	components: {
		JoinDatasetsForm,
		JoinDataSlot,
		VariableFacets
	},

	computed: {
		joinDatasets(): string[] {
			return routeGetters.getRouteJoinDatasets(this.$store);
		},
		getVariableSummaries(): VariableSummary[] {
			return routeGetters.getJoinDatasetsVariableSummaries(this.$store);
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
		highlightRootStr(): string {
			return routeGetters.getRouteHighlightRoot(this.$store);
		},
		joinedVarsPage(): number {
			return routeGetters.getRouteJoinDatasetsVarsParge(this.$store);
		},
		joinDatasetsTableData(): Dictionary<TableData> {
			return datasetGetters.getJoinDatasetsTableData(this.$store);
		},
		topColumn(): TableColumn {
			const colKey = routeGetters.getJoinDatasetColumnA(this.$store);
			return colKey ? this.topDatasetFields[colKey] : null;
		},
		topDataset(): string {
			return this.joinDatasets.length >= 1 ? this.joinDatasets[0] : null;
		},
		topDatasetTableData(): TableData {
			return this.topDataset ? this.joinDatasetsTableData[this.topDataset] : null;
		},
		topDatasetItems(): TableRow[] {
			return getTableDataItems(this.topDatasetTableData);
		},
		topDatasetFields(): Dictionary<TableColumn> {
			return getTableDataFields(this.topDatasetTableData);
		},
		topDatasetNumRows(): number {
			return this.topDatasetTableData ? this.topDatasetTableData.numRows : 0;
		},
		topDatasetHasData(): boolean {
			return !!this.topDatasetTableData;
		},
		bottomColumn(): TableColumn {
			const colKey = routeGetters.getJoinDatasetColumnB(this.$store);
			return colKey ? this.bottomDatasetFields[colKey] : null;
		},
		bottomDataset(): string {
			return this.joinDatasets.length >= 2 ? this.joinDatasets[1] : null;
		},
		bottomDatasetTableData(): TableData {
			return this.bottomDataset ? this.joinDatasetsTableData[this.bottomDataset] : null;
		},
		bottomDatasetItems(): TableRow[] {
			return getTableDataItems(this.bottomDatasetTableData);
		},
		bottomDatasetFields(): Dictionary<TableColumn> {
			return getTableDataFields(this.bottomDatasetTableData);
		},
		bottomDatasetNumRows(): number {
			return this.bottomDatasetTableData ? this.bottomDatasetTableData.numRows : 0;
		},
		bottomDatasetHasData(): boolean {
			return !!this.bottomDatasetTableData;
		}
	},

	watch: {
		highlightRootStr() {
			viewActions.updateJoinDatasetsData(this.$store);
		},
		joinedVarsPage() {
			viewActions.updateJoinDatasetsData(this.$store);
		}
	},

	beforeMount() {
		viewActions.fetchJoinDatasetsData(this.$store);
	},

	methods: {
		onTopColumnClicked(column) {
			const entry = overlayRouteEntry(this.$route, {
				joinColumnA: column ? column.key : null
			});
			this.$router.push(entry);
		},
		onBottomColumnClicked(column) {
			const entry = overlayRouteEntry(this.$route, {
				joinColumnB: column ? column.key : null
			});
			this.$router.push(entry);
		}
	}
});

</script>

<style>
.join-view .nav-link {
	padding: 1rem 0 0.25rem 0;
	border-bottom: 1px solid #E0E0E0;
	color: rgba(0,0,0,.87);
}
.header-label {
	padding: 1rem 0 0.5rem 0;
	font-weight: bold;
}
.join-view .responsive-flex {
	flex:4;
}
@media (min-width: 1200px) {
	.join-view .responsive-flex {
		flex:6;
	}
}
</style>
