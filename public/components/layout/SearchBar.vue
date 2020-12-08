<template>
  <div class="search-bar-container">
    <header>Search</header>
    <main ref="lexcontainer" class="lex-container" />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import _ from "lodash";
import { h } from "preact";
import {
  LabelState,
  Lex,
  NumericEntryState,
  TextEntryState,
  TransitionFactory,
  ValueState,
  ValueStateValue,
} from "@uncharted.software/lex";
import { Variable } from "../../store/dataset/index";
import "../../../node_modules/@uncharted.software/lex/dist/lex.css";

/** SearchBar component to display LexBar utility
 *
 * @param {string} [filters] - Accept filters to fill the LexBar with a query.
 * @param {Variable[]} [variables] - list of Variable used to fill the LexBar suggestions.
 * @event lexFilter - when the user interact with the search bar, this event fire with the new filters.
 */
export default Vue.extend({
  name: "SearchBar",

  data: () => ({
    lex: null,
  }),

  props: {
    filters: { type: String, default: "" }, // TODO - random type for now.
    variables: { type: Array as () => Variable[], default: [] },
  },

  computed: {
    language(): any {
      return Lex.from("field", ValueState, {
        name: "Choose a variable to filter",
        icon: '<i class="fa fa-filter" />',
        suggestions: this.suggestions,
      }).branch(
        Lex.from(LabelState, {
          label: "From",
          ...TransitionFactory.valueMetaCompare({ type: "numeric" }),
        })
          .to("lower bound", NumericEntryState, { name: "Enter lower bound" })
          .to(LabelState, { label: "To" })
          .to("upper bound", NumericEntryState, { name: "Enter upper bound" })
      );
    },

    suggestions(): ValueStateValue[] {
      if (_.isEmpty(this.variables)) return;
      return this.variables.map((variable) => {
        const name = _.capitalize(variable.colDisplayName);
        const options = { type: "numeric" }; // variable.colType
        return new ValueStateValue(name, options);
      });
    },
  },

  watch: {
    filters(n, o) {
      if (n !== o) {
        this.setQuery();
      }
    },

    variables(n, o) {
      if (n !== o) {
        this.renderLex();
      }
    },
  },

  methods: {
    renderLex(): void {
      // Initialize lex instance
      this.lex = new Lex({
        language: this.language,
        tokenXIcon: '<i class="fa fa-times" />',
      });

      this.lex.on("query changed", (
        ...args /* [newModel, oldModel, newUnboxedModel, oldUnboxedModel, nextTokenStarted] */
      ) => {
        const filters = []; // TODO - lexModelToFilters(args[0]);
        this.$emit("lex-filter", filters);
      });

      // Render our search bar into our desired element
      this.lex.render(this.$refs.lexcontainer);
      this.setQuery();
    },

    setQuery(): void {
      if (!this.lex) return;
      const lexQuery = []; // TODO - filtersToLexQuery(this.filters);
      this.lex.setQuery(lexQuery, false);
    },
  },
});
</script>

<style scoped>
header {
  font-style: bold;
}
</style>

<style>
.lex-container div.lex-box button.btn {
  line-height: 1em !important;
}

.lex-assistant-box {
  z-index: var(--z-index-lexbar-assistant);
}
</style>
