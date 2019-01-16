<template>
	<div class='card card-result'>
		<div class='dataset-card-header hover card-header'  variant="dark">
			<b>{{name}}</b>
		</div>
		<div class='card-body'>
			<div class='row align-items-center justify-content-center'>
				<div class='col-6'>
					<div><b>Features:</b> {{variables.length}}</div>
					<div><b>Rows:</b> {{numRows}}</div>
					<div><b>Size:</b> {{formatBytes(numBytes)}}</div>
				</div>
				<!-- <div class='col-6'>
					<span><b>Top features:</b></span>
					<ul>
						<li :key="variable.name" v-for='variable in topVariables'>
							{{variable.colDisplayName}}
						</li>
					</ul>
				</div> -->
			</div>
		</div>
	</div>

</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import { sortVariablesByImportance } from '../util/data';
import { formatBytes } from '../util/bytes';
import { Variable } from '../store/dataset/index';


const NUM_TOP_FEATURES = 5;

export default Vue.extend({
	name: 'dataset-preview-card',

	props: {
		name: String as () => string,
		variables: Array as () => Variable[],
		numRows: Number as () => number,
		numBytes: Number as () => number
	},

	computed: {
		topVariables(): Variable[] {
			return sortVariablesByImportance(this.variables.slice(0)).slice(0, NUM_TOP_FEATURES);
		}
	},

	methods: {
		formatBytes(n: number): string {
			return formatBytes(n);
		}
	}
});
</script>

<style>
.dataset-card-header {
	display: flex;
	padding: 4px 8px;
	color: white;
	justify-content: space-between;
	border: none;
}
.card-result .card-header {
	background-color: #424242;
}

</style>
