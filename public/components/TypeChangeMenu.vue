<template>
	<div class="enable-type-change-menu">
		<b-dropdown :text="type" variant="secondary" class="var-type-button">
			<b-dropdown-item
				v-for="suggested in addMissingSuggestions()"
				@click.stop="onTypeChange(suggested)"
				:key="suggested">
				{{suggested}}
			</b-dropdown-item>
		</b-dropdown>
	</div>
</template>

<script lang="ts">

import { actions as datasetActions, getters as datasetGetters } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { addTypeSuggestions, getLabelFromType, getTypeFromLabel } from '../util/types';
import _ from 'lodash';
import Vue from 'vue';

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
