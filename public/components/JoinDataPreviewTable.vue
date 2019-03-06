<template>
	<fixed-header-table>
		<b-table
			bordered
			hover
			small
			:items="items"
			:fields="fields"
			@sort-changed="onSortChanged">

			<template v-for="imageField in imageFields" :slot="imageField" slot-scope="data">
				<image-preview :key="imageField" :image-url="data.item[imageField]"></image-preview>
			</template>

			<template v-for="timeseriesField in timeseriesFields" :slot="timeseriesField" slot-scope="data">
				<sparkline-preview :key="timeseriesField" :timeseries-url="data.item[timeseriesField]"></sparkline-preview>
			</template>

		</b-table>
	</fixed-header-table>

</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import FixedHeaderTable from './FixedHeaderTable';
import SparklinePreview from './SparklinePreview';
import ImagePreview from './ImagePreview';
import { Dictionary } from '../util/dict';
import { TableColumn, TableRow, D3M_INDEX_FIELD } from '../store/dataset/index';
import { getters as routeGetters } from '../store/route/module';
import { IMAGE_TYPE, TIMESERIES_TYPE } from '../util/types';

export default Vue.extend({
	name: 'join-data-preview-table',

	components: {
		ImagePreview,
		SparklinePreview,
		FixedHeaderTable,
	},

	props: {
		items: Array as () => TableRow[],
		fields: Object as () => Dictionary<TableColumn>,
		instanceName: String as () => string
	},

	computed: {

		imageFields(): string[] {
			return _.map(this.fields, (field, key) => {
				return {
					key: key,
					type: field.type
				};
			})
			.filter(field => field.type === IMAGE_TYPE)
			.map(field => field.key);
		},

		timeseriesFields(): string[] {
			return _.map(this.fields, (field, key) => {
				return {
					key: key,
					type: field.type
				};
			})
			.filter(field => field.type === TIMESERIES_TYPE)
			.map(field => field.key);
		}
	},

	methods: {
		onSortChanged() {
			// need a `nextTick` otherwise the cells get immediately overwritten
			Vue.nextTick(() => {
				const fixedHeaderTable = this.$refs.fixedHeaderTable as any;
				fixedHeaderTable.resizeTableCells();
			});
		}
	}
});
</script>

<style>

</style>
