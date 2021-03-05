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
  <facet-bars :data.prop="facetData">
    <div slot="footer" class="facet-footer" />
  </facet-bars>
</template>

<script lang="ts">
import Vue from "vue";

import "@uncharted.software/facets-core";
import { FacetBarsData } from "@uncharted.software/facets-core/dist/types/facet-bars/FacetBars";

import { VariableSummary } from "../../store/dataset";

export default Vue.extend({
  name: "FacetLoading",

  props: {
    summary: Object as () => VariableSummary,
  },

  computed: {
    facetData(): FacetBarsData | null {
      const summary = this.summary;
      return {
        label: summary.label.toUpperCase(),
        values: [
          null,
          null,
          null,
          null,
          null,
          null,
          null,
          null,
          null,
          null,
          null,
          null,
          null,
          null,
          null,
        ],
      };
    },
  },
});
</script>

<style scoped>
.facet-footer {
  height: 25px;
}
</style>
