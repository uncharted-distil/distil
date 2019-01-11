<template>
	<div class="select-data-slot">
		<p>
			<b-nav tabs>
				<b-nav-item class="font-weight-bold" @click="includedActive=true" :active="includedActive">Samples to Model From</b-nav-item>
				<b-nav-item class="font-weight-bold" @click="includedActive=false" :active="!includedActive">Excluded Samples</b-nav-item>

				<b-form-group class="view-button ml-auto">
					<b-form-radio-group buttons v-model="viewType" button-variant="outline-secondary">
						<b-form-radio value="image" v-if="isImageDataset" class="view-button">
							<i class="fa fa-image"></i>
						</b-form-radio >
						<b-form-radio value="table" class="view-button">
							<i class="fa fa-columns"></i>
						</b-form-radio >
						<b-form-radio value="graph" class="view-button">
							<i class="fa fa-share-alt"></i>
						</b-form-radio >
						<b-form-radio value="geo" class="view-button">
							<i class="fa fa-globe"></i>
						</b-form-radio >
						<b-form-radio value="timeseries" class="view-button">
							<i class="fa fa-line-chart"></i>
						</b-form-radio >
					</b-form-radio-group>
				</b-form-group>
			</b-nav>
		</p>

		<div class="select-search-bar">
			<div class="fake-search-input">
				<div class="filter-badges">
					<filter-badge v-if="activeFilter && includedActive"
						active-filter
						:filter="activeFilter">
					</filter-badge>
					<filter-badge v-if="!includedActive && filter.type !== 'row'" v-for="filter in filters" :key="filter.key" :filter="filter">
					</filter-badge>
				</div>
			</div>
		</div>

		<p class="small-margin">
			<b-button class="float-right" v-if="includedActive"
				variant="outline-secondary"
				:disabled="!isFilteringHighlights && !isFilteringSelection"
				@click="onExcludeClick">
				<i class="fa fa-minus-circle pr-1" v-bind:class="{'exclude-highlight': isFilteringHighlights, 'exclude-selection': isFilteringSelection}"></i>Exclude
			</b-button>
			<b-button class="float-right" v-if="!includedActive"
				variant="outline-secondary"
				:disabled="!isFilteringSelection"
				@click="onReincludeClick">
				<i class="fa fa-plus-circle pr-1" v-bind:class="{'include-selection': isFilteringSelection}"></i>Reinclude
			</b-button>
			<small class="row-number-label" v-html="tableTitle"></small>
		</p>

		<div class="select-data-container">
			<div class="select-data-no-results" v-if="!hasData">
				<div v-html="spinnerHTML"></div>
			</div>
			<template v-if="hasData">
				<select-data-table v-if="viewType==='table'" :included-active="includedActive" :instance-name="instanceName"></select-data-table>
				<select-image-mosaic v-if="viewType==='image'" :included-active="includedActive" :instance-name="instanceName"></select-image-mosaic>
				<select-graph-view v-if="viewType==='graph'" :included-active="includedActive" :instance-name="instanceName"></select-graph-view>
				<select-geo-plot v-if="viewType==='geo'" :included-active="includedActive" :instance-name="instanceName"></select-geo-plot>
				<select-timeseries-view v-if="viewType==='timeseries'" :included-active="includedActive" :instance-name="instanceName"></select-timeseries-view>
			</template>
		</div>

	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import { spinnerHTML } from '../util/spinner';
import SelectDataTable from './SelectDataTable';
import SelectImageMosaic from './SelectImageMosaic';
import SelectTimeseriesView from './SelectTimeseriesView';
import SelectGeoPlot from './SelectGeoPlot';
import SelectGraphView from './SelectGraphView';
import FilterBadge from './FilterBadge';
import { getters as datasetGetters } from '../store/dataset/module';
import { TableRow, D3M_INDEX_FIELD, Variable } from '../store/dataset/index';
import { Highlight, RowSelection } from '../store/highlights/index';
import { getters as routeGetters } from '../store/route/module';
import { Filter, addFilterToRoute, EXCLUDE_FILTER, INCLUDE_FILTER } from '../util/filters';
import { getHighlights, clearHighlightRoot, createFilterFromHighlightRoot } from '../util/highlights';
import { addRowSelection, removeRowSelection, clearRowSelection, isRowSelected, getNumIncludedRows, getNumExcludedRows, createFilterFromRowSelection } from '../util/row';

export default Vue.extend({
	name: 'select-data-slot',

	components: {
		FilterBadge,
		SelectDataTable,
		SelectImageMosaic,
		SelectGraphView,
		SelectGeoPlot,
		SelectTimeseriesView
	},

	data() {
		return {
			instanceName: 'select-data',
			viewType: 'table',
			includedActive: true
		};
	},

	computed: {

		spinnerHTML(): string {
			return spinnerHTML();
		},

		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		variables(): Variable[] {
			return datasetGetters.getVariables(this.$store);
		},

		highlights(): Highlight {
			return getHighlights();
		},

		numRows(): number {
			return datasetGetters.getIncludedTableDataNumRows(this.$store);
		},

		hasData(): boolean {
			return this.includedActive ? datasetGetters.hasIncludedTableData(this.$store) : datasetGetters.hasExcludedTableData(this.$store);
		},

		// extracts the table data from the store
		items(): TableRow[] {
			return this.includedActive ? datasetGetters.getIncludedTableDataItems(this.$store) : datasetGetters.getExcludedTableDataItems(this.$store);
		},

		numItems(): number {
			return this.items ? this.items.length : 0;
		},

		activeFilter(): Filter {
			if (!this.highlights ||
				!this.highlights.root ||
				!this.highlights.root.value) {
				return null;
			}
			if (this.includedActive) {
				return createFilterFromHighlightRoot(this.highlights.root, INCLUDE_FILTER);
			}
			return createFilterFromHighlightRoot(this.highlights.root, EXCLUDE_FILTER);
		},

		filters(): Filter[] {
			if (this.includedActive) {
				return this.invertFilters(routeGetters.getDecodedFilters(this.$store));
			}
			return routeGetters.getDecodedFilters(this.$store);
		},

		rowSelection(): RowSelection {
			return routeGetters.getDecodedRowSelection(this.$store);
		},

		tableTitle(): string {
			if (this.includedActive) {
				const included = getNumIncludedRows(this.rowSelection);
				if (included > 0) {
					return `${this.numItems} <b class="matching-color">matching</b> samples of ${this.numRows} to model, ${included} <b class="selected-color">selected</b>`;
				} else {
					return `${this.numItems} <b class="matching-color">matching</b> samples of ${this.numRows} to model`;
				}
			} else {
				const excluded = getNumExcludedRows(this.rowSelection);
				if (excluded > 0) {
					return `${this.numItems} <b class="matching-color">matching</b> samples of ${this.numRows} to model, ${excluded} <b class="selected-color">selected</b>`;
				} else {
					return `${this.numItems} <b class="matching-color">matching</b> samples of ${this.numRows} to model`;
				}
			}
		},

		isFilteringHighlights(): boolean {
			return !this.isFilteringSelection && !!this.highlights.root;
		},

		isFilteringSelection(): boolean {
			return !!this.rowSelection;
		},

		isImageDataset(): boolean {
			return this.variables.filter(v => v.colType === 'image').length  > 0;
		}
	},

	methods: {
		onExcludeClick() {
			let filter = null;
			if (this.isFilteringHighlights) {
				filter = createFilterFromHighlightRoot(this.highlights.root, EXCLUDE_FILTER);
			} else {
				filter = createFilterFromRowSelection(this.rowSelection, EXCLUDE_FILTER);
			}

			addFilterToRoute(this.$router, filter);

			if (this.isFilteringHighlights) {
				clearHighlightRoot(this.$router);
			} else {
				clearRowSelection(this.$router);
			}
		},
		onReincludeClick() {
			let filter = null;
			if (this.isFilteringHighlights) {
				filter = createFilterFromHighlightRoot(this.highlights.root, INCLUDE_FILTER);
			} else {
				filter = createFilterFromRowSelection(this.rowSelection, INCLUDE_FILTER);
			}

			addFilterToRoute(this.$router, filter);

			if (this.isFilteringHighlights) {
				clearHighlightRoot(this.$router);
			} else {
				clearRowSelection(this.$router);
			}
		},
		onRowClick(row: TableRow) {
			if (!isRowSelected(this.rowSelection, row[D3M_INDEX_FIELD])) {
				addRowSelection(this.$router, this.instanceName, this.rowSelection, row[D3M_INDEX_FIELD]);
			} else {
				removeRowSelection(this.$router, this.instanceName, this.rowSelection, row[D3M_INDEX_FIELD]);
			}
		},
		invertFilters(filters: Filter[]): Filter[] {
			// TODO: invert filters
			return filters;
		}
	}
});
</script>

<style>

.select-data-container {
	display: flex;
	background-color: white;
	overflow: auto;
	flex-flow: wrap;
	height: 100%;
	width: 100%;
}
.select-data-no-results {
	width: 100%;
	background-color: #eee;
	padding: 8px;
	text-align: center;
}
table.b-table>tfoot>tr>th.sorting:before,
table.b-table>thead>tr>th.sorting:before,
table.b-table>tfoot>tr>th.sorting:after,
table.b-table>thead>tr>th.sorting:after {
	top: 0;
}

table tr {
	cursor: pointer;
}
.select-data-table .small-margin {
	margin-bottom: 0.5rem
}
.select-view .nav-tabs .nav-item a {
	padding-left: 0.5rem;
	padding-right: 0.5rem;
}
.select-view .nav-tabs .nav-link {
	color: #757575;
}
.select-view .nav-tabs .nav-link.active {
	color: rgba(0, 0, 0, 0.87);
}
.include-highlight,
.exclude-highlight {
	color: #00c6e1;
}
.include-selection,
.exclude-selection {
	color: #ff0067;
}
.row-number-label {
	position: relative;
	top: 20px;
}
.matching-color {
	color: #00c6e1;
}
.fake-search-input {
	position: relative;
	height: 38px;
	padding: 2px 2px;
	margin-bottom: 4px;
	background-color: #eee;
	border: 1px solid #ccc;
	border-radius: 0.2rem;
}
.selected-color {
	color: #ff0067;
}
.view-button {
	cursor: pointer;
}
</style>
