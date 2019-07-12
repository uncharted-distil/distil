<template>
	<div class="type-change-menu">
		<div class="type-change-dropdown-wrapper">
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
					<icon-base v-if="suggested.isRecommended" icon-name="bookmark" class="recommended-icon"><icon-bookmark /></icon-base>
				</b-dropdown-item>

				<template v-if="showGroupingOptions">
					<b-dropdown-divider></b-dropdown-divider>
					<b-dropdown-item
						v-for="grouping in groupingOptions()"
						@click.stop="onGroupingSelect(grouping.type)"
						:key="grouping.type">
						{{grouping.label}}
					</b-dropdown-item>
				</template>


			</b-dropdown>
			<i v-if="isUnsure" class="unsure-type-icon fa fa-circle"></i>
		</div>
		<b-tooltip :delay="delay" :disabled="!isDisabled" target="type-change-dropdown">
			Cannot change type when actively filtering
		</b-tooltip>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import IconBase from './icons/IconBase';
import IconBookmark from './icons/IconBookmark';
import { SuggestedType, Variable, Highlight } from '../store/dataset/index';
import { actions as datasetActions, getters as datasetGetters } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';
import { addTypeSuggestions, getLabelFromType, TIMESERIES_TYPE, getTypeFromLabel, isEquivalentType, isLocationType, normalizedEquivalentType, BASIC_SUGGESTIONS } from '../util/types';
import { hasFilterInRoute } from '../util/filters';
import { createRouteEntry } from '../util/routes';
import { GROUPING_ROUTE } from '../store/route';

const PROBABILITY_THRESHOLD = 0.8;

export default Vue.extend({
	name: 'type-change-menu',

	components: {
		IconBase,
		IconBookmark,
	},
	props: {
		dataset: String as () => string,
		field: String as () => string
	},

	computed: {
		variable(): Variable {
			const vars = datasetGetters.getVariables(this.$store);
			if (!vars) {
				return null;
			}
			return vars.find(v => {
				return v.colName.toLowerCase() === this.field.toLowerCase() &&
					v.datasetName === this.dataset;
			});
		},
		type(): string {
			return this.variable ? this.variable.colType : '';
		},
		isColTypeReviewed(): boolean {
			return this.variable ? this.variable.isColTypeReviewed : false;
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
		suggestedNonSchemaTypes(): SuggestedType[] {
			const nonSchemaTypes = _.filter(this.suggestedTypes, t => {
				return t.provenance !== 'schema';
			});
			return nonSchemaTypes;
		},
		topNonSchemaType(): SuggestedType {
			return this.suggestedNonSchemaTypes.length > 0 ? this.suggestedNonSchemaTypes[0] : undefined;
		},
		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},
		highlight(): Highlight {
			return routeGetters.getDecodedHighlight(this.$store);
		},
		isDisabled(): boolean {
			return hasFilterInRoute(this.field) || (this.highlight && this.highlight.key === this.field);
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
				!this.isColTypeReviewed && // check if user ever reviewed the col type (client)
				this.hasSchemaType && this.hasNonSchemaTypes &&
				this.topNonSchemaType.probability >= PROBABILITY_THRESHOLD && // it has both schema and ML types
				!isEquivalentType(this.schemaType.type, this.topNonSchemaType.type)); // they don't agree
		},
		delay(): any {
			return {
				show: 10,
				hide: 10
			};
		},
		showGroupingOptions(): boolean {
			return true;
		}
	},

	methods: {

		groupingOptions() {
			return [
				{
					type: TIMESERIES_TYPE,
					label: 'Timeseries'
				}
			];
		},

		onGroupingSelect() {
			const entry = createRouteEntry(GROUPING_ROUTE, {
				dataset: routeGetters.getRouteDataset(this.$store)
			});
			this.$router.push(entry);
		},

		addMissingSuggestions() {
			const flatSuggestedTypes = this.suggestedTypes.map(st => st.type);
			const missingSuggestions = addTypeSuggestions(flatSuggestedTypes);
			const nonSchemaSuggestions = this.suggestedNonSchemaTypes.map(suggested => normalizedEquivalentType(suggested.type));
			const menuSuggestions = _.uniq([
				...nonSchemaSuggestions,
				...missingSuggestions
			]);
			return menuSuggestions;
		},
		getSuggestedList() {
			const currentNormalizedType = normalizedEquivalentType(this.type);
			const combinedSuggestions = this.addMissingSuggestions().map(type => {
				const normalizedType = normalizedEquivalentType(type);
				return {
					type: normalizedType,
					label: getLabelFromType(normalizedType),
					isRecommended: this.topNonSchemaType && this.topNonSchemaType.type.toLowerCase() === type.toLowerCase(),
					isSelected: currentNormalizedType === normalizedType,
				};
			});
			return combinedSuggestions;
		},
		onTypeChange(suggestedType) {
			const type = suggestedType;
			const dataset = this.dataset;
			const field = this.field;
			datasetActions.setVariableType(this.$store, {
				dataset: dataset,
				field: field,
				type: type
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
	},

	mounted() {
		this.$root.$on('bv::dropdown::show', () => {
			const dataset = this.dataset;
			const field = this.field;
			datasetActions.reviewVariableType(this.$store, {
				dataset: dataset,
				field: field,
				isColTypeReviewed: true,
			});
		});
	},
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
    bottom: 5px;
}
.unsure-type-icon {
	position: absolute;
    color: #dc3545;
    top: -5px;
    right: -5px;
    z-index: 2;
}
.type-change-dropdown-wrapper {
	position: relative;
}
</style>
