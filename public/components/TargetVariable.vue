<template>
	<div>
		<h6 class="nav-link">Target Feature</h6>
		<div class="target-no-target" v-if="variables.length===0">
			<div class="text-danger">
				<i class="fa fa-times missing-icon"></i><strong>No Target Feature Selected</strong>
			</div>
		</div>
		<variable-facets v-if="variables.length>0"
			:variables="variables"
			:dataset="dataset"
			:html="html"></variable-facets>
	</div>
</template>

<script>

import { createRouteEntryFromRoute } from '../util/routes';
import VariableFacets from '../components/VariableFacets';
import 'font-awesome/css/font-awesome.css';

export default {
	name: 'target-variables',

	components: {
		VariableFacets
	},

	computed: {
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		variables() {
			return this.$store.getters.getTargetVariableSummaries();
		},
		html() {
			return () => {
				const container = document.createElement('div');
				const remove = document.createElement('button');
				remove.className += 'btn btn-sm btn-outline-danger mb-2';
				remove.innerHTML = 'Remove';
				remove.addEventListener('click', () => {
					const entry = createRouteEntryFromRoute(this.$store.getters.getRoute(), {
						target: ''
					});
					this.$router.push(entry);
				});
				container.appendChild(remove);
				return container;
			};
		}
	}
};
</script>

<style>
.target-no-target {
	width: 100%;
	background-color: #eee;
	padding: 8px;
}
.missing-icon {
	padding-right: 4px;
}
</style>
