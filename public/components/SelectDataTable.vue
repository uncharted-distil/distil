<template>
	<div class="select-data-table">
		<h6 class="nav-link">Training Set Samples</h6>
		<div class="select-data-table-container">
			<div class="select-data-no-results" v-if="items.length===0">
				<div class="text-danger">
					<i class="fa fa-times missing-icon"></i><strong>No Training features Selected</strong>
				</div>
			</div>
			<b-table v-if="items.length>0"
				bordered
				hover
				striped
				small
				@row-hovered="onRowHovered"
				@mouseout.native="onMouseOut"
				:items="items"
				:fields="fields">

				<template :slot="`HEAD_${field.label}`" v-for="field in fields">
					{{field.label}}
					<type-change-menu
						:key="field.label"
						:field="field.label"></type-change-menu>
				</template>

			</b-table>
		</div>

	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import { getters as dataGetters, actions, mutations } from '../store/data/module';
import { Dictionary } from '../util/dict';
import { FieldInfo } from '../store/data/index';
import { Filter } from '../util/filters';
import { getters as routeGetters } from '../store/route/module';
import { ValueHighlights } from '../store/data/index';
import { updateTableHighlights } from '../util/highlights';
import TypeChangeMenu from '../components/TypeChangeMenu';

const SELECT_TABLE_HIGHLIGHT = 'select_table_highlight';

export default Vue.extend({
	name: 'selected-data-table',

	components: {
		TypeChangeMenu
	},

	computed: {
		// get dataset from route
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		// extracts the table data from the store
		items(): Dictionary<any> {
			const data = dataGetters.getSelectedDataItems(this.$store);
			const rangeHighlights = dataGetters.getHighlightedFeatureRanges(this.$store);
			const valueHighlights = dataGetters.getHighlightedFeatureValues(this.$store);
			updateTableHighlights(data, rangeHighlights, valueHighlights, SELECT_TABLE_HIGHLIGHT);
			return data;
		},

		// extract the table field header from the store
		fields(): Dictionary<FieldInfo> {
			return dataGetters.getSelectedDataFields(this.$store);
		},
		filters(): Filter[] {
			return dataGetters.getSelectedFilters(this.$store);
		}
	},

	mounted() {
		this.fetch();
	},

	watch: {
		'$route.query.training'() {
			this.fetch();
		},
		'$route.query.target'() {
			this.fetch();
		},
		'$route.query.filters'() {
			this.fetch();
		}
	},

	methods: {
		fetch() {
			actions.updateSelectedData(this.$store, {
				dataset: this.dataset,
				filters: this.filters
			});
		},
		onRowHovered(event) {
			// set new values
			const highlights = <ValueHighlights>{
				context: SELECT_TABLE_HIGHLIGHT,
				values: {}
			};
			_.forIn(this.fields, (field, key) => {
				highlights.values[key] = event[key];
			});
			mutations.highlightFeatureValues(this.$store, highlights);
		},
		onMouseOut() {
			mutations.clearFeatureHighlightValues(this.$store);
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
}
.select-data-no-results {
	width: 100%;
	background-color: #eee;
	padding: 8px;
}
.missing-icon {
	padding-right: 4px;
}
.var-type-button {
	width: 100%;
}
.var-type-button button {
	border: none;
	padding: 0;
	width: 100%;
	text-align: left;
	outline: none;
	font-size: 0.9rem;
}
.table-sm th, .table-sm td {
	font-size: 0.9rem;
}
.var-type-button button:hover,
.var-type-button button:active,
.var-type-button button:focus,
.var-type-button.show > .dropdown-toggle  {
	border: none;
	border-radius: 0;
	padding: 0;
	color: inherit;
	background-color: inherit;
	border-color: inherit;
}
table.b-table>tfoot>tr>th.sorting:before,
table.b-table>thead>tr>th.sorting:before,
table.b-table>tfoot>tr>th.sorting:after,
table.b-table>thead>tr>th.sorting:after {
	top: 0;
}
</style>
