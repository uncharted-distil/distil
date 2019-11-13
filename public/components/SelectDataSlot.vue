<template>
	<div class="select-data-slot">

		<view-type-toggle v-model="viewTypeModel" has-tabs :variables="variables">
			<b-nav-item class="font-weight-bold" @click="setIncludedActive" :active="includedActive">Samples to Model From</b-nav-item>
			<b-nav-item class="font-weight-bold" @click="setExcludedActive" :active="!includedActive">Excluded Samples</b-nav-item>
		</view-type-toggle>

		<div class="fake-search-input">
			<div class="filter-badges">
				<filter-badge v-if="activeFilter"
					active-filter
					:filter="activeFilter">
				</filter-badge>
				<filter-badge v-if="filter.type !== 'row'" v-for="filter in filters" :key="filterHash(filter)" :filter="filter">
				</filter-badge>
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

		<div class="select-data-container" v-bind:class="{ pending: !hasData }">
			<div class="select-data-no-results" v-if="!hasData">
				<div v-html="spinnerHTML"></div>
			</div>
			<template>
				<select-data-table v-if="viewType===TABLE_VIEW" :included-active="includedActive" :instance-name="instanceName"></select-data-table>
				<image-mosaic v-if="viewType===IMAGE_VIEW" :included-active="includedActive" :instance-name="instanceName"></image-mosaic>
				<select-graph-view v-if="viewType===GRAPH_VIEW" :included-active="includedActive" :instance-name="instanceName"></select-graph-view>
				<select-geo-plot v-if="viewType===GEO_VIEW" :included-active="includedActive" :instance-name="instanceName"></select-geo-plot>
				<select-timeseries-view v-if="viewType===TIMESERIES_VIEW" :included-active="includedActive" :instance-name="instanceName"></select-timeseries-view>
			</template>
		</div>

	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import { spinnerHTML } from '../util/spinner';
import SelectDataTable from './SelectDataTable';
import ImageMosaic from './ImageMosaic';
import SelectTimeseriesView from './SelectTimeseriesView';
import SelectGeoPlot from './SelectGeoPlot';
import SelectGraphView from './SelectGraphView';
import FilterBadge from './FilterBadge';
import ViewTypeToggle from './ViewTypeToggle';
import { overlayRouteEntry } from '../util/routes';
import { getters as datasetGetters } from '../store/dataset/module';
import { TableRow, D3M_INDEX_FIELD, Variable, Highlight, RowSelection } from '../store/dataset/index';
import { getters as routeGetters } from '../store/route/module';
import { Filter, addFilterToRoute, EXCLUDE_FILTER, INCLUDE_FILTER } from '../util/filters';
import { clearHighlight, createFilterFromHighlight } from '../util/highlights';
import { addRowSelection, removeRowSelection, clearRowSelection, isRowSelected, getNumIncludedRows, getNumExcludedRows, createFilterFromRowSelection } from '../util/row';
import { actions as appActions } from '../store/app/module';
import { Feature, Activity } from '../util/userEvents';

const TABLE_VIEW = 'table';
const IMAGE_VIEW = 'image';
const GRAPH_VIEW = 'graph';
const GEO_VIEW = 'geo';
const TIMESERIES_VIEW = 'timeseries';

export default Vue.extend({
	name: 'select-data-slot',

	components: {
		FilterBadge,
		SelectDataTable,
		ImageMosaic,
		SelectGraphView,
		SelectGeoPlot,
		SelectTimeseriesView,
		ViewTypeToggle
	},

	data() {
		return {
			instanceName: 'select-data',
			viewTypeModel: null,
			TABLE_VIEW: TABLE_VIEW,
			IMAGE_VIEW: IMAGE_VIEW,
			GRAPH_VIEW: GRAPH_VIEW,
			GEO_VIEW: GEO_VIEW,
			TIMESERIES_VIEW: TIMESERIES_VIEW
		};
	},

	created() {
		this.viewTypeModel = TABLE_VIEW;
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

		includedActive(): boolean {
			return routeGetters.getRouteInclude(this.$store);
		},

		highlight(): Highlight {
			return routeGetters.getDecodedHighlight(this.$store);
		},

		numRows(): number {
			return  this.includedActive ? datasetGetters.getIncludedTableDataNumRows(this.$store) : datasetGetters.getExcludedTableDataNumRows(this.$store);
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
			if (!this.highlight ||
				!this.highlight.value) {
				return null;
			}
			if (this.includedActive) {
				return createFilterFromHighlight(this.highlight, INCLUDE_FILTER);
			}
			return createFilterFromHighlight(this.highlight, EXCLUDE_FILTER);
		},

		filters(): Filter[] {
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
			return !this.isFilteringSelection && !!this.highlight;
		},

		isFilteringSelection(): boolean {
			return !!this.rowSelection;
		},

		viewType(): string {
			return this.viewTypeModel;
		}
	},

	methods: {
		filterHash(filter: Filter) {
			return JSON.stringify(filter);
		},
		onExcludeClick() {
			let filter = null;
			if (this.isFilteringHighlights) {
				filter = createFilterFromHighlight(this.highlight, EXCLUDE_FILTER);
			} else {
				filter = createFilterFromRowSelection(this.rowSelection, EXCLUDE_FILTER);
			}

			addFilterToRoute(this.$router, filter);

			if (this.isFilteringHighlights) {
				clearHighlight(this.$router);
			} else {
				clearRowSelection(this.$router);
			}

			appActions.logUserEvent(this.$store, {
				feature: Feature.FILTER_DATA,
				activity: Activity.DATA_PREPARATION,
				details: {}
			});
		},
		onReincludeClick() {
			let filter = null;
			if (this.isFilteringHighlights) {
				filter = createFilterFromHighlight(this.highlight, INCLUDE_FILTER);
			} else {
				filter = createFilterFromRowSelection(this.rowSelection, INCLUDE_FILTER);
			}

			addFilterToRoute(this.$router, filter);

			if (this.isFilteringHighlights) {
				clearHighlight(this.$router);
			} else {
				clearRowSelection(this.$router);
			}

			appActions.logUserEvent(this.$store, {
				feature: Feature.FILTER_DATA,
				activity: Activity.DATA_PREPARATION,
				details: {}
			});
		},
		setIncludedActive() {
			const entry = overlayRouteEntry(this.$route, {
				include: 'true',
			});
			this.$router.push(entry);

			clearRowSelection(this.$router);
		},
		setExcludedActive() {
			const entry = overlayRouteEntry(this.$route, {
				include: 'false',
			});
			this.$router.push(entry);

			clearRowSelection(this.$router);
		}
	}
});
</script>

<style>

.select-data-container {
	overflow: auto;
	position: relative;
	display: flex;
	background-color: white;
	flex-flow: wrap;
	height: 100%;
	width: 100%;
}
.select-data-no-results {
	position: absolute;
	display: block;
	top: 0;
	height: 100%;
	width: 100%;
	padding: 32px;
	text-align: center;
	opacity: 1;
	z-index: 1;
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
.pending {
	opacity: 0.5;
}
.selected-color {
	color: #ff0067;
}

</style>
