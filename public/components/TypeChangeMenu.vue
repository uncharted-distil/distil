<template>
	<div class="type-change-menu">
		<b-dropdown variant="secondary" class="var-type-button"
			id="type-change-dropdown"
			:text="label"
			:disabled="isDisabled">
			<b-dropdown-item
				v-for="suggested in getSuggestedList()"
				v-bind:class="{ selected: suggested.isSelected, recommended: suggested.isRecommended }"
				@click.stop="onTypeChange(suggested.type)"
				:key="suggested.type">
				<i v-if="suggested.isSelected" class="fa fa-check" aria-hidden="true"></i>
				{{suggested.label}}
				<img v-if="suggested.isRecommended" src="/images/recommended.svg" class="recommended-icon"/>
			</b-dropdown-item>
		</b-dropdown>
		<i v-if="isUnsure" class="unsure-type-icon fa fa-circle"></i>
		<b-tooltip :delay="delay" :disabled="!isDisabled" target="type-change-dropdown">
			Cannot change type when actively filtering
		</b-tooltip>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import '../assets/images/recommended.svg';
import { SuggestedType, Variable } from '../store/dataset/index';
import { HighlightRoot } from '../store/highlights/index';
import { actions as datasetActions, getters as datasetGetters } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';
import { addTypeSuggestions, getLabelFromType, getTypeFromLabel, isEquivalentType, isLocationType, BASIC_SUGGESTIONS } from '../util/types';
import { hasFilterInRoute } from '../util/filters';

const PROBABILITY_THRESHOLD = 0.8;

export default Vue.extend({
	name: 'type-change-menu',

	props: {
		dataset: String as () => string,
		field: String as () => string,
		values: Array as () => any[],
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
			return this.variable ? this.variable.colType : '';
		},
		isColTypeChanged(): boolean {
			return this.variable ? this.variable.isColTypeChanged : false;
		},
		label(): string {
			return this.type !== '' ? getLabelFromType(this.type) : '';
		},
		originalType(): string {
			return this.variable ? this.variable.colOriginalType : '';
		},
		suggestedTypes(): SuggestedType[] {
			const suggestedType = this.variable ? this.variable.suggestedTypes : [];
			return _.orderBy(suggestedType, 'probability' , 'desc');
		},
		sggestedNonSchemaTypes(): SuggestedType[] {
			const nonSchemaTypes = _.filter(this.suggestedTypes, t => {
				return t.provenance !== 'schema';
			});
			return nonSchemaTypes;
		},
		topNonSchemaType(): SuggestedType {
			return this.sggestedNonSchemaTypes.length > 0 ? this.sggestedNonSchemaTypes[0] : undefined;
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
		isUnsure(): boolean {
			return (this.type === this.originalType && // we haven't changed the type (check from server)
				!this.isColTypeChanged && // check if user ever changed the col type (client)
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
		addMissingSuggestions() {
			if (this.sggestedNonSchemaTypes.length === 0 && (this.label === '' || this.values.length === 0)) {
				return BASIC_SUGGESTIONS;
			}
			const missingSuggestions = addTypeSuggestions(this.type, this.values);
			const suggestions = [
				...this.sggestedNonSchemaTypes.map(suggested => suggested.type),
				...missingSuggestions
			];
			return _.uniq(suggestions);
		},
		getSuggestedList() {
			return this.addMissingSuggestions().map(type => {
				return {
					type,
					label: getLabelFromType(type),
					isRecommended: this.topNonSchemaType.type === type,
					isSelected: this.type === type,
				};
			});
		},
		onTypeChange(suggestedType) {
			const type = suggestedType;
			const dataset = this.dataset;
			const field = this.field;
			datasetActions.setVariableType(this.$store, {
				dataset: dataset,
				field: field,
				type: type,
				isTypeChanged: true,
			}).then(() => {
				if (this.target) {
					datasetActions.fetchVariableRankings(this.$store, {
						dataset: dataset,
						target: this.target
					});
				}
				if (isLocationType(type)) {
					datasetActions.geocodeVariable(this.$store, {
						dataset: dataset,
						field: field
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
.type-change-menu .dropdown-item {
	font-size: 0.867rem;
	text-transform: none;
	position: relative;
}
.type-change-menu .dropdown-item.selected {
	font-size: 0.867rem;
	text-transform: none;
	padding-left: 0;
}
.recommended-icon {
	position: absolute;
    right: 10px;
    bottom: 7px;
}
.unsure-type-icon {
	position: absolute;
    color: #dc3545;
    top: -5px;
    right: -5px;
    z-index: 2;

}
</style>
