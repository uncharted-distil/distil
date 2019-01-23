<template>
	<div>
		<b-modal
			v-model="show"
			cancel-disabled
			hide-header
			hide-footer>
			<div class="row justify-content-center">
				<i class="fa fa-exclamation-triangle fa-3x fail-icon"></i>
				<div><b>{{title}}:</b> {{error}}</div>
			</div>
			<div class="row justify-content-center">
				<b-btn class="mt-3 join-modal-button" variant="outline-secondary" block @click="onClose">OK</b-btn>
			</div>
		</b-modal>
	</div>

</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import { createRouteEntry } from '../util/routes';
import { formatBytes } from '../util/bytes';
import { sortVariablesByImportance } from '../util/data';
import { getters as routeGetters } from '../store/route/module';
import { Dataset, Variable } from '../store/dataset/index';
import { actions as datasetActions } from '../store/dataset/module';
import { SELECT_TARGET_ROUTE } from '../store/route/index';
import localStorage from 'store';

const NUM_TOP_FEATURES = 5;

export default Vue.extend({
	name: 'dataset-preview',

	props: {
		title: {
			type: String as () => string,
			default: 'Error'
		},
		error: {
			type: String as () => string,
			default: 'Internal server error'
		},
		show: Boolean as () => boolean
	},

	methods: {
		onClose() {
			this.$emit('close');
		}
	}

});
</script>

<style>
</style>
