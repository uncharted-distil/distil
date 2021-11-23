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
  <div class="search-bar-container d-flex">
    <component :is="styleSheet" v-html="cssStyle" />
    <header v-if="searchTitle.length">{{ searchTitle }}</header>
    <main ref="lexcontainer" class="lex-container" />
    <b-button
      v-if="hasHighlightsOrFilters"
      class="exit-button d-flex justify-content-center align-items-center m-auto"
      variant="outline-dark"
      @click="removeAllHighlightsAndFilters"
    >
      <i class="fas fa-times" />
    </b-button>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import _ from "lodash";
import { Lex } from "@uncharted.software/lex";
import { Variable } from "../../store/dataset/index";
import { getters as routeGetters } from "../../store/route/module";
import {
  variablesToLexLanguage,
  filterParamsToLexQuery,
  variableAggregation,
  TemplateInfo,
  lexQueryToFiltersAndHighlight,
} from "../../util/lex";
import { deepUpdateFiltersInRoute } from "../../util/filters";
import { updateHighlight, UPDATE_ALL } from "../../util/highlights";
import "../../../node_modules/@uncharted.software/lex/dist/lex.css";
import "../../../node_modules/flatpickr/dist/flatpickr.min.css";
import { EventList } from "../../util/events";
import { BIconX } from "bootstrap-vue";
import { overlayRouteEntry } from "../../util/routes";
/** SearchBar component to display LexBar utility
 *
 * @param {string} [filters] - Accept filter from queryString to fill the LexBar with a query.
 * @param {string} [highlight] - Accept highlight from queryString to fill the LexBar with a query.
 * @param {Variable[]} [variables] - list of Variable used to fill the LexBar suggestions.
 */
export default Vue.extend({
  name: "SearchBar",
  components: {
    BIconX,
  },
  props: {
    highlights: { type: String, default: null },
    filters: { type: String, default: null },
    variables: { type: Array as () => Variable[], default: [] },
    isSelectView: { type: Boolean as () => boolean, default: false },
    handleUpdates: { type: Boolean as () => boolean, default: false },
    searchTitle: { type: String as () => string, default: "" },
  },

  data: () => ({
    lex: null,
    isHighlightActive: true,
  }),

  computed: {
    hasHighlightsOrFilters(): boolean {
      return this.highlights !== null || this.filters !== null;
    },
    styleSheet(): string {
      return "style";
    },
    cssStyle(): string {
      let result = "";
      const end = this.templateInfo.activeVariables.length - 1;
      for (let i = 0; i < end; ++i) {
        result += `.lex-box > *:nth-child(${i + 1})::after {
          content: "${
            this.templateInfo.activeVariables[i].isEndOfSet ? "OR" : "&"
          }";
          font-weight: bolder;
          font-size: 0.938rem;
          margin-right: 9px;
          }`;
      }
      return result;
    },
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    language(): Lex {
      return variablesToLexLanguage(
        this.templateInfo,
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
      return _.cloneDeep(
        !this.isSelectView
          ? [this.filters, this.highlights, this.variables]
          : this.isHighlightActive
          ? [null, this.highlights, this.variables]
          : [this.filters, null, this.variables]
      );
    },
  },

  watch: {
    isHighlightActive() {
      this.renderLex();
    },

    variableInfo(n, o) {
      if (!_.isEqual(n, o)) {
        this.renderLex();
      }
    },
  },
  mounted() {
    this.renderLex();
  },
  methods: {
    removeAllHighlightsAndFilters() {
      const entry = overlayRouteEntry(this.$route, {
        highlights: "",
        filters: "",
      });
      this.$router.push(entry).catch((err) => console.warn(err));
    },
    async renderLex(): Promise<void> {
      // Initialize lex instance
      this.lex = new Lex({
        language: this.language,
        tokenXIcon: '<i class="fa fa-times" />',
      });

      this.lex.on("query changed", (
        ...args /* [newModel, oldModel, newUnboxedModel, oldUnboxedModel, nextTokenStarted] */
      ) => {
        if (!this.handleUpdates) {
          this.$emit(EventList.LEXBAR.QUERY_CHANGE_EVENT, args);
          return;
        }
        const lqfh = lexQueryToFiltersAndHighlight(
          args,
          this.dataset,
          this.variables
        );
        if (
          !this.isSelectView ||
          (this.isSelectView && !this.isHighlightActive)
        ) {
          deepUpdateFiltersInRoute(this.$router, lqfh.filters);
        }
        if (
          !this.isSelectView ||
          (this.isSelectView && this.isHighlightActive)
        ) {
          updateHighlight(this.$router, lqfh.highlights, UPDATE_ALL);
        }
      });

      // Render our search bar into our desired element
      this.lex.render(this.lexContainerRef);
      await this.setQuery();
    },

    async setQuery(): Promise<void> {
      if (!this.lex || !(this.filters || this.highlights)) return;
      const lexQuery = filterParamsToLexQuery(
        this.templateInfo,
        this.variableMap
      );
      await this.lex.setQuery(lexQuery, false);
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
div.lex-assistant-box ul li.selectable.active,
div.lex-assistant-box ul li.selectable.hoverable:hover {
  background-color: #255dcc;
}
div.lex-assistant-box {
  z-index: 999;
}
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
.lex-container {
  width: 95vw;
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
div.lex-box.focused {
  border-color: #80bdff;
}

.lex-container div.lex-box {
  min-height: 53px;
  border: none;
}
div.lex-box.active,
div.lex-box.focused {
  -webkit-box-shadow: 0 0 0 0.2rem rgba(0, 123, 255, 0.25);
  box-shadow: 0 0 0 0.2rem rgba(0, 123, 255, 0.25);
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
.include-filter {
  color: var(--white) !important;
  background-color: #255dcc !important;
}
.exclude-filter {
  color: var(--white) !important;
  background-color: #333333 !important;
}
.search-bar-container {
  border: 1px solid var(--border-color);
}
.exit-button {
  width: 36px;
  height: 36px;
}
</style>
