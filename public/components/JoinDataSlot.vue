<template>
	<div class="join-data-slot d-flex flex-column">

		<div class="fake-search-input">
			<div class="filter-badges">
				<filter-badge v-if="activeFilter"
					active-filter
					:filter="activeFilter">
				</filter-badge>
			</div>
		</div>

		<div class="join-data-container flex-1">
			<div class="join-data-no-results" v-if="!hasData">
				<div v-html="spinnerHTML"></div>
			</div>
			<template v-if="hasData">
				<join-data-table
					:dataset="dataset"
					:items="items"
					:fields="fields"
					:numRows="numRows"
					:hasData="hasData"
					:instance-name="instanceName"
					:selected-column="selectedColumn"
					:other-selected-column="otherSelectedColumn"
					@col-clicked="onColumnClicked"></join-data-table>
			</template>
		</div>

	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import { spinnerHTML } from '../util/spinner';
import { Dictionary } from '../util/dict';
import JoinDataTable from './JoinDataTable';
import FilterBadge from './FilterBadge';
import { TableRow, TableColumn } from '../store/dataset/index';
import { Highlight } from '../store/highlights/index';
import { getHighlights, createFilterFromHighlightRoot } from '../util/highlights';
import { Filter, INCLUDE_FILTER  } from '../util/filters';

export default Vue.extend({
	name: 'join-data-slot',

	components: {
		FilterBadge,
		JoinDataTable
	},

	props: {
		dataset: String as () => string,
		items: Array as () => TableRow[],
		fields: Object as () => Dictionary<TableColumn>,
		numRows: Number as () => number,
		hasData: Boolean as () => boolean,
		selectedColumn: Object as () => TableColumn,
		otherSelectedColumn: Object as () => TableColumn,
		instanceName: String as () => string
	},

	computed: {
		spinnerHTML(): string {
			return spinnerHTML();
		},

		highlights(): Highlight {
			return getHighlights();
		},

		activeFilter(): Filter {
			if (!this.highlights ||
				!this.highlights.root ||
				!this.highlights.root.value) {
				return null;
			}
			if (this.highlights.root.dataset !== this.dataset) {
				return null;
			}
			return createFilterFromHighlightRoot(this.highlights.root, INCLUDE_FILTER);
		}
	},

	methods: {
		onColumnClicked(field) {
			this.$emit('col-clicked', field);
		}
	}
});
</script>

<style>

.join-data-container {
	display: flex;
	background-color: white;
	overflow: auto;
	flex-flow: wrap;
	height: 100%;
	width: 100%;
}
.join-data-no-results {
	width: 100%;
	background-color: #eee;
	padding: 8px;
	text-align: center;
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
</style>
