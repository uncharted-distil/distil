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
</style>
