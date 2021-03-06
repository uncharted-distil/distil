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
  <b-modal :id="modalId" size="huge" :title="title">
    <label-header-buttons @button-event="passThrough" />
    <div class="row flex-1 pb-3 h-100">
      <div class="col-12 col-md-6 d-flex flex-column h-100">
        <h5 class="header-title">Most Similar</h5>
        <component
          :is="viewComponent"
          :instanceName="instanceName"
          :data-items="items"
          :data-fields="dataFields"
          :summaries="summaries"
          includedActive
        />
      </div>
      <div class="col-12 col-md-6 d-flex flex-column h-100">
        <h5 class="header-title">Randomly Chosen</h5>
        <component
          :is="viewComponent"
          :instanceName="instanceName"
          :data-items="randomItems"
          :data-fields="dataFields"
          :summaries="summaries"
          includedActive
        />
      </div>
    </div>
    <template #modal-footer="{}">
      <!-- Emulate built in modal footer ok and cancel button actions -->
      <b-button size="lg" variant="secondary" @click="onApply">
        <template v-if="isLoading">
          <div v-html="spinnerHTML"></div>
        </template>
        <template v-else> Apply </template>
      </b-button>
    </template>
  </b-modal>
</template>

<script lang="ts">
import Vue from "vue";
import { Dictionary } from "../../util/dict";
import {
  TableRow,
  VariableSummary,
  TableColumn,
} from "../../store/dataset/index";
import {
  LOW_SHOT_LABEL_COLUMN_NAME,
  LowShotLabels,
  getRandomInt,
} from "../../util/data";
import SelectDataTable from "../SelectDataTable.vue";
import ImageMosaic from "../ImageMosaic.vue";
import LabelHeaderButtons from "./LabelHeaderButtons.vue";
import { circleSpinnerHTML } from "../../util/spinner";

export default Vue.extend({
  name: "label-score-pop-up",
  components: {
    SelectDataTable,
    ImageMosaic,
    LabelHeaderButtons,
  },
  props: {
    data: { type: Array as () => TableRow[], default: () => [] as TableRow[] },
    instanceName: {
      type: String as () => string,
      default: "label-score-pop-up",
    },
    summaries: {
      type: Array as () => VariableSummary[],
      default: [],
    },
    dataFields: {
      type: Object as () => Dictionary<TableColumn>,
      default: {},
    },
    title: { type: String as () => string, default: "Scores" },
    isLoading: { type: Boolean as () => boolean, default: false },
    modalId: { type: String as () => string, default: "score-popup" },
    isRemoteSensing: { type: Boolean as () => boolean, default: false }, // default to false to support every dataset
  },
  data() {
    return { sampleSize: 10 };
  },
  computed: {
    itemMap(): Map<number, TableRow> {
      return new Map(
        this.data?.map((d) => {
          return [d.d3mIndex, d];
        })
      );
    },
    items(): TableRow[] {
      if (!this.data) {
        return [];
      }
      return this.data.slice(0, this.sampleSize);
    },
    randomItems(): TableRow[] {
      const random = this.data?.filter((d) => {
        return d[LOW_SHOT_LABEL_COLUMN_NAME]?.value === LowShotLabels.unlabeled;
      });
      if (!random) {
        return null;
      }
      const randomIdx = getRandomInt(random.length - this.sampleSize);
      return random.slice(randomIdx, randomIdx + this.sampleSize);
    },
    spinnerHTML(): string {
      return circleSpinnerHTML();
    },
    viewComponent(): string {
      if (!this.isRemoteSensing) {
        return "SelectDataTable";
      }
      return "ImageMosaic";
    },
  },
  methods: {
    passThrough(event: string) {
      this.$emit("button-event", event);
    },
    onApply() {
      this.$emit("apply");
    },
  },
});
</script>

<style>
@media (min-width: 992px) {
  .modal .modal-huge {
    max-width: 90% !important;
    width: 90% !important;
  }
}
</style>
