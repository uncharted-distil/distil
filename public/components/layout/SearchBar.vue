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
    <b-tabs
      v-if="isSelectView"
      content-class="pr-0 pl-0 ml-15px"
      nav-wrapper-class="pl-0"
      pills
      vertical
      end
    >
      <b-tab
        title="☀"
        :active="isHighlightActive"
        aria-h="Highlight"
        @click="isHighlightActive = !isHighlightActive"
        title-link-class="searchbar-nav btn-outline-secondary btn border-bottom-right-radius-0"
      >
        <main ref="lexcontainerHighlight" class="lex-container select" />
      </b-tab>
      <b-tab
        title="≠"
        :active="!isHighlightActive"
        aria-label="Exclude"
        title-link-class="searchbar-nav btn-outline-secondary btn border-top-right-radius-0"
        @click="isHighlightActive = !isHighlightActive"
      >
        <main ref="lexcontainerExclude" class="lex-container select" />
      </b-tab>
    </b-tabs>
    <main v-else ref="lexcontainer" class="lex-container" />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import _ from "lodash";
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
    isSelectView: { type: Boolean as () => boolean, default: false },
  },

  data: () => ({
    lex: null,
    isHighlightActive: true,
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
      return variableAggregation(...this.variableInfo);
    },
    variableMap(): Map<string, Variable> {
      return new Map(
        this.variables.map((v) => {
          return [v.key, v];
        })
      );
    },
    lexContainerRef(): Vue | Element | Vue[] | Element[] {
      return !this.isSelectView
        ? this.$refs.lexcontainer
        : this.isHighlightActive
        ? this.$refs.lexcontainerHighlight
        : this.$refs.lexcontainerExclude;
    },
    variableInfo(): [string, string, Variable[]] {
      return !this.isSelectView
        ? [this.filters, this.highlights, this.variables]
        : this.isHighlightActive
        ? [null, this.highlights, this.variables]
        : [this.filters, null, this.variables];
    },
  },

  watch: {
    isHighlightActive() {
      this.renderLex();
    },
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
  mounted() {
    this.renderLex();
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
      this.lex.render(this.lexContainerRef);
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
.ml-15px {
  margin-left: 15px;
}
.border-top-right-radius-0 {
  border-top-right-radius: 0px !important;
}
.border-bottom-right-radius-0 {
  border-bottom-right-radius: 0px !important;
}
.searchbar-nav {
  height: 40px;
  border-top-left-radius: 0% !important;
  border-bottom-left-radius: 0% !important;
  box-shadow: none !important;
}
.lex-container.select div.lex-box {
  min-height: 80px;
  overflow-y: hidden;
  height: 80px;
  background-color: var(--gray-300);
}
.lex-container.select div.lex-box:hover {
  z-index: 1000;
  top: 0%;
  position: absolute !important;
  display: block;
  height: auto !important;
  max-height: 40vh;
}
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
