<template>
	<div class="select-data-table">
		<h6 class="nav-link">Training Set Samples</h6>
		<div class="select-data-table-container">
			<div class="select-data-no-results" v-if="items.length===0">
				<div class="text-danger">
					<i class="fa fa-times missing-icon"></i><strong>No Training Features Selected</strong>
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
					<div>
						<b-dropdown :text="data.type" variant="outline-primary" class="var-type-button">
							<b-dropdown-item @click.stop="onTypeChange(data, suggested)":key="suggested.name" v-for="suggested in data.suggested">{{suggested.type}} ({{suggested.probability.toFixed(2)}})</b-dropdown-item>
						</b-dropdown>
					</div>
				</template>

			</b-table>
		</div>

	</div>
</template>

<script>
import _ from 'lodash';

export default {
	name: 'selected-data-table',

	computed: {
		// get dataset from route
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		// extracts the table data from the store
		items() {
			return this.$store.getters.getSelectedDataItems();
		},
		// extract the table field header from the store
		fields() {
			return this.$store.getters.getSelectedDataFields();
		},
		filters() {
			return this.$store.getters.getSelectedFilters();
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
			this.$store.dispatch('updateSelectedData', {
				dataset: this.dataset,
				filters: this.filters
			});
		},
		onTypeChange(field, suggested) {
			this.$store.dispatch('setVariableType', {
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
			this.$store.dispatch('highlightFeatureValues', highlights);
		},
		onMouseOut() {
			this.$store.dispatch('clearFeatureHighlightValues');
		}
	}
};
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
