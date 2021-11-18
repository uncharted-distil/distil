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
  <div class="b-table-sticky-header" style="max-height: 95%">
    <table
      role="table"
      aria-busy="false"
      aria-colcount="5"
      class="table b-table tableFixHead table-hover"
    >
      <thead role="rowgroup">
        <th
          role="columnheader"
          scope="col"
          aria-colindex="1"
          class="table-b-table-default border-top-0"
        />
        <th
          role="columnheader"
          scope="col"
          aria-colindex="2"
          class="table-b-table-default border-top-0"
        >
          <div>Model Name</div>
        </th>
        <th
          role="columnheader"
          scope="col"
          aria-colindex="3"
          class="table-b-table-default border-top-0"
        >
          <div>Dataset Name</div>
        </th>
        <th
          role="columnheader"
          scope="col"
          aria-colindex="4"
          class="table-b-table-default border-top-0"
        >
          <div>Features</div>
        </th>
        <th
          role="columnheader"
          scope="col"
          aria-colindex="5"
          class="table-b-table-default border-top-0"
        >
          <div>Target</div>
        </th>
      </thead>
      <tbody role="rowgroup">
        <template v-for="(item, i) in items">
          <tr :key="i" role="row" class="mt-3">
            <td
              v-for="(ele, _, ii) in item"
              :key="ii"
              :aria-colindex="ii + 1"
              role="cell"
              class="p-3"
              :class="item.open ? 'border-bottom-0' : ''"
            >
              <i
                v-if="ii == 0"
                :class="ele ? 'fas fa-chevron-down' : 'fas fa-chevron-right'"
                @click="onExpand(i)"
              />
              <span v-else @click="setActiveModel(i)">{{ ele }}</span>
            </td>
          </tr>
          <tr v-if="item.open" :key="`expanded-${i}`">
            <td role="cell" class="border-top-0" />
            <td role="cell" class="border-top-0 text-decoration-none">
              <div class="row m-0 p-0"><b>Top Features:</b></div>
              <div class="row m-0 p-0"><b>Description:</b></div>
              <div class="row m-0 p-0"><b>All Variables:</b></div>
            </td>
            <td
              role="cell"
              colspan="3"
              class="border-top-0 text-decoration-none"
            >
              <div class="row m-0 p-0">{{ expandedItems[i].topVariables }}</div>
              <div class="row m-0 p-0">{{ expandedItems[i].description }}</div>
              <div class="row m-0 p-0">{{ expandedItems[i].allVariables }}</div>
              <div class="row ml-0 mr-0 mb-3 mt-3 p-0 flex-row-reverse">
                <b-button
                  variant="danger"
                  size="sm"
                  class="mr-3"
                  @click="onDeleteClicked(i)"
                >
                  <i class="fa fa-trash" aria-hidden="true" />
                </b-button>
              </div>
            </td>
          </tr>
        </template>
      </tbody>
    </table>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { DATA_EXPLORER_ROUTE } from "../../store/route";
import { getters as routeGetters } from "../../store/route/module";
import { EventList } from "../../util/events";
import { createRouteEntry } from "../../util/routes";
import _ from "lodash";
import { Model } from "../../store/model";
import { openModelSolution } from "../../util/solutions";
import { ExplorerStateNames } from "../../util/explorer";

const NUM_TOP_FEATURES = 5;

export default Vue.extend({
  name: "ModelPreviewTable",

  props: {
    models: {
      type: Array as () => Model[],
      default: () => [] as Model[],
    },
  },
  data() {
    return {
      items: [],
      expandedItems: [],
    };
  },
  computed: {
    terms(): string {
      return routeGetters.getRouteTerms(this.$store);
    },
  },
  watch: {
    datasets() {
      const openMap = new Map(
        this.items.map((item) => {
          return [item.DatasetName, item.open];
        })
      );
      this.items = this.formatItems(openMap);
      this.expandedItems = this.formatExpandedItems();
    },
  },
  beforeMount() {
    this.items = this.formatItems(new Map());
    this.expandedItems = this.formatExpandedItems();
  },
  methods: {
    highlightedDescription(datasetDescription: string): string {
      const terms = this.terms;
      if (_.isEmpty(terms)) {
        return datasetDescription;
      }
      const split = terms.split(/[ ,]+/); // split on whitespace
      const joined = split.join("|"); // join
      const regex = new RegExp(`(${joined})(?![^<]*>)`, "gm");
      return datasetDescription.replace(
        regex,
        '<span class="highlight">$1</span>'
      );
    },
    onExpand(index: number) {
      this.items[index].open = !this.items[index].open;
    },
    onDeleteClicked(index: number) {
      this.$emit(EventList.MODEL.DELETE_EVENT, this.models[index]);
    },
    formatItems(openMap: Map<string, boolean>) {
      return this.models.map((m) => {
        return {
          open: openMap.get(m.modelName),
          modelName: m.modelName,
          DatasetName: m.datasetName,
          Features: m.variables.length,
          Target: m.target.displayName,
        };
      });
    },
    formatExpandedItems() {
      return this.models.map((m) => {
        const sortedVars = m.variableDetails
          .slice()
          .sort((a, b) => b.rank - a.rank);
        return {
          topVariables: sortedVars
            .slice(0, NUM_TOP_FEATURES)
            .map((a) => a.displayName)
            .join(", "),
          description: m.modelDescription || "n/a",
          allVariables: sortedVars.map((v) => v.displayName).join(", "),
        };
      });
    },
    async setActiveModel(index: number) {
      const model = this.models[index];
      const route = DATA_EXPLORER_ROUTE;
      const args = await openModelSolution(this.$router, {
        datasetId: model.datasetId,
        targetFeature: model.target.key,
        fittedSolutionId: model.fittedSolutionId,
        variableFeatures: model.variables,
      });
      args.dataExplorerState = ExplorerStateNames.RESULT_VIEW;
      const entry = createRouteEntry(route, args);
      this.$router.push(entry).catch((err) => console.debug(err));
    },
  },
});
</script>
<style scoped>
.tableFixHead thead th {
  position: sticky;
  top: 0;
  z-index: 1;
  background-color: white;
}
.table-hover tr td:hover {
  text-decoration: underline;
}
.table-hover tr td {
  text-decoration: none;
}
</style>
