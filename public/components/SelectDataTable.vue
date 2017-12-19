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

				<template :slot="`HEAD_${data.label}`" v-for="data in fields">
					{{data.label}}
					<div :key="data.name">
						<b-dropdown :text="data.type" variant="outline-primary" class="var-type-button">
							<b-dropdown-item
								@click.stop="onTypeChange(data, suggested)"
								:key="suggested.name"
								v-for="suggested in addMissingSuggestions(data.suggested, data.type)">
									{{suggested.type}}
							</b-dropdown-item>
						</b-dropdown>
					</div>
				</template>

			</b-table>
		</div>

	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import { getters as dataGetters, actions } from '../store/data/module';
import { Dictionary } from '../util/dict';
import { FieldInfo } from '../store/data/index';
import { Filter } from '../util/filters';
import { getters as routeGetters } from '../store/route/module';
import { updateTableHighlights } from '../util/highlights';
import { addMissingSuggestions } from '../util/types';

export default Vue.extend({
	name: 'selected-data-table',

	computed: {
		// get dataset from route
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		// extracts the table data from the store
		items(): Dictionary<any> {
			const data = dataGetters.getSelectedDataItems(this.$store);
			const highlights = dataGetters.getHighlightedFeatureRanges(this.$store);
			updateTableHighlights(data, highlights);
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
		addMissingSuggestions(suggested, type) {
			return addMissingSuggestions(suggested, type);
		},
		onTypeChange(field, suggested) {
			actions.setVariableType(this.$store, {
				dataset: this.dataset,
				field: field.label,
				type: suggested.type
			});
		},
		onRowHovered(event) {
			// set new values
			const highlights = {};
			_.forIn(this.fields, (field, key) => {
				highlights[key] = event[key];
			});
			actions.highlightFeatureValues(this.$store, highlights);
		},
		onMouseOut() {
			actions.clearFeatureHighlightValues(this.$store);
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
