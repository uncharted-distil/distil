<script lang="ts">

import VariableFacets from './VariableFacets.vue';
import { mutations as dataMutations } from '../store/data/module';
import { NUMERICAL_FILTER, CATEGORICAL_FILTER } from '../util/filters';
import { updateResultHighlights } from '../util/highlights';
import Vue from 'vue';
import _ from 'lodash';

const RESULT_VARIABLE_FACET_HIGHLIGHTS = 'result-variable-facet';

export default Vue.extend({
	extends: VariableFacets,

	name: 'results-variable-facets',

	methods: {
		onHistogramClick(key: string, value: any) {
			// on histogram click event, publish the highlight/clear highlight to the
			// rest of the app
			dataMutations.clearFeatureHighlights(this.$store);

			if (key && value) {
				const selectFilter = {
					name: key,
					type: NUMERICAL_FILTER,
					enabled: true,
					min:  _.toNumber(value.label[0]),
					max: _.toNumber(value.toLabel[value.toLabel.length-1])
				};
				updateResultHighlights(this, key, selectFilter, RESULT_VARIABLE_FACET_HIGHLIGHTS);
			}
		},

		onFacetClick(key: string, value: string) {
			// clear existing highlights
			dataMutations.clearFeatureHighlights(this.$store);

			if (key && value) {
				// extract the var name from the key
				const selectFilter = {
					name: key,
					type: CATEGORICAL_FILTER,
					enabled: true,
					categories: [value]
				};
				updateResultHighlights(this, key, selectFilter, RESULT_VARIABLE_FACET_HIGHLIGHTS);
			}
		},
	}
});

</script>
