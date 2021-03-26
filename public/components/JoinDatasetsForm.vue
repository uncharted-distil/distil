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
      modal-class="join-preview-modal"
      cancel-disabled
      hide-header
      hide-footer
    >
      <join-datasets-preview
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
    <search-input
      class="p-2"
      :header-title="'Join Relationships'"
    ></search-input>
    <div class="row justify-content-center">
      <b-button variant="primary" :disabled="disableJoin">
        Add Join Relationship
      </b-button>
      <b-button
        class="join-button"
        :disabled="disableJoin"
        :variant="joinVariant"
        @click="previewJoin"
      >
        <div class="row justify-content-center">
          <i class="fa fa-check-circle fa-2x mr-2" />
          <b>Join Datasets</b>
        </div>
      </b-button>
    </div>

    <div class="join-progress">
      <b-progress
        v-if="isPending"
        :value="percentComplete"
        variant="outline-secondary"
        striped
        :animated="true"
      />
    </div>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
// components
import JoinDatasetsPreview from "../components/JoinDatasetsPreview.vue";
import ErrorModal from "../components/ErrorModal.vue";
import SearchInput from "./SearchInput.vue";
import FilterBadge from "./FilterBadge.vue";
import { getters as routeGetters } from "../store/route/module";
import { Dataset, TableColumn, TableRow } from "../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../store/dataset/module";
import { Dictionary } from "../util/dict";
import { getTableDataItems, getTableDataFields } from "../util/data";
import { isJoinable } from "../util/types";
import { loadJoinedDataset } from "../util/join";

export default Vue.extend({
  name: "JoinDatasetsForm",

  components: {
    JoinDatasetsPreview,
    ErrorModal,
    SearchInput,
    FilterBadge,
  },

  props: {
    datasetIdA: String as () => string,
    datasetIdB: String as () => string,
    datasetAColumn: Object as () => TableColumn,
    datasetBColumn: Object as () => TableColumn,
    joinAccuracy: Number as () => number,
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
      return routeGetters.getRouteTargetVariable(this.$store);
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
    disableJoin(): boolean {
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
    joinedColumn(): string {
      const a = this.datasetAColumn ? this.datasetAColumn.key : "";
      const b = this.datasetBColumn ? this.datasetBColumn.key : "";
      // Note: It looks like joined column name is set to same as left column (a) name
      return a;
    },
  },

  methods: {
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
          datasetAColumn: this.datasetAColumn.key,
          datasetBColumn: this.datasetBColumn.key,
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
.join-button {
  margin: 0 8px;
  width: 35%;
  line-height: 32px !important;
}

.join-preview-modal .modal-dialog {
  position: relative;
  max-width: 80% !important;
  max-height: 90%;
}

.join-progress {
  margin: 6px 10%;
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
</style>
