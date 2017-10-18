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
					<div>
						<b-dropdown :text="data.type" variant="outline-primary" class="var-type-button">
							<b-dropdown-item @click.stop="onTypeChange(data, suggested)" v-for="suggested in data.suggested">{{suggested}}</b-dropdown-item>
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
	name: 'explore-data-table',

	computed: {
		// get dataset from route
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		// extracts the table data from the store
		items() {
			return this.$store.getters.getFilteredDataItems();
		},
		// extract the table field header from the store
		fields() {
			return this.$store.getters.getFilteredDataFields();
		},
		filters() {
			return this.$store.getters.getFilters();
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
			this.$store.dispatch('updateFilteredData', {
				dataset: this.dataset,
				filters: this.filters
			});
		},
		onTypeChange(field, type) {
			console.log(field, type);
			/*
			const index = field.suggested.indexOf(type);
			field.suggested.splice(index, 1);
			field.suggested.push(field.type);
			field.type = type;
			*/
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
