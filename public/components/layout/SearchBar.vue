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
import {
  variablesToLexLanguage,
  filterParamsToLexQuery,
  variableAggregation,
  TemplateInfo,
} from "../../util/lex";
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
      return variablesToLexLanguage(
        this.templateInfo.activeVariables,
        this.variables,
        this.variableMap
      );
    },
    templateInfo(): TemplateInfo {
      return variableAggregation(this.filters, this.highlights, this.variables);
    },
    variableMap(): Map<string, Variable> {
      return new Map(
        this.variables.map((v) => {
          return [v.key, v];
        })
      );
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
        this.templateInfo,
        this.variableMap
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
  right: 0px;
  color: var(--white) !important;
}
.lex-container div.lex-box button.btn:hover {
  color: #999999 !important;
}
.lex-assistant-box {
  z-index: var(--z-index-lexbar-assistant);
}
.token-vkey-operator {
  font-weight: bold !important;
}
.token {
  white-space: normal !important;
}
.lex-box > *:not(:nth-last-child(-n + 2))::after {
  content: "&";
  font-weight: bolder;
  font-size: 0.938rem;
  margin-right: 9px;
}
.include-filter {
  color: var(--white) !important;
  background-color: #255dcc !important;
}
.exclude-filter {
  color: var(--white) !important;
  background-color: #333333 !important;
}
</style>
