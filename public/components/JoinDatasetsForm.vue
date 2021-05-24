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
  <div class="create-solutions-form">
    <b-modal
      v-model="showJoinSuccess"
      size="xlg"
      modal-class="join-preview-modal"
      cancel-disabled
      hide-header
      hide-footer
    >
      <join-datasets-preview
        ref="datasetPreview"
        :preview-table-data="previewTableData"
        :dataset-a="datasetA"
        :dataset-b="datasetB"
        :joined-column="joinedColumn"
        :path="joinedPath"
        @success="onJoinCommitSuccess"
        @failure="onJoinCommitFailure"
        @close="showJoinSuccess = !showJoinSuccess"
      />
    </b-modal>
    <save-modal subject="Dataset" modalId="join-view-save" @save="onSave" />
    <error-modal
      :show="showJoinFailure"
      title="Join Failed"
      @close="showJoinFailure = !showJoinFailure"
    />

    <div
      v-if="columnTypesDoNotMatch"
      class="row justify-content-center mt-3 mb-3 warning-text"
    >
      <i class="fa fa-exclamation-triangle warning-icon mr-2" />
      <span v-html="joinWarning" />
    </div>
    <div class="d-flex justify-content-between bottom-margin form-height">
      <div class="d-flex">
        <b-button variant="primary" @click="swapDatasets" class="join-button">
          Swap Datasets
        </b-button>
      </div>
      <main class="d-flex w-50">
        <badge
          v-for="(pair, index) in joinPairs"
          class="d-flex justify-content-center align-items-center"
          :key="index"
          :content="`${pair.first}->${pair.second}`"
          :identifier="pair"
          @removed="badgeRemoved"
        />
      </main>
      <div class="d-flex">
        <b-button-group>
          <b-button
            variant="primary"
            class="h-100"
            :disabled="disableAdd"
            @click="addJoinRelation"
          >
            Add Join Relationship
          </b-button>
          <b-button
            variant="primary"
            class="h100"
            :disabled="disableJoin"
            v-b-modal.join-accuracy-modal
          >
            <i class="fa fa-cog" aria-hidden="true" />
          </b-button>
        </b-button-group>
        <b-button
          class="join-button"
          :disabled="disableJoin"
          :variant="joinVariant"
          @click="previewJoin"
        >
          <div class="d-flex justify-content-center align-items-center">
            <b v-if="!isPending">Join Datasets</b>
            <b-spinner v-if="isPending" small />
          </div>
        </b-button>
      </div>
    </div>
    <join-accuracy-modal />
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
// components
import vueSlider from "vue-slider-component";
import JoinDatasetsPreview from "../components/JoinDatasetsPreview.vue";
import ErrorModal from "../components/ErrorModal.vue";
import JoinAccuracyModal from "../components/JoinAccuracyModal.vue";
import Badge from "./Badge.vue";
import SaveModal from "./SaveModal.vue";
import { getters as routeGetters } from "../store/route/module";
import { Dataset, TableColumn, TableRow } from "../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../store/dataset/module";
import { Dictionary } from "../util/dict";
import { getTableDataItems, getTableDataFields, JoinPair } from "../util/data";
import { isJoinable } from "../util/types";
import { loadJoinedDataset } from "../util/join";
import { overlayRouteEntry, overlayRouteReplace } from "../util/routes";

export default Vue.extend({
  name: "JoinDatasetsForm",

  components: {
    JoinDatasetsPreview,
    ErrorModal,
    Badge,
    SaveModal,
    vueSlider,
    JoinAccuracyModal,
  },

  props: {
    datasetIdA: String as () => string,
    datasetIdB: String as () => string,
    datasetAColumn: Object as () => TableColumn,
    datasetBColumn: Object as () => TableColumn,
    joinAccuracy: Array as () => number[],
    joinAbsolute: Array as () => boolean[],
  },

  data() {
    return {
      pending: false,
      showJoin: false,
      showJoinSuccess: false,
      showJoinFailure: false,
      joinErrorMessage: null,
      previewTableData: null,
      joinedPath: "",
      datasetA: null,
      datasetB: null,
    };
  },
  computed: {
    datasets(): Dataset[] {
      return datasetGetters.getDatasets(this.$store);
    },
    target(): string {
      return routeGetters.getRoutePreviousTarget(this.$store);
    },
    returnPath(): string {
      return routeGetters.getPriorPath(this.$store);
    },
    columnsSelected(): boolean {
      return !!this.datasetAColumn && !!this.datasetBColumn;
    },
    columnTypesDoNotMatch(): boolean {
      return (
        this.datasetAColumn &&
        this.datasetBColumn &&
        !isJoinable(this.datasetAColumn.type, this.datasetBColumn.type)
      );
    },
    isPending(): boolean {
      return this.pending;
    },
    joinWarning(): string {
      if (this.columnTypesDoNotMatch) {
        return `Unable to join column <b>${this.datasetAColumn.key}</b> of type <b>${this.datasetAColumn.type}</b> with <b>${this.datasetBColumn.key}</b> of type <b>${this.datasetBColumn.type}</b>`;
      }
    },
    disableAdd(): boolean {
      return (
        this.isPending || !this.columnsSelected || this.columnTypesDoNotMatch
      );
    },
    joinVariant(): string {
      return !this.disableJoin ? "success" : "outline-secondary";
    },
    percentComplete(): number {
      return 100;
    },
    joinDataPreviewItems(): TableRow[] {
      return getTableDataItems(this.previewTableData);
    },
    joinDataPreviewFields(): Dictionary<TableColumn> {
      return getTableDataFields(this.previewTableData);
    },
    joinDataPreviewNumRows(): number {
      return this.previewTableData ? this.previewTableData.numRows : 0;
    },
    joinDataPreviewHasData(): boolean {
      return !!this.previewTableData;
    },
    disableJoin(): boolean {
      return this.joinPairs.length < 1;
    },
    joinedColumn(): string {
      const a = this.datasetAColumn ? this.datasetAColumn.key : "";
      const b = this.datasetBColumn ? this.datasetBColumn.key : "";
      // Note: It looks like joined column name is set to same as left column (a) name
      return a;
    },
    joinPairs(): JoinPair<string>[] {
      return routeGetters.getJoinPairs(this.$store);
    },
  },

  methods: {
    swapDatasets() {
      this.$emit("swap-datasets");
    },
    badgeRemoved(joinPair: JoinPair<string>) {
      const pairs = this.joinPairs.filter((jp) => {
        return jp.first !== joinPair.first || jp.second !== joinPair.second;
      });
      const strs = pairs.map((jp) => {
        return JSON.stringify(jp);
      });
      const entry = overlayRouteReplace(this.$route, {
        joinPairs: strs.length ? strs : null,
      });
      this.$router.push(entry).catch((err) => console.warn(err));
    },
    addJoinRelation() {
      if (
        this.joinPairs.some((jp) => {
          return (
            jp.first === this.datasetAColumn.key &&
            jp.second === this.datasetBColumn.key
          );
        })
      ) {
        return;
      }
      const pair = JSON.stringify({
        first: this.datasetAColumn.key,
        second: this.datasetBColumn.key,
      });
      const entry = overlayRouteEntry(this.$route, {
        joinPairs: [
          ...this.joinPairs.map((jp) => {
            return JSON.stringify(jp);
          }),
          pair,
        ],
      });
      this.$router.push(entry).catch((err) => console.warn(err));
    },
    previewJoin() {
      // flag as pending
      this.pending = true;

      const a = _.find(this.datasets, (d) => {
        return d.id === this.datasetIdA;
      });

      const b = _.find(this.datasets, (d) => {
        return d.id === this.datasetIdB;
      });
      this.datasetA = a;
      this.datasetB = b;

      // dispatch action that triggers request send to server
      datasetActions
        .joinDatasetsPreview(this.$store, {
          datasetA: a,
          datasetB: b,
          joinAccuracy: this.joinAccuracy,
          joinPairs: this.joinPairs,
        })
        .then((tableData) => {
          this.pending = false;
          this.showJoinSuccess = true;
          this.joinedPath = tableData.path;
          // sealing the return to prevent slow, unnecessary deep reactivity.
          this.previewTableData = Object.seal(tableData.data);
        })
        .catch((err) => {
          // display error modal
          this.pending = false;
          this.showJoinFailure = true;
          this.previewTableData = null;
          this.joinedPath = "";
          console.warn(err);
        });
    },
    onSave(args) {
      const datasetPreview = this.$refs.datasetPreview as InstanceType<
        typeof JoinDatasetsPreview
      >;
      datasetPreview.onSave(args);
    },
    onJoinCommitSuccess(datasetID: string) {
      loadJoinedDataset(this.$router, datasetID, this.target);
    },
    onJoinCommitFailure() {
      this.showJoinFailure = true;
    },
  },
});
</script>
<style>
@media (min-width: 1200px) {
  .modal-xlg.modal-dialog {
    width: 90% !important;
    max-width: 90% !important;
    height: 70% !important;
  }
}
</style>
<style scoped>
.bottom-margin {
  margin-bottom: 30px;
}
.join-button {
  margin: 0 8px;
  line-height: 1.5 !important;
}

.join-preview-modal .modal-dialog {
  position: relative;
  max-width: 80% !important;
  max-height: 90%;
}
.form-height {
  max-height: 40px;
}

.join-accuracy-label {
  text-align: center;
}
.check-message-container {
  display: flex;
  justify-content: flex-start;
  flex-direction: row;
  align-items: center;
}

.warning-icon {
  color: #ee0701;
}

.warning-text {
  line-height: 16px;
  font-size: 16px;
}
main {
  background-color: var(--gray-300);
  border: 1px solid var(--gray-500);
  border-radius: 0.2rem;
  display: flex;
  flex-shrink: 0; /* To avoid it to collapse and have the badges overflow. */
  flex-wrap: wrap;
  min-height: 2.5rem;
  padding: 0.2rem;
}
</style>
