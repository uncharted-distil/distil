<template>

	<div class="select-geo-plot">
		<div id="map"></div>
	</div>

</template>

<script lang="ts">

import _ from 'lodash';
import leaflet from 'leaflet';
import Vue from 'vue';
import ImagePreview from './ImagePreview';
import { getters as datasetGetters } from '../store/dataset/module';
import { RowSelection } from '../store/highlights/index';
import { getters as routeGetters } from '../store/route/module';
import { Dictionary } from '../util/dict';
import { TableColumn, TableRow, D3M_INDEX_FIELD } from '../store/dataset/index';
import { addRowSelection, removeRowSelection, isRowSelected, updateTableRowSelection } from '../util/row';

export default Vue.extend({
	name: 'select-geo-plot',

	props: {
		instanceName: String as () => string,
		includedActive: Boolean as () => boolean
	},

	data() {
		return {
			map: null
		};
	},

	mounted() {
		this.map = leaflet.map('map');
	},

	computed: {
	},

	methods: {
	}
});
</script>

<style>

.select-geo-plot {
	display: flex;
	overflow: auto;
	flex-flow: wrap;
}

</style>
