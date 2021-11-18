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
          <div>Dataset Name</div>
        </th>
        <th
          role="columnheader"
          scope="col"
          aria-colindex="3"
          class="table-b-table-default border-top-0"
        >
          <div>Features</div>
        </th>
        <th
          role="columnheader"
          scope="col"
          aria-colindex="4"
          class="table-b-table-default border-top-0"
        >
          <div>Rows</div>
        </th>
        <th
          role="columnheader"
          scope="col"
          aria-colindex="5"
          class="table-b-table-default border-top-0"
        >
          <div>Size</div>
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
              <span v-else @click="setActiveDataset(i)">{{ ele }}</span>
            </td>
          </tr>
          <tr v-if="item.open" :key="`expanded-${i}`">
            <td role="cell" class="border-top-0" />
            <td role="cell" class="border-top-0 text-decoration-none">
              <div class="row m-0 p-0"><b>Top Features:</b></div>
              <div class="row m-0 p-0">
                <b>May relate to topics such as:</b>
              </div>
              <div class="row m-0 p-0"><b>Summary:</b></div>
              <div v-if="expandedItems[i].moreDetails" class="row m-0 p-0">
                <b>Full Description</b>
              </div>
              <div class="row ml-0 mr-0 mb-3 mt-3 p-0">
                <b-button
                  v-if="expandedItems[i].fullDescription"
                  size="sm"
                  @click="onMoreDetails(i)"
                >
                  <span v-if="!expandedItems[i].moreDetails">More Details</span>
                  <span v-else>Less Details</span>
                </b-button>
              </div>
            </td>
            <td
              role="cell"
              colspan="3"
              class="border-top-0 text-decoration-none"
            >
              <div class="row m-0 p-0">{{ expandedItems[i].topVariables }}</div>
              <div class="row m-0 p-0">{{ expandedItems[i].summaryML }}</div>
              <div class="row m-0 p-0">{{ expandedItems[i].summary }}</div>
              <div
                v-if="expandedItems[i].moreDetails"
                class="row m-0 p-0"
                v-html="expandedItems[i].fullDescription"
              />
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
import { appActions } from "../../store";
import { Dataset } from "../../store/dataset";
import { DATA_EXPLORER_ROUTE } from "../../store/route";
import { formatBytes } from "../../util/bytes";
import { getters as routeGetters } from "../../store/route/module";
import {
  addRecentDataset,
  filterVariablesByFeature,
  sortVariablesByPCARanking,
} from "../../util/data";
import { EventList } from "../../util/events";
import { createRouteEntry } from "../../util/routes";
import { Activity, Feature, SubActivity } from "../../util/userEvents";
import _ from "lodash";

const NUM_TOP_FEATURES = 5;
interface Item {
  open: boolean;
  DatasetName: string;
  Features: number;
  Rows: number;
  Size: string;
}
interface ExpandedItem {
  topVariables: string;
  summaryML: string;
  summary: string;
  fullDescription: string;
  id: string;
  moreDetails: boolean;
}
export default Vue.extend({
  name: "DatasetPreviewTable",

  props: {
    datasets: {
      type: Array as () => Dataset[],
      default: () => [] as Dataset[],
    },
  },
  data() {
    return {
      items: [] as Item[],
      expandedItems: [] as ExpandedItem[],
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
      const moreDetailsMap = new Map(
        this.expandedItems.map((item) => {
          return [item.id, item.moreDetails];
        })
      );
      this.items = this.formatItems(openMap);
      this.expandedItems = this.formatExpandedItems(moreDetailsMap);
    },
  },
  beforeMount() {
    this.items = this.formatItems(new Map());
    this.expandedItems = this.formatExpandedItems(new Map());
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
    onMoreDetails(index: number) {
      this.expandedItems[index].moreDetails = !this.expandedItems[index]
        .moreDetails;
    },
    onExpand(index: number) {
      this.items[index].open = !this.items[index].open;
    },
    onDeleteClicked(index: number) {
      this.$emit(EventList.DATASETS.DELETE_EVENT, this.datasets[index]);
    },
    formatItems(openMap: Map<string, boolean>): Item[] {
      return this.datasets.map((d) => {
        return {
          open: openMap.get(d.name),
          DatasetName: d.name,
          Features: filterVariablesByFeature(d.variables).length,
          Rows: d.numRows,
          Size: formatBytes(d.numBytes),
        };
      });
    },
    formatExpandedItems(openMap: Map<string, boolean>): ExpandedItem[] {
      return this.datasets.map((d) => {
        return {
          topVariables: sortVariablesByPCARanking(
            filterVariablesByFeature(d.variables).slice(0)
          )
            .slice(0, NUM_TOP_FEATURES)
            .map((d) => d.colDisplayName)
            .join(", "),
          summaryML: d.summaryML || "n/a",
          summary: d.summary || "n/a",
          fullDescription: this.highlightedDescription(d.description),
          id: d.id,
          moreDetails: openMap.get(d.id),
        };
      });
    },
    setActiveDataset(index: number) {
      const dataset = this.expandedItems[index].id;
      const entry = createRouteEntry(DATA_EXPLORER_ROUTE, {
        dataset,
      });
      this.$router.push(entry).catch((err) => console.warn(err));
      addRecentDataset(dataset);
      appActions.logUserEvent(this.$store, {
        feature: Feature.SELECT_DATASET,
        activity: Activity.DATA_PREPARATION,
        subActivity: SubActivity.DATA_OPEN,
        details: { dataset },
      });
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
