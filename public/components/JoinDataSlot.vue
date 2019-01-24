<template>
	<div class="join-data-slot">

		<!-- <div class="join-search-bar">
			<div class="fake-search-input">
				<div class="filter-badges">
					<filter-badge v-if="activeFilter"
						active-filter
						:filter="activeFilter">
					</filter-badge>
					</filter-badge>
				</div>
			</div>
		</div> -->

		<div class="join-data-container">
			<div class="join-data-no-results" v-if="!hasData">
				<div v-html="spinnerHTML"></div>
			</div>
			<template v-if="hasData">
				<join-data-table
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

export default Vue.extend({
	name: 'join-data-slot',

	components: {
		FilterBadge,
		JoinDataTable
	},

	props: {
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
table.b-table>tfoot>tr>th.sorting:before,
table.b-table>thead>tr>th.sorting:before,
table.b-table>tfoot>tr>th.sorting:after,
table.b-table>thead>tr>th.sorting:after {
	top: 0;
}

table tr {
	cursor: pointer;
}
.join-data-table .small-margin {
	margin-bottom: 0.5rem
}
.join-view .nav-tabs .nav-item a {
	padding-left: 0.5rem;
	padding-right: 0.5rem;
}
.join-view .nav-tabs .nav-link {
	color: #757575;
}
.join-view .nav-tabs .nav-link.active {
	color: rgba(0, 0, 0, 0.87);
}
.include-highlight,
.exclude-highlight {
	color: #00c6e1;
}
.include-joinion,
.exclude-joinion {
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
.joined-color {
	color: #ff0067;
}
.view-button {
	cursor: pointer;
}
</style>
