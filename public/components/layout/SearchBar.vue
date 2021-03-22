<!--

    Copyright © 2021 Uncharted Software Inc.

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
-->

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
import { Lex } from "@uncharted.software/lex";
import { Variable } from "../../store/dataset/index";
import { variablesToLexLanguage, filterParamsToLexQuery } from "../../util/lex";
import "../../../node_modules/@uncharted.software/lex/dist/lex.css";
import "../../../node_modules/flatpickr/dist/flatpickr.min.css";

/** SearchBar component to display LexBar utility
 *
 * @param {string} [filters] - Accept filter from queryString to fill the LexBar with a query.
 * @param {string} [highlight] - Accept highlight from queryString to fill the LexBar with a query.
 * @param {Variable[]} [variables] - list of Variable used to fill the LexBar suggestions.
 */
export default Vue.extend({
  name: "SearchBar",

  props: {
    highlights: { type: String, default: "" },
    filters: { type: String, default: "" },
    variables: { type: Array as () => Variable[], default: [] },
  },

  data: () => ({
    lex: null,
  }),

  computed: {
    language(): Lex {
      return variablesToLexLanguage(this.variables);
    },
  },

  watch: {
    filters(n, o) {
      if (n !== o) {
        this.setQuery();
      }
    },

    highlights(n, o) {
      if (n !== o) {
        this.setQuery();
      }
    },

    language(n, o) {
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
        this.$emit("lex-query", args);
      });

      // Render our search bar into our desired element
      this.lex.render(this.$refs.lexcontainer);
      this.setQuery();
    },

    setQuery(): void {
      if (!this.lex || !(this.filters || this.highlights)) return;
      const lexQuery = filterParamsToLexQuery(
        this.filters,
        this.highlights,
        this.variables
      );
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
