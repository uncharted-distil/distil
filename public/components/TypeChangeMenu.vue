<template>
	<div class="enable-type-change-menu">
		<i v-if="isUnsure" class="unsure-type-icon fa fa-exclamation"></i>
		<b-dropdown variant="secondary" class="var-type-button"
			id="type-change-dropdown"
			:text="label"
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
import { SuggestedType, Variable } from '../store/dataset/index';
import { HighlightRoot } from '../store/highlights/index';
import { actions as datasetActions, getters as datasetGetters } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';
import { addTypeSuggestions, getLabelFromType, getTypeFromLabel, isEquivalentType, BASIC_SUGGESTIONS } from '../util/types';
import { hasFilterInRoute } from '../util/filters';

const PROBABILITY_THRESHOLD = 0.8;

export default Vue.extend({
	name: 'enable-type-change-menu',

	props: {
		field: String as () => string,
		values: Array as () => Array<any>,
	},

	computed: {
		variable(): Variable {
			const vars = datasetGetters.getVariablesMap(this.$store);
			if (!vars || !vars[this.field.toLowerCase()]) {
				return null;
			}
			return vars[this.field.toLowerCase()];
		},
		type(): string {
			return this.variable ? this.variable.type : '';
		},
		label(): string {
			return this.type !== '' ? getLabelFromType(this.type) : '';
		},
		originalType(): string {
			return this.variable ? this.variable.originalType : '';
		},
		suggestedTypes(): SuggestedType[] {
			return this.variable ? this.variable.suggestedTypes : [];
		},
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},
		highlightRoot(): HighlightRoot {
			return routeGetters.getDecodedHighlightRoot(this.$store);
		},
		isDisabled(): boolean {
			return hasFilterInRoute(this.field) || (this.highlightRoot && this.highlightRoot.key === this.field);
		},
		hasSchemaType(): boolean {
			return !!this.schemaType;
		},
		hasNonSchemaTypes(): boolean {
			return _.find(this.suggestedTypes, t => {
				return t.provenance !== 'schema';
			}) !== undefined;
		},
		schemaType(): SuggestedType {
			return _.find(this.suggestedTypes, t => {
				return t.provenance === 'schema';
			});
		},
		topNonSchemaType(): SuggestedType {
			const nonSchemaTypes = _.filter(this.suggestedTypes, t => {
				return t.provenance !== 'schema';
			});
			nonSchemaTypes.sort((a: any, b: any) => {
				return b.probability - a.probability;
			});
			return nonSchemaTypes.length > 0 ? nonSchemaTypes[0] : undefined;
		},
		isUnsure(): boolean {
			return (this.type === this.originalType && // we haven't changed the type
				this.hasSchemaType && this.hasNonSchemaTypes &&
				this.topNonSchemaType.probability >= PROBABILITY_THRESHOLD && // it has both schema and ML types
				!isEquivalentType(this.schemaType.type, this.topNonSchemaType.type)); // they don't agree
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
			if (this.label === '' || this.values.length === 0) {
				return _.map(BASIC_SUGGESTIONS, t => getLabelFromType(t));
			}
			const type = getTypeFromLabel(this.label);
			return _.map(addTypeSuggestions(type, this.values), t => getLabelFromType(t));
		},
		onTypeChange(suggested) {
			const type = getTypeFromLabel(suggested);
			datasetActions.setVariableType(this.$store, {
				dataset: this.dataset,
				field: this.field,
				type: type
			}).then(() => {
				if (this.target) {
					datasetActions.fetchVariableRankings(this.$store, {
						dataset: this.dataset,
						target: this.target
					});
				}
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
.unsure-type-icon {
	color: #dc3545;
}
</style>
