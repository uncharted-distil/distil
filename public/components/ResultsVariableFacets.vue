<script lang="ts">

import VariableFacets from './VariableFacets.vue';
import { updateHighlightRoot, clearHighlightRoot, Range } from '../util/highlights';
import Vue from 'vue';

export default Vue.extend({
	extends: VariableFacets,

	name: 'results-variable-facets',

	methods: {
		onHistogramClick(context: string, key: string, value: Range) {
			if (key && value) {
				updateHighlightRoot(this, {
					context: context,
					key: key,
					value: value
				});
			} else {
				clearHighlightRoot(this);
			}
		},

		onFacetClick(context: string, key: string, value: string) {
			if (key && value) {
				updateHighlightRoot(this, {
					context: context,
					key: key,
					value: value
				});
			} else {
				// clear existing highlights
				clearHighlightRoot(this);
			}
		},
	}
});

</script>
