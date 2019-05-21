<template>
	<geo-plot
		:instance-name="instanceName"
		:data-fields="fields"
		:data-items="items"
		:selection="rowSelection"
		@selectmarker="onSelectMakrer">
	</geo-plot>
</template>

<script lang="ts">

import Vue from 'vue';
import GeoPlot from './GeoPlot';
import { getters as datasetGetters } from '../store/dataset/module';
import { Dictionary } from '../util/dict';
import { getters as routeGetters } from '../store/route/module';
import { TableColumn, TableRow, D3M_INDEX_FIELD } from '../store/dataset/index';
import { addRowSelection, removeRowSelection, isRowSelected, updateTableRowSelection } from '../util/row';
import { RowSelection } from '../store/highlights/index'

export default Vue.extend({
	name: 'select-geo-plot',

	components: {
		GeoPlot
	},

	props: {
		instanceName: String as () => string,
		includedActive: Boolean as () => boolean
	},

	computed: {

		fields(): Dictionary<TableColumn> {
			return this.includedActive ? datasetGetters.getIncludedTableDataFields(this.$store) : datasetGetters.getExcludedTableDataFields(this.$store);
		},

		items(): TableRow[] {
			return this.includedActive ? datasetGetters.getIncludedTableDataItems(this.$store) : datasetGetters.getExcludedTableDataItems(this.$store);
		},
		itemsWithSelction(): TableRow[] {
			const selection = this.rowSelection;
			return this.items.map(item => {
				return item;
			});
		},
		rowSelection(): RowSelection {
			return routeGetters.getDecodedRowSelection(this.$store);
		},

	},

	methods: {
		onSelectMakrer(data) {
			const row = data.point.row;
			if (data.isSelected) {
				addRowSelection(this.$router, this.instanceName, this.rowSelection, row[D3M_INDEX_FIELD]);
			} else {
				removeRowSelection(this.$router, this.instanceName, this.rowSelection, row[D3M_INDEX_FIELD]);
			}
		},
	}
});

</script>

<style>
</style>
