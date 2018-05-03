<template>
	<div class="error-threshold-slider" v-if="showSlider">

		<div>
			<div class="error-header">
				Error:
				<div class="asym-button float-right" v-bind:class="{ active: !symmetricSlider }" @click="enableAsymmetric">
					<div class="button-line"></div>
					<div class="button-center"></div>
					<div class="button-left-circle"></div>
				</div>
				<div class="sym-button float-right" v-bind:class="{ active: symmetricSlider }" @click="enableSymmetric">
					<div class="button-line"></div>
					<div class="button-center"></div>
					<div class="button-left-circle"></div>
					<div class="button-right-circle"></div>
				</div>
			</div>
		</div>

		<div class="error-slider">
			<div class="error-center-line"></div>
			<div class="error-center-label">0</div>
			<vue-slider ref="slider"
				:min="residualExtrema.min"
				:max="residualExtrema.max"
				:interval="interval"
				:value="initialValue"
				:formatter="formatter"
				:lazy="true"
				width=100%
				tooltip-dir="bottom"
				@callback="onSlide"/>
		</div>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import { overlayRouteEntry } from '../util/routes';
import { Extrema } from '../store/data/index';
import { getters as dataGetters} from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import vueSlider from 'vue-slider-component';
import Vue from 'vue';
import 'font-awesome/css/font-awesome.css';

const DEFAULT_PERCENTILE = 0.25;
const NUM_STEPS = 100;
const ERROR_DECIMALS = 0;

export default Vue.extend({
	name: 'error-threshold-slider',

	components: {
		vueSlider,
	},

	data() {
		return {
			formatter(arg) {
				return arg ? arg.toFixed(2) : '';
			},
			symmetricSlider: true,
			min: NaN,
			max: NaN
		};
	},

	computed: {

		showSlider(): boolean {
			return !_.isNaN(this.interval);
		},

		initialValue(): number[] {
			const min = routeGetters.getRouteResidualThresholdMin(this.$store);
			const max = routeGetters.getRouteResidualThresholdMax(this.$store);
			if (min === undefined || max === undefined) {
				if (!_.isNaN(this.defaultValue[0]) && !_.isNaN(this.defaultValue[1])) {
					this.updateThreshold(this.defaultValue[0], this.defaultValue[1]);
				}
				return this.defaultValue;
			}
			const nmin = _.toNumber(min);
			const nmax = _.toNumber(max);
			// NOTE: the slider component discards the values if they are
			// not within the extrema. We have to read the extrema here so
			// that the values are re-computed when the extrema is computed.
			const extrema = this.residualExtrema;
			if (nmin < extrema.min || nmax > extrema.max) {
				return [ NaN, NaN ];
			}
			return [
				nmin,
				nmax
			];
		},

		range(): number {
			if (_.isNaN(this.residualExtrema.min) ||
				_.isNaN(this.residualExtrema.max)) {
				return NaN;
			}
			return this.residualExtrema.max - this.residualExtrema.min;
		},

		defaultValue(): number[] {
			return [
				-this.range/2 * DEFAULT_PERCENTILE,
				this.range/2 * DEFAULT_PERCENTILE
			];
		},

		interval(): number {
			return this.range / NUM_STEPS;
		},

		residualExtrema(): Extrema {
			const extrema = dataGetters.getResidualExtrema(this.$store);
			if (!extrema) {
				return extrema;
			}
			return {
				min: _.round(extrema.min, ERROR_DECIMALS),
				max: _.round(extrema.max, ERROR_DECIMALS)
			};
		}
	},

	methods: {

		enableAsymmetric() {
			this.symmetricSlider = false;
		},

		enableSymmetric() {
			this.symmetricSlider = true;
			const newVal = Math.min(Math.abs(this.min), this.max);
			this.forceSymmetric([ -newVal, newVal ]);
		},

		forceSymmetric(value: number[]): number[] {
			const newValues = [ value[0], value[1] ];
			if (this.symmetricSlider) {
				if (value[0] !== this.min) {
					// min changed
					newValues[1] = -value[0];
				}
				if (value[1] !== this.max) {
					// max changed
					newValues[0] = -value[1];
				}
				const $slider = <any>this.$refs.slider;
				$slider.setValue(newValues, true);
			}
			return newValues;
		},

		updateThreshold(min: number, max: number) {
			this.min = min;
			this.max = max;
			const entry = overlayRouteEntry(this.$route, {
				residualThresholdMin: `${min}`,
				residualThresholdMax: `${max}`
			});
			this.$router.push(entry);
		},

		onSlide(value: number[]) {
			const newValues = this.forceSymmetric(value);
			this.updateThreshold(newValues[0], newValues[1]);
		}

	}
});
</script>

<style>

.error-threshold-slider {
	position: relative;
	width: 100%;
}

.error-header {
	position: relative;
	margin: 0 15px;
}

.error-slider {
	position: relative;
	margin: 8px 10%;
}


.error-threshold-slider .vue-slider-component .vue-slider-process {
	background-color: #9e9e9e;
}

.error-threshold-slider .vue-slider-component .vue-slider-tooltip {
	border: 1px solid #9e9e9e;
	background-color: #9e9e9e;
}

.error-threshold-slider .vue-slider-component .vue-slider-piecewise {
	background-color: #ee0701;
}


.facets-facet-vertical.select-highlight .facet-bar-selected {
	box-shadow: inset 0 0 0 1000px #007bff;
}

.error-center-line {
	position:absolute;
	left: 50%;
	height: 22px;
	width: 1px;
	background-color: #666;
}

.error-center-label {
	position:absolute;
	top: 22px;
	width: 100%;
	color: #666;
	text-align: center;
}

.sym-button,
.asym-button {
	position: relative;
	width: 36px;
	height: 23px;
	border-radius: 4px;
	margin: 2px;
	cursor: pointer;
}

.button-line {
	position: absolute;
	width: 26px;
	height: 2px;
	left: 4px;
	top: 10px;
}

.button-center {
	position: absolute;
	width: 1px;
	height: 14px;
	left: 17px;
	top: 4px;
}

.button-left-circle {
	position: absolute;
	width: 8px;
	height: 8px;
	border-radius: 8px;
	top: 7px;
	left: 3px;
}

.button-right-circle {
	position: absolute;
	width: 8px;
	height: 8px;
	border-radius: 8px;
	top: 7px;
	right: 3px;
}

.sym-button,
.asym-button {
	background-color: #eee;
	border: 1px solid #9e9e9e;
}

.sym-button:hover,
.asym-button:hover {
	opacity: 0.75;
}

.sym-button.active,
.asym-button.active {
	background-color: #9e9e9e;
	border: 1px solid #eee;
}

.button-line,
.button-center,
.button-left-circle,
.button-right-circle {
	background-color: #9e9e9e;
}

.active .button-line,
.active .button-center,
.active .button-left-circle,
.active .button-right-circle {
	background-color: #fff;
}
</style>
