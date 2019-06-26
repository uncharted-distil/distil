<template>

	<div class="container-fluid d-flex flex-column h-100">
		<div class="row flex-0-nav"></div>

		<div class="row justify-content-center h-100">
			<div class="col-12 col-md-10 flex-column d-flex h-100">

				<div class="row mt-1 mb-1" v-for="(idCol, index) in idCols" :key="idCol.value">

					<div class="col-3">
						<template v-if="index===0">
							<b>Group ID:</b>
						</template>
					</div>

					<div class="col-5">
						<b-form-select v-model="idCol.value" :options="idOptions(idCol.value)" @input="onIdChange"/>
					</div>

					<div class="col-4">
						<b-form-checkbox button v-model="hideIdCol[index]">
							Hide Column
						</b-form-checkbox>
					</div>
				</div>

				<div class="row mt-1 mb-1">

					<div class="col-3">
						<b>Group Type:</b>
					</div>
					<div class="col-5">
						<b-form-select v-model="groupingType" :options="typeOptions"/>
					</div>
				</div>

				<div v-if="groupingType==='timeseries'">
					<div class="row justify-content-center mt-3 mb-3">
						<b>Timeseries Grouping</b>
					</div>

					<div class="row mt-1 mb-1">
						<div class="col-3">
							<b>X-Axis:</b>
						</div>

						<div class="col-5">
							<b-form-select v-model="xCol" :options="xColOptions" />
						</div>

						<div class="col-4">
							<b-form-checkbox button v-model="hideXCol">
								Hide Column
							</b-form-checkbox>
						</div>
					</div>

					<div class="row mt-1 mb-1">
						<div class="col-3">
							<b>Y-Axis:</b>
						</div>

						<div class="col-5">
							<b-form-select v-model="yCol" :options="yColOptions" />
						</div>

						<div class="col-4">
							<b-form-checkbox button v-model="hideYCol">
								Hide Column
							</b-form-checkbox>
						</div>
					</div>

					<div class="row mt-1 mb-1">
						<div class="col-3">
							<b>Featurize:</b>
						</div>

						<div class="col-5">
							<b-form-select v-model="clusterCol" :options="clusterColOptions" />
						</div>

						<div class="col-4">
							<b-form-checkbox button v-model="hideClusterCol">
								Hide Column
							</b-form-checkbox>
						</div>
					</div>

				</div>

				<div v-if="isReady" class="row justify-content-center">
					<b-btn class="mt-3 var-grouping-button" variant="outline-success" @click="onGroup">
						<div class="row justify-content-center">
							<i class="fa fa-check-circle fa-2x mr-2"></i>
							<b>Submit</b>
						</div>
					</b-btn>
					<b-btn class="mt-3 var-grouping-button" variant="outline-danger" @click="onClose">
						<div class="row justify-content-center">
							<i class="fa fa-times-circle fa-2x mr-2"></i>
							<b>Cancel</b>
						</div>
					</b-btn>
				</div>
			</div>
		</div>

	</div>

</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import { Variable } from '../store/dataset/index';
import { getters as datasetGetters, actions as datasetActions } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';
import { actions as viewActions } from '../store/view/module';
import { INTEGER_TYPE, TEXT_TYPE, ORDINAL_TYPE, TIMESTAMP_TYPE, CATEGORICAL_TYPE,
	DATE_TIME_TYPE, REAL_TYPE } from '../util/types';

export default Vue.extend({
	name: 'variable-grouping',

	components: {
	},

	data() {
		return {
			groupingType: null,
			idCols: [ { value: null } ],
			prevIdCols: 0,
			xCol: null,
			yCol: null,
			clusterCol: null,
			hideIdCol: [ true ],
			hideXCol: true,
			hideYCol: true,
			hideClusterCol: true,
			other: []
		};
	},
	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		variables(): Variable[] {
			return datasetGetters.getVariables(this.$store);
		},
		typeOptions(): Object[] {
			return [
				{ value: null, text: 'Choose group type', disabled: true },
				{ value: 'timeseries', text: 'Timeseries' }
			];
		},
		xColOptions(): Object[] {
			const X_COL_TYPES = {
				[INTEGER_TYPE]: true,
				[DATE_TIME_TYPE]: true,
				[TIMESTAMP_TYPE]: true
			};
			const def = [
				{ value: null, text: 'Choose column', disabled: true }
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
				{ value: null, text: 'Choose column', disabled: true }
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

		clusterColOptions(): Object[] {
			const def = [
				{ value: null, text: 'Choose column', disabled: true }
			];
			const suggestions = this.variables
				.filter(v => !this.isIDCol(v.colName))
				.filter(v => !this.isXCol(v.colName))
				.filter(v => !this.isYCol(v.colName))
				.map(v => {
					return {value: v.colName, text: v.colDisplayName };
				});

			// if (suggestions.length === 1) {
			// 	this.clusterCol = suggestions[0].value;
			// 	return suggestions;
			// }

			return [].concat(def, suggestions);
		},

		isReady(): boolean {
			return this.idCols.length > 1 && this.xCol && this.yCol && this.groupingType;
		}
	},

	beforeMount() {
		viewActions.fetchSelectTargetData(this.$store, false);
	},

	methods: {
		idOptions(idCol): Object[] {
			const ID_COL_TYPES = {
				[INTEGER_TYPE]: true,
				[TEXT_TYPE]: true,
				[ORDINAL_TYPE]: true,
				[CATEGORICAL_TYPE]: true
			};
			const def = [
				{ value: null, text: 'Choose ID', disabled: true }
			];
			const suggestions = this.variables
				.filter(v => ID_COL_TYPES[v.colType])
				.filter(v => v.colName === idCol || !this.isIDCol(v.colName))
				.filter(v => !this.isXCol(v.colName))
				.filter(v => !this.isYCol(v.colName))
				.map(v => {
					return { value: v.colName, text: v.colDisplayName };
				});

			return [].concat(def, suggestions);
		},
		onIdChange(arg) {
			const values = this.idCols.map(c => c.value).filter(v => v);
			if (values.length === this.prevIdCols) {
				return;
			}
			this.idCols.push({ value: null });
			this.prevIdCols++;
		},
		isIDCol(arg): boolean {
			return !!this.idCols.find(id => id.value === arg);
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
			const hidden = {
				// [this.idCol]: this.hideIdCol,
				[this.xCol]: this.hideXCol,
				[this.yCol]: this.hideYCol,
				[this.clusterCol]: this.hideClusterCol
			};
			this.idCols.forEach((id, index) => {
				if (id) {
					hidden[id.value] = this.hideIdCol[index];
				}
			});

			let idKey = '';
			let p = null;
			const ids = this.idCols.map(c => c.value).filter(v => v);
			if (ids.length > 1) {
				idKey = ids.join('-');
				hidden[idKey] = true;
				p = datasetActions.composeVariables(this.$store, {
					dataset: this.dataset,
					key: idKey,
					vars: ids
				});
			} else {
				idKey = ids[0];
				p = new Promise(resolve => {
					resolve();
				});
			}

			return p.then(() => {
				const grouping =  {
					type: this.groupingType,
					dataset: this.dataset,
					idCol: idKey,
					hidden: Object.keys(hidden).filter(v => hidden[v]),
					properties: {
						xCol: this.xCol,
						yCol: this.yCol,
						clusterCol: this.clusterCol
					}
				};
				datasetActions.setGrouping(this.$store, {
					dataset: this.dataset,
					grouping: grouping
				});
				this.$emit('close');
			});
		},
		onClose() {
			this.$emit('close');
		}
	}

});
</script>

<style>
.var-grouping-button {
	margin: 0 8px;
	width: 25% !important;
	line-height: 32px !important;
}
</style>
