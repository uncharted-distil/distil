<template>
	<div class="explore-data-table">
		<h6 class="nav-link">Samples</h6>
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

				<template :slot="`HEAD_${data.label}`" v-for="data in fields">
					{{data.label}}
					<div :key="data.label">
						<b-dropdown :text="data.type" variant="outline-primary" class="var-type-button">
							<b-dropdown-item
								v-bind:class="probabilityCategoryClass(suggested.probability)"
								@click.stop="onTypeChange(data, suggested)"
								:key="suggested.name"
								v-for="suggested in data.suggested">
									{{suggested.type}} ({{probabilityCategoryText(suggested.probability)}})
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
import { getters as dataGetters } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { actions } from '../store/data/module';
import { Dictionary } from '../store/data/index';
import { FilterMap } from '../util/filters';
import { FieldInfo } from '../store/data/getters';
import { updateTableHighlights } from '../util/highlights';

const LOW_PROBABILITY = 0.33;
const MED_PROBABILITY = 0.66;

export default Vue.extend({
	name: 'explore-data-table',

	computed: {
		// get dataset from route
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		// extracts the table data from the store
		items(): Dictionary<any> {
			const data = dataGetters.getFilteredDataItems(this.$store);
			updateTableHighlights(data, dataGetters.getHighlightedFeatureRanges(this.$store));
			return data;
		},
		// extract the table field header from the store
		fields(): Dictionary<FieldInfo> {
			return dataGetters.getFilteredDataFields(this.$store);
		},
		filters(): FilterMap {
			return routeGetters.getDecodedFilters(this.$store);
		}
	},

	mounted() {
		this.fetch();
	},

	watch: {
		'$route.query.filters'() {
			this.fetch();
		},
		'$route.query.dataset'() {
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
		probabilityCategoryText(probability) {
			if (probability < LOW_PROBABILITY) {
				return 'Low';
			}
			if (probability < MED_PROBABILITY) {
				return 'Med';
			}
			return 'High';
		},
		probabilityCategoryClass(probability) {
			if (probability < LOW_PROBABILITY) {
				return 'text-danger';
			}
			if (probability < MED_PROBABILITY) {
				return 'text-warning';
			}
			return 'text-success';
		},
		onTypeChange(field: { label: string }, suggested: { type: string }) {
			actions.setVariableType(this.$store, {
				dataset: this.dataset,
				field: field.label,
				type: suggested.type
			});
		},
		onRowHovered(event: Event) {
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
