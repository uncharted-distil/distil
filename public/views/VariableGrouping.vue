<template>

	<div class="container-fluid d-flex flex-column h-100">
		<div class="row flex-0-nav"></div>

		<div class="row flex-shrink-0 align-items-center bg-white">
			<div class="col-4 offset-md-1">
				<h5 class="header-label">Configure Time Series</h5>
			</div>
		</div>

		<div class="row justify-content-center h-100 p-3">

			<div class="col-12 col-md-8 flex-column d-flex h-100">

				<div class="row mt-1 mb-1" v-for="(idCol, index) in idCols" :key="idCol.value">

					<div class="col-3">
						<template v-if="index===0">
							<b>Series ID Column(s):</b>
						</template>
					</div>

					<div class="col-5">
						<b-form-select v-model="idCol.value" :options="idOptions(idCol.value)" @input="onIdChange"/>
					</div>
				</div>

				<div class="row mt-1 mb-1">
					<div class="col-3">
						<b>Time Column:</b>
					</div>

					<div class="col-5">
						<b-form-select v-model="xCol" :options="xColOptions" />
					</div>
				</div>

				<div class="row mt-1 mb-1">
					<div class="col-3">
						<b>Value Column:</b>
					</div>

					<div class="col-5">
						<b-form-select v-model="yCol" :options="yColOptions" />
					</div>
				</div>

				<div v-if="isReady"  class="row justify-content-center">
					<b-btn class="mt-3 var-grouping-button" variant="outline-success" :disabled="isPending" @click="onGroup">
						<div class="row justify-content-center">
							<i class="fa fa-check-circle fa-2x mr-2"></i>
							<b>Submit</b>
						</div>
					</b-btn>
					<b-btn class="mt-3 var-grouping-button" variant="outline-danger" :disabled="isPending" @click="onClose">
						<div class="row justify-content-center">
							<i class="fa fa-times-circle fa-2x mr-2"></i>
							<b>Cancel</b>
						</div>
					</b-btn>
				</div>

				<div class="grouping-progress">
					<b-progress v-if="isPending"
						:value="percentComplete"
						variant="outline-secondary"
						striped
						:animated="true"></b-progress>
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
import { getComposedVariableKey } from '../util/data';
import { SELECT_TARGET_ROUTE } from '../store/route/index';
import { createRouteEntry, overlayRouteEntry } from '../util/routes';

export default Vue.extend({
	name: 'variable-grouping',

	components: {
	},

	data() {
		return {
			groupingType: 'timeseries',
			idCols: [ { value: null } ],
			prevIdCols: 0,
			xCol: null,
			yCol: null,
			hideIdCol: [ false ],
			hideXCol: true,
			hideYCol: true,
			hideClusterCol: true,
			other: [],
			isPending: false,
			percentComplete: 100
		};
	},
	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		variables(): Variable[] {
			return datasetGetters.getVariables(this.$store);
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

		isReady(): boolean {
			return this.xCol !== null && this.groupingType !== null;
			// return this.idCols.length > 1 && this.xCol && this.yCol && this.groupingType;
		},

		isTimeseriesAnalysis(): boolean {
			const ids = this.idCols.filter(id => !!id.value);
			console.log(this.xCol, ids.length, this.yCol);
			return this.xCol && (ids.length === 0) && !this.yCol;
		}
	},

	beforeMount() {
		viewActions.fetchSelectTargetData(this.$store, false);
	},
	methods: {
		idOptions(idCol): Object[] {
			const ID_COL_TYPES = {
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
			this.hideIdCol.push(false);
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
			console.log(this.isTimeseriesAnalysis, 'isTimeseriesAnalysis');
			if (this.isTimeseriesAnalysis) {
				this.submitTimeseriesAnalysis();
			} else {
				this.submitGrouping();
			}
		},
		submitGrouping() {
			const hidden = {
				[this.xCol]: true,
				[this.yCol]: true
			};
			const ids = this.idCols.map(c => c.value).filter(v => v);

			ids.forEach((id, index) => {
				hidden[id] = this.hideIdCol[index];
			});

			let idKey = '';
			let p = null;
			if (ids.length > 1) {
				idKey = getComposedVariableKey(ids);
				hidden[idKey] = false;
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

			this.isPending = true;

			return p.then(() => {
				const grouping =  {
					type: this.groupingType,
					dataset: this.dataset,
					idCol: idKey,
					subIds: ids,
					hidden: Object.keys(hidden).filter(v => hidden[v]),
					properties: {
						xCol: this.xCol,
						yCol: this.yCol,
					}
				};
				datasetActions.setGrouping(this.$store, {
					dataset: this.dataset,
					grouping: grouping
				}).then(() => {
					this.gotoTargetSelection();
				});
			});
		},
		submitTimeseriesAnalysis() {
			const entry = createRouteEntry(SELECT_TARGET_ROUTE, {
				dataset: this.dataset,
				timeseriesAnalysis: this.xCol
			});
			this.$router.push(entry);
		},
		onClose() {
			this.gotoTargetSelection();
		},
		gotoTargetSelection() {
			const entry = createRouteEntry(SELECT_TARGET_ROUTE, {
				dataset: this.dataset
			});
			this.$router.push(entry);
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
.grouping-progress {
	margin: 6px 10%;
}
</style>