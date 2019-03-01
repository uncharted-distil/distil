<template>

	<div>
		<b-modal v-model="show">
			<div class="row justify-content-center mb-3">
				<b>Build Variable Grouping</b>
			</div>
			<div class="row justify-content-center">
				<b-form-select v-model="idCol" :options="idOptions"/>
				<b-form-select v-model="groupingType" :options="typeOptions"/>
			</div>
			<div v-if="groupingType==='timeseries'">
				<div class="row justify-content-center mt-3 mb-3">
					<b>Timeseries Grouping</b>
				</div>
				<div class="row justify-content-center">
					<div class="ml-3 mr-3">
						<b>X-Axis:</b>
					</div>
					<b-form-select v-model="xCol" :options="xColOptions" />
					<div class="ml-3 mr-3">
						<b>Y-Axis:</b>
					</div>
					<b-form-select v-model="yCol" :options="yColOptions" />
				</div>
			</div>

			<div v-if="isReady" class="row justify-content-center">
				<b-btn class="mt-3 grouping-modal-button" variant="outline-success" @click="onGroup">
					<div class="row justify-content-center">
						<i class="fa fa-check-circle fa-2x mr-2"></i>
						<b>Submit</b>
					</div>
				</b-btn>
				<b-btn class="mt-3 grouping-modal-button" variant="outline-danger" @click="onClose">
					<div class="row justify-content-center">
						<i class="fa fa-times-circle fa-2x mr-2"></i>
						<b>Cancel</b>
					</div>
				</b-btn>
			</div>
		</b-modal>
	</div>

</template>

<script lang="ts">

import Vue from 'vue';
import { Variable } from '../store/dataset/index';
import { getters as datasetGetters } from '../store/dataset/module';
import { INTEGER_TYPE, TEXT_TYPE, ORDINAL_TYPE, CATEGORICAL_TYPE,
	DATE_TIME_TYPE, REAL_TYPE } from '../util/types';

export default Vue.extend({
	name: 'group-model',

	props: {
		show: Boolean as () => boolean
	},

	data() {
		return {
			groupingType: null,
			idCol: null,
			xCol: null,
			yCol: null,
			other: []
		};
	},
	computed: {
		variables(): Variable[] {
			return datasetGetters.getVariables(this.$store);
		},
		idOptions(): Object[] {
			const ID_COL_TYPES = {
				[INTEGER_TYPE]: true,
				[TEXT_TYPE]: true,
				[ORDINAL_TYPE]: true,
				[CATEGORICAL_TYPE]: true
			};
			const def = [
				{ value: null, text: 'Choose id to group under' }
			];
			const suggestions = this.variables
				.filter(v => ID_COL_TYPES[v.colType])
				.filter(v => !this.isXCol(v.colName))
				.filter(v => !this.isYCol(v.colName))
				.map(v => {
					return {value: v.colName, text: v.colDisplayName };
				});

			return [].concat(def, suggestions);
		},
		typeOptions(): Object[] {
			return [
				{ value: null, text: 'Choose a group type' },
				{ value: 'timeseries', text: 'Timeseries' },
			];
		},
		xColOptions(): Object[] {
			const X_COL_TYPES = {
				[INTEGER_TYPE]: true,
				[DATE_TIME_TYPE]: true,
				// [TIMESTAMP_TYPE]: true
			};
			const def = [
				{ value: null, text: 'Choose x column' }
			];
			const suggestions = this.variables
				.filter(v => X_COL_TYPES[v.colType])
				.filter(v => !this.isIDCol(v.colName))
				.filter(v => !this.isYCol(v.colName))
				.map(v => {
					return {value: v.colName, text: v.colDisplayName };
				});

			if (suggestions.length === 1) {
				this.xCol = suggestions[0].value;
				return suggestions;
			}

			return [].concat(def, suggestions);
		},

		yColOptions(): Object[] {
			const Y_COL_TYPES = {
				[INTEGER_TYPE]: true,
				[REAL_TYPE]: true
			};
			const def = [
				{ value: null, text: 'Choose y column' }
			];
			const suggestions = this.variables
				.filter(v => Y_COL_TYPES[v.colType])
				.filter(v => !this.isIDCol(v.colName))
				.filter(v => !this.isXCol(v.colName))
				.map(v => {
					return {value: v.colName, text: v.colDisplayName };
				});

			if (suggestions.length === 1) {
				this.yCol = suggestions[0].value;
				return suggestions;
			}

			return [].concat(def, suggestions);
		},

		otherOptions(): Object[] {
			const def = [
				{ value: null, text: 'Choose other variables' }
			];
			const suggestions = this.variables
				.filter(v => !this.isIDCol(v.colName))
				.filter(v => !this.isXCol(v.colName))
				.filter(v => !this.isYCol(v.colName))
				.filter(v => !this.isOtherCol(v.colName))
				.map(v => {
					return {value: v.colName, text: v.colDisplayName };
				});

			return [].concat(def, suggestions);
		},

		isReady(): boolean {
			return this.idCol && this.xCol && this.yCol && this.groupingType;
		}
	},

	methods: {
		isIDCol(arg): boolean {
			return arg === this.idCol;
		},
		isXCol(arg): boolean {
			return arg === this.xCol;
		},
		isYCol(arg): boolean {
			return arg === this.yCol;
		},
		isOtherCol(arg): boolean {
			return this.other.indexOf(arg) !== -1;
		},
		onGroup() {
			console.log('submit grouping req');
			this.$emit('close');
		},
		onClose() {
			this.$emit('close');
		}
	}

});

</script>

<style>
.grouping-modal-button {
	margin: 0 8px;
	width: 25% !important;
	line-height: 32px !important;
}
</style>
