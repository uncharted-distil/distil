<template>
	<div class="data-table">
		<div class="nav bg-faded rounded-top">
			<h6 class="nav-link">Values</h6>
		</div>
		<div class="table-container">
			<b-table
				responsive
				bordered
				hover
				striped
				small
				:items="items"
				:fields="fields"
				:current-page="currentPage">
			</b-table>
		</div>
		<!--

		:per-page="perPage"

		<div v-if="items.length>0" class="justify-content-center row my-1">
			<b-pagination
				:total-rows="items.length"
				:per-page="perPage"
				v-model="currentPage" />
		</div>
		-->
	</div>
</template>

<script>

import _ from 'lodash';

export default {
	name: 'data-table',

	data() {
		return {
			perPage: 10,
			currentPage: 1
		};
	},

	computed: {
		// extracts the table data from the store
		items() {
			const data = this.$store.getters.getFilteredData();
			if (!_.isEmpty(data)) {
				return _.map(data.values, d => {
					const rowObj = {};
					for (const [idx, varMeta] of data.metadata.entries()) {
						rowObj[varMeta.name] = d[idx];
					}
					return rowObj;
				});
			} else {
				return [];
			}
		},
		// extract the table field header from the store
		fields() {
			const data = this.$store.getters.getFilteredData();
			if (!_.isEmpty(data)) {
				const result = {};
				for (let varMeta of data.metadata) {
					result[varMeta.name] = {
						label: varMeta.name
					};
				}
				return result;
			} else {
				return {};
			}
		}
	}
};
</script>

<style>
.data-table {
}
.table-container {
	overflow: auto;
}
</style>
