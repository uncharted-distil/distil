<script lang="ts">

import VariableFacets from './VariableFacets.vue';
import { mutations as dataMutations } from '../store/data/module';
import { NUMERICAL_FILTER, CATEGORICAL_FILTER } from '../util/filters';
import { updateResultHighlights } from '../util/highlights';
import { Range } from '../store/data/index';
import Vue from 'vue';

export default Vue.extend({
	extends: VariableFacets,

	name: 'results-variable-facets',

	methods: {
		onHistogramClick(context: string, key: string, value: Range) {
			if (key && value) {
				const selectFilter = {
					name: key,
					type: NUMERICAL_FILTER,
					enabled: true,
					min:  value.from,
					max: value.to
				};
				updateResultHighlights(this, context, key, value, selectFilter);
			} else {
				dataMutations.clearFeatureHighlights(this.$store);
			}
		},

		onFacetClick(context: string, key: string, value: string) {
			if (key && value) {
				// extract the var name from the key
				const selectFilter = {
					name: key,
					type: CATEGORICAL_FILTER,
					enabled: true,
					categories: [value]
				};
				updateResultHighlights(this, context, key, value, selectFilter);
			} else {
				// clear existing highlights
				dataMutations.clearFeatureHighlights(this.$store);
			}
		},
	}
});

</script>
