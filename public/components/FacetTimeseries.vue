<template>
    <div class="facet-timeseries">
        <facet-entry
            :summary="summary"
            :highlight="highlight"
            :row-selection="rowSelection"
			:enable-type-change="enableTypeChange"
            :enable-highlighting="Boolean(enableHighlighting) && enableHighlighting[0]"
            :ignore-highlights="Boolean(ignoreHighlights) && ignoreHighlights[0]"
            :instanceName="instanceName"
            :html="customHtml"
			@html-appended="onHtmlAppend"
            @numerical-click="onNumericalClick"
            @categorical-click="onCategoricalClick"
            @range-change="onRangeChange"
        >
        </facet-entry>
        <facet-entry
			v-if="expanded"
            :summary="summaryHistogram"
            :highlight="highlight"
            :row-selection="rowSelection"
            :instanceName="instanceName"
            :enable-highlighting="Boolean(enableHighlighting) && enableHighlighting[1]"
            :ignore-highlights="Boolean(ignoreHighlights) && ignoreHighlights[1]"
			:html="footerHtml"
            @numerical-click="onHistogramNumericalClick"
            @categorical-click="onHistogramCategoricalClick"
            @range-change="onHistogramRangeChange"
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
		expanded: Boolean as () => boolean,
		highlight: Object as () => Highlight,
		rowSelection: Object as () => RowSelection,
		instanceName: String as () => string,
		enableTypeChange: Boolean as () => boolean,
		enableHighlighting: Array as () => boolean[],
		ignoreHighlights: Array as () => boolean[],
		html: [ String as () => string, Object as () => any, Function as () => Function ],
	},

	data() {
		return {
			customHtml: this.html,
			footerHtml: undefined
		};
	},

	computed: {
	},

	watch: {
		expanded(newExpanded: boolean) {
			if (!newExpanded) {
				// new function that returns the result of this.html
				this.customHtml = (...args) => this.html(...args);
			}
		},
	},
	methods: {
		onCategoricalClick(...args) {
			this.$emit('categorical-click', ...args);
		},
		onNumericalClick(...args) {
			this.$emit('numerical-click', ...args);
		},
		onRangeChange(...args) {
			this.$emit('range-change', ...args);
		},
		onHistogramCategoricalClick(...args) {
			this.$emit('histogram-categorical-click', ...args);
		},
		onHistogramNumericalClick(...args) {
			this.$emit('histogram-numerical-click', ...args);
		},
		onHistogramRangeChange(...args) {
			this.$emit('histogram-range-change', ...args);
		},
		onHtmlAppend(html: HTMLDivElement) {
			// Once html is rendered in top facets, move the element to the bottom facets
			// So that custom html are rendered at the bottom of the coumpound facets
			this.footerHtml = () => html;
		},
	}
});

</script>

<style>

.facet-timeseries .facets-root:first-child {
	margin-bottom: 1px;
}

</style>
