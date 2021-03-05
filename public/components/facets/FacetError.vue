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
  <!-- facet-container is undocumented! shhhhh.... -->
  <facet-container>
    <div slot="header" class="facet-header">
      <i class="fa fa-exclamation-circle" />
      <span>{{ summary.label.toUpperCase() }}</span>
      <type-change-menu
        v-if="facetEnableTypeChanges"
        class="facet-header-dropdown"
        :dataset="summary.dataset"
        :field="summary.key"
      />
    </div>
    <div slot="content" class="facet-content">
      <span>{{ summary.err }}</span>
    </div>
  </facet-container>
</template>

<script lang="ts">
import Vue from "vue";

import "@uncharted.software/facets-core";
import { VariableSummary } from "../../store/dataset";
import { facetTypeChangeState } from "../../util/facets";
import TypeChangeMenu from "../TypeChangeMenu";

export default Vue.extend({
  name: "FacetError",

  components: {
    TypeChangeMenu,
  },

  props: {
    summary: Object as () => VariableSummary,
    enabledTypeChanges: Array as () => string[],
  },

  computed: {
    facetEnableTypeChanges(): boolean {
      return facetTypeChangeState(
        this.summary.dataset,
        this.summary.key,
        this.enabledTypeChanges
      );
    },
  },
});
</script>

<style scoped>
.facet-header {
  height: 20px;
  color: #1a1b1c;
  font-family: "IBM Plex Sans", sans-serif;
  font-size: 14px;
  font-style: normal;
  font-weight: 600;
  line-height: 20px;
  padding: 6px 12px 5px;
  box-sizing: content-box;

  color: rgba(0, 0, 0, 0.54);
  display: flex;
  align-items: center;
  overflow-y: scroll !important;
}

.facet-header > i {
  color: red;
  padding-right: 6px;
}

.facet-header .dropdown-menu {
  max-height: 200px;
  overflow-y: auto;
}

.facet-header-dropdown {
  position: absolute;
  right: 12px;
  top: 5px;
}

.facet-content {
  padding: 4px 12px 25px;
}
</style>
