<template>
	<div class="facet-timeseries">
		timeseries
		<facet-entry
			:summary="summary"
			:highlight="highlight"
			:row-selection="rowSelection"
			:enabled-type-changes="enabledTypeChanges"
			:enable-highlighting="Boolean(enableHighlighting) && enableHighlighting[0]"
			:ignore-highlights="Boolean(ignoreHighlights) && ignoreHighlights[0]"
			:instanceName="instanceName"
			:html="customHtml"
			@html-appended="onHtmlAppend"
			@numerical-click="onNumericalClick"
			@categorical-click="onCategoricalClick"
			@range-change="onRangeChange">
		</facet-entry>
		<facet-entry v-if="!!timelineSummary"
			:summary="timelineSummary"
			:highlight="highlight"
			:row-selection="rowSelection"
			:enabled-type-changes="enabledTypeChanges"
			:instanceName="instanceName"
			:enable-highlighting="Boolean(enableHighlighting) && enableHighlighting[1]"
			:ignore-highlights="Boolean(ignoreHighlights) && ignoreHighlights[1]"
			:html="footerHtml"
			@numerical-click="onHistogramNumericalClick"
			@categorical-click="onHistogramCategoricalClick"
			@range-change="onHistogramRangeChange">
		</facet-entry>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import FacetEntry from '../components/FacetEntry';
import { getters as datasetGetters } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';
import { Dataset, Variable, VariableSummary, Highlight, RowSelection, Row, NUMERICAL_SUMMARY } from '../store/dataset/index';
import { INTEGER_TYPE } from '../util/types';

export default Vue.extend({
	name: 'facet-timeseries',

	components: {
		FacetEntry
	},

	props: {
		summary: Object as () => VariableSummary,
		highlight: Object as () => Highlight,
		rowSelection: Object as () => RowSelection,
		instanceName: String as () => string,
		enabledTypeChanges: Array as () => string[],
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
		variables(): Variable[] {
			return datasetGetters.getVariables(this.$store);
		},
		variable(): Variable {
			return this.variables.find(v => v.colName === this.summary.key);
		},
		timelineSummary(): VariableSummary {

			if (this.summary.pending) {
				return null;
			}

			let timeVarName = '';
			let timeVarType = '';
			let timeVar = null;
			const summaryVar = this.variables.find(v => v.colName === this.summary.key);
			if (!summaryVar) {
				return null;
			}

			const grouping = this.variable.grouping;
			if (!grouping) {
				return null;
			}
			timeVarName = grouping.properties.xCol;
			timeVar = this.variables.find(v => v.colName === timeVarName);
			timeVarType = timeVar ? timeVar.colType : INTEGER_TYPE;

			if (this.summary.pending || !this.variable) {
				return null;
			}

			return {
				label: timeVarName,
				key: timeVarName,
				dataset: this.summary.dataset,
				description: this.summary.description,
				type: NUMERICAL_SUMMARY,
				varType: timeVarType,
				baseline: this.summary.timeline
			};
		}
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
