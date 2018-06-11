<template>
	<div class="select-data-table">
		<p>
			<b-nav tabs>
				<b-nav-item class="font-weight-bold" @click="includedActive=true" :active="includedActive">Samples to Model From</b-nav-item>
				<b-nav-item class="font-weight-bold" @click="includedActive=false" :active="!includedActive">Excluded Samples</b-nav-item>
			</b-nav>
		</p>

		<div class="table-search-bar">
			<div class="fake-search-input">
				<div class="filter-badges">
					<filter-badge v-if="activeFilter && includedActive"
						active-filter
						:filter="activeFilter">
					</filter-badge>
					<filter-badge v-if="!includedActive && filter.type !== 'row'" v-for="filter in filters" :filter="filter">
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

		<div class="select-data-table-container">
			<div class="select-data-no-results" v-if="!hasData">
				<div v-html="spinnerHTML"></div>
			</div>
			<div class="select-data-no-results" v-if="hasData && items.length===0">
				No data available
			</div>
			<b-table v-if="items.length>0"
				ref="selectTable"
				bordered
				hover
				small
				responsive
				:items="items"
				:fields="fields"
				@row-clicked="onRowClick">

				<template v-for="imageField in imageFields" :slot="imageField" slot-scope="data">
					<image-preview :key="imageField" :image-url="data.item[imageField]"></image-preview>
				</template>

				<template v-for="timeseriesField in timeseriesFields" :slot="timeseriesField" slot-scope="data">
					<sparkline-preview :key="timeseriesField" :time-series-url="data.item[timeseriesField]"></sparkline-preview>
				</template>

			</b-table>
		</div>

	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import { spinnerHTML } from '../util/spinner';
import Vue from 'vue';
import FilterBadge from './FilterBadge';
import SparklinePreview from './SparklinePreview';
import ImagePreview from './ImagePreview';
import { getters as datasetGetters } from '../store/dataset/module';
import { Dictionary } from '../util/dict';
import { Filter } from '../util/filters';
import { TableColumn, D3M_INDEX_FIELD } from '../store/dataset/index';
import { Highlight, RowSelection } from '../store/highlights/index';
import { getters as routeGetters } from '../store/route/module';
import { TableRow } from '../store/dataset/index';
import { addFilterToRoute, EXCLUDE_FILTER, INCLUDE_FILTER } from '../util/filters';
import { getHighlights, clearHighlightRoot, createFilterFromHighlightRoot } from '../util/highlights';
import { addRowSelection, removeRowSelection, clearRowSelection, isRowSelected, getNumIncludedRows, getNumExcludedRows, updateTableRowSelection, createFilterFromRowSelection } from '../util/row';

export default Vue.extend({
	name: 'selected-data-table',

	components: {
		FilterBadge,
		ImagePreview,
		SparklinePreview
	},

	props: {
		instanceName: { type: String, default: 'select-table-highlight' }
	},

	data() {
		return {
			includedActive: true
		};
	},

	computed: {
		// get dataset from route
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		highlights(): Highlight {
			return getHighlights(this.$store);
		},

		numRows(): number {
			return datasetGetters.getIncludedTableDataNumRows(this.$store);
		},

		hasData(): boolean {
			return datasetGetters.hasIncludedTableData(this.$store);
		},

		// extracts the table data from the store
		items(): TableRow[] {
			const items = this.includedActive ? datasetGetters.getIncludedTableDataItems(this.$store) : datasetGetters.getExcludedTableDataItems(this.$store);
			return updateTableRowSelection(items, this.rowSelection, this.instanceName);
		},

		// extract the table field header from the store
		fields(): Dictionary<TableColumn> {
			return this.includedActive ? datasetGetters.getIncludedTableDataFields(this.$store) : datasetGetters.getExcludedTableDataFields(this.$store);
		},

		imageFields(): string[] {
			return _.map(this.fields, (field, name) => {
				return {
					name: name,
					type: field.type
				};
			})
			.filter(field => field.type === 'image')
			.map(field => field.name);
		},

		timeseriesFields(): string[] {
			return _.map(this.fields, (field, name) => {
				return {
					name: name,
					type: field.type
				};
			})
			.filter(field => field.type === 'timeseries')
			.map(field => field.name);
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

		spinnerHTML(): string {
			return spinnerHTML();
		},

		tableTitle(): string {
			if (this.includedActive) {
				const included = getNumIncludedRows(this, this.rowSelection);
				if (included > 0) {
					return `${this.items.length} <b class="matching-color">matching</b> samples of ${this.numRows} to model, ${included} <b class="selected-color">selected</b>`;
				} else {
					return `${this.items.length} <b class="matching-color">matching</b> samples of ${this.numRows} to model`;
				}
			} else {
				const excluded = getNumExcludedRows(this, this.rowSelection);
				if (excluded > 0) {
					return `${this.items.length} <b class="matching-color">matching</b> samples of ${this.numRows} to model, ${excluded} <b class="selected-color">selected</b>`;
				} else {
					return `${this.items.length} <b class="matching-color">matching</b> samples of ${this.numRows} to model`;
				}
			}
		},

		isFilteringHighlights(): boolean {
			return !this.isFilteringSelection && !!this.highlights.root;
		},

		isFilteringSelection(): boolean {
			return !!this.rowSelection;
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

			addFilterToRoute(this, filter);

			if (this.isFilteringHighlights) {
				clearHighlightRoot(this);
			} else {
				clearRowSelection(this);
			}
		},
		onReincludeClick() {
			let filter = null;
			if (this.isFilteringHighlights) {
				filter = createFilterFromHighlightRoot(this.highlights.root, INCLUDE_FILTER);
			} else {
				filter = createFilterFromRowSelection(this.rowSelection, INCLUDE_FILTER);
			}

			addFilterToRoute(this, filter);

			if (this.isFilteringHighlights) {
				clearHighlightRoot(this);
			} else {
				clearRowSelection(this);
			}
		},
		onRowClick(row: TableRow) {
			if (!isRowSelected(this.rowSelection, row[D3M_INDEX_FIELD])) {
				addRowSelection(this, this.instanceName, this.rowSelection, row[D3M_INDEX_FIELD]);
			} else {
				removeRowSelection(this, this.instanceName, this.rowSelection, row[D3M_INDEX_FIELD]);
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

.select-data-table {
	display: flex;
	flex-direction: column;
}
.select-data-table-container {
	display: flex;
	overflow: auto;
	background-color: white;
}
.select-data-no-results {
	width: 100%;
	background-color: #eee;
	padding: 8px;
	text-align: center;
}
.missing-icon {
	padding-right: 4px;
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
.table-search-bar {
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
.filter-badges {

}
.selected-color {
	color: #ff0067;
}
.table-selected-row {
	border-left: 4px solid #ff0067;
	background-color: rgba(255, 0, 103, 0.2);
}
</style>
