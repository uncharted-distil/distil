<template>
  <div class="search-bar-container">
    <header>Search</header>
    <main ref="lexContainer" />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import _ from "lodash";
import { h } from "preact";
import { Lex, ValueState } from "@uncharted.software/lex";
import "../../styles/lex.css";

export default Vue.extend({
  name: "SearchBar",

  data: () => ({
    lex: null,
    suggestions: [],
  }),

  mounted() {
    // Defines a list of searchable fields for LEX
    this.suggestions = [];

    const language = Lex.from("field", ValueState, {
      name: "Choose a variable to filter",
      suggestions: this.suggestions,
    });

    // Initialize lex instance
    this.lex = new Lex({
      language: language,
      tokenXIcon: '<i class="fa fa-times" />',
    });

    this.lex.on("query changed", (
      ...args /* [newModel, oldModel, newUnboxedModel, oldUnboxedModel, nextTokenStarted] */
    ) => {
      console.debug("lex event `query changed`");
      // this.setFilters(newFilters);
    });

    // Render our search bar into our desired element
    this.lex.render(this.$refs.lexContainer);
    // this.setQuery();
  },

  methods: {
    // setQuery(): void {
    //   if (!this.lex) return;
    //   const lexQuery = [];
    //   // this.getFilters();
    //   this.lex.setQuery(lexQuery, false);
    // },
  },
});
</script>

<style scoped>
header {
  font-style: bold;
}
</style>
