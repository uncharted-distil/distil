<template>
	<div class="explore-data-table">
		<p class="nav-link font-weight-bold">Samples</p>
		<div class="explore-data-table-container">
			<div class="explore-data-no-results" v-if="items.length===0">
				No results
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
import { getters as dataGetters } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { actions } from '../store/data/module';
import { FieldInfo } from '../store/data/index';
import { Dictionary } from '../util/dict';
import { Filter } from '../util/filters';
import { encodeHighlights } from '../util/highlights';
import { updateTableHighlights, highlightFeatureValues, clearFeatureHighlightValues } from '../util/highlights';
import TypeChangeMenu from '../components/TypeChangeMenu';

export default Vue.extend({
	name: 'explore-data-table',

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
			const data = dataGetters.getFilteredDataItems(this.$store);
			const highlights = dataGetters.getHighlightedFeatureRanges(this.$store);
			updateTableHighlights(data, highlights);
			return data;
		},
		// extract the table field header from the store
		fields(): Dictionary<FieldInfo> {
			return dataGetters.getFilteredDataFields(this.$store);
		},
		filters(): Filter[] {
			return routeGetters.getDecodedFilters(this.$store);
		}
	},

	mounted() {
		this.fetch();
	},

	watch: {
		filters() {
			this.fetch();
		}
	},

	methods: {
		fetch() {
			actions.updateFilteredData(this.$store, {
				dataset: this.dataset,
				filters: this.filters
			});
		},
		onRowHovered(event: Event) {
			// set new values
			const highlights = {};
			_.forIn(this.fields, (field, key) => {
				highlights[key] = event[key];
			});
			highlightFeatureValues(this, highlights);
		},
		onMouseOut() {
			const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
				highlights: null
			});
			this.$router.push(entry);
		}
	}
});
</script>

<style>

.explore-data-table {
	display: flex;
	flex-direction: column;
}
.explore-data-table-container {
	display: flex;
	overflow: auto;
}
.explore-data-no-results {
	width: 100%;
	background-color: #eee;
	padding: 8px;
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
.table-sm th, .table-sm td {
	font-size: 0.9rem;
}
table.b-table>tfoot>tr>th.sorting:before,
table.b-table>thead>tr>th.sorting:before,
table.b-table>tfoot>tr>th.sorting:after,
table.b-table>thead>tr>th.sorting:after {
	top: 0;
}

</style>
