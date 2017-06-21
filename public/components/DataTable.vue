<template>
	<div id="data-table">
		<b-table responsive bordered hover :items="items" :fields="fields">
		</b-table>
	</div>
</template>

<script>

import _ from 'lodash';

export default {
	name: 'data-table',
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

</style>
