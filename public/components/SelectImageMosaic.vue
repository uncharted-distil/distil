<template>

	<div class="select-image-mosaic">
		<template v-for="imageField in imageFields">
			<template v-for="item in items">
				<div class="image-tile">
					<template v-for="(fieldInfo, fieldKey) in fields">
						<image-preview v-if="fieldKey===imageField"
							class="image-preview"
							:row="item"
							:image-url="item[fieldKey]"
							:width="imageWidth"
							:height="imageHeight"
							:on-click="onImageClick"></image-preview>
						<div v-if="fieldKey!==imageField && fieldKey[0] !== '_'"
							class="image-label">
							{{item[fieldKey]}}
						</div>
					</template>
				</div>
			</template>
		</template>
	</div>

</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import ImagePreview from './ImagePreview';
import { getters as datasetGetters } from '../store/dataset/module';
import { RowSelection, TableColumn, TableRow, D3M_INDEX_FIELD } from '../store/dataset/index';
import { getters as routeGetters } from '../store/route/module';
import { Dictionary } from '../util/dict';
import { addRowSelection, removeRowSelection, isRowSelected, updateTableRowSelection } from '../util/row';
import { IMAGE_TYPE } from '../util/types';

export default Vue.extend({
	name: 'select-image-mosaic',

	components: {
		ImagePreview
	},

	props: {
		instanceName: String as () => string,
		includedActive: Boolean as () => boolean
	},

	data() {
		return {
			imageWidth: 128,
			imageHeight: 128
		};
	},

	computed: {

		items(): TableRow[] {
			const items = this.includedActive ? datasetGetters.getIncludedTableDataItems(this.$store) : datasetGetters.getExcludedTableDataItems(this.$store);
			return updateTableRowSelection(items, this.rowSelection, this.instanceName);
		},

		fields(): Dictionary<TableColumn> {
			return this.includedActive ? datasetGetters.getIncludedTableDataFields(this.$store) : datasetGetters.getExcludedTableDataFields(this.$store);
		},

		rowSelection(): RowSelection {
			return routeGetters.getDecodedRowSelection(this.$store);
		},

		imageFields(): string[] {
			return _.map(this.fields, (field, key) => {
				return {
					key: key,
					type: field.type
				};
			})
			.filter(field => field.type === IMAGE_TYPE)
			.map(field => field.key);
		}
	},

	methods: {
		onImageClick(event: any) {
			if (!isRowSelected(this.rowSelection, event.row[D3M_INDEX_FIELD])) {
				addRowSelection(this.$router, this.instanceName, this.rowSelection, event.row[D3M_INDEX_FIELD]);
			} else {
				removeRowSelection(this.$router, this.instanceName, this.rowSelection, event.row[D3M_INDEX_FIELD]);
			}
		}
	}
});
</script>

<style>

.select-image-mosaic {
	display: flex;
	overflow: auto;
	flex-flow: wrap;
}

.image-tile {
	display: flex;
	position: relative;
	flex-grow: 1;
	margin: 4px;
}

.image-preview {
	position: relative;
}

.image-label {
	position: absolute;
	top: 0;
	left: 0;
	background-color: #424242;
	color: #fff;
	padding: 0 4px;
}
</style>
