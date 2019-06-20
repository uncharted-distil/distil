<template>
    <div class="facet-timeseries">
        <facet-entry
            :summary="summary"
            :highlight="highlight"
            :row-selection="rowSelection"
			:enable-type-change="enableTypeChange"
            :enable-highlighting="enableHighlighting"
            :ignore-highlights="ignoreHighlights"
            :instanceName="instanceName"
            @numerical-click="onNumericalClick"
            @categorical-click="onCategoricalClick"
            @range-change="onRangeChange"
        >
        </facet-entry>
        <facet-entry
			v-if="expanded"
            :summary="summaryHistogram"
            :row-selection="rowSelection"
            :html="html"
            :enable-highlighting="enableHighlighting"
            :ignore-highlights="ignoreHighlights"
            :instanceName="instanceName"
            @numerical-click="onNumericalClick"
            @categorical-click="onCategoricalClick"
            @range-change="onRangeChange"
        >
        </facet-entry>
    </div>
</template>

<script lang="ts">

import FacetEntry from '../components/FacetEntry';
import { VariableSummary, Highlight, RowSelection, Row } from '../store/dataset/index';
import Vue from 'vue';

export default Vue.extend({
	name: 'facet-timeseries',

	components: {
		FacetEntry
	},

	props: {
		summary: Object as () => VariableSummary,
		summaryHistogram: Object as () => VariableSummary,
		expanded: Object as () => boolean,
		highlight: Object as () => Highlight,
		rowSelection: Object as () => RowSelection,
		enableTypeChange: Boolean as () => boolean,
		enableHighlighting: Boolean as () => boolean,
		ignoreHighlights: Boolean as () => boolean,
		instanceName: String as () => string,
		html: [ String as () => string, Object as () => any, Function as () => Function ],
	},

	data() {
		return {
		};
	},

	computed: {
	},

	methods: {
		onCategoricalClick(context: string, ...rest) {
			this.$emit('categorical-click', ...rest);
		},
		onNumericalClick(context: string, ...rest) {
			this.$emit('numerical-click', ...rest);
		},
		onRangeChange(context: string, ...rest) {
			this.$emit('range-change', ...rest);
		},
	}
});

</script>

<style>

.facet-timeseries .facets-root:first-child {
	margin-bottom: 1px;
}

</style>
