<template>
	<div class="enable-type-change-menu">
		<b-dropdown variant="secondary" class="var-type-button"
			id="type-change-dropdown"
			:text="type"
			:disabled="isDisabled">
			<b-dropdown-item
				v-for="suggested in addMissingSuggestions()"
				@click.stop="onTypeChange(suggested)"
				:key="suggested">
				{{suggested}}
			</b-dropdown-item>
		</b-dropdown>
		<b-tooltip :delay="delay" :disabled="!isDisabled" target="type-change-dropdown">
			Cannot change type when actively filtering
		</b-tooltip>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import { HighlightRoot } from '../store/highlights/index';
import { actions as datasetActions, getters as datasetGetters } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';
import { addTypeSuggestions, getLabelFromType, getTypeFromLabel } from '../util/types';
import { hasFilterInRoute } from '../util/filters';

export default Vue.extend({
	name: 'enable-type-change-menu',

	props: {
		field: String,
		values: Array
	},

	computed: {
		type(): string {
			const vars = datasetGetters.getVariablesMap(this.$store);
			if (!vars || !vars[this.field.toLowerCase()]) {
				return '';
			}
			const type = vars[this.field.toLowerCase()].type;
			return getLabelFromType(type);
		},
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		highlightRoot(): HighlightRoot {
			return routeGetters.getDecodedHighlightRoot(this.$store);
		},
		isDisabled(): boolean {
			return hasFilterInRoute(this.field) || (this.highlightRoot && this.highlightRoot.key === this.field);
		},
		delay(): any {
			return {
				show: 10,
				hide: 10
			};
		}
	},

	methods: {
		addMissingSuggestions(): string[] {
			const type = getTypeFromLabel(this.type);
			return _.map(addTypeSuggestions(type, this.values), t => getLabelFromType(t));
		},
		onTypeChange(suggested) {
			const type = getTypeFromLabel(suggested);
			datasetActions.setVariableType(this.$store, {
				dataset: this.dataset,
				field: this.field,
				type: type
			});
		},
	}
});
</script>

<style>
.var-type-button button {
	border: none;
	border-radius: 0;
	padding: 2px 4px;
	width: 100%;
	text-align: left;
	outline: none;
	font-size: 0.750rem;
	color: white;
}
.var-type-button button:hover,
.var-type-button button:active,
.var-type-button button:focus,
.var-type-button.show > .dropdown-toggle  {
	border: none;
	border-radius: 0;
	padding: 2px 4px;
	color: white;
	background-color: #424242;
	border-color: #424242;
	box-shadow: none;
}
.enable-type-change-menu .dropdown-item {
	font-size: 0.867rem;
	text-transform: none;
}
</style>
