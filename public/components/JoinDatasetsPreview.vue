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
  <div>
    <div class="row justify-content-center">
      <div class="check-message-container">
        <h5 class="mt-4 mb-4"><b>Join Preview</b></h5>
      </div>
    </div>

    <join-data-preview-slot
      :items="joinDataPreviewItems"
      :fields="emphasizedFields"
      :numRows="joinDataPreviewNumRows"
      :hasData="joinDataPreviewHasData"
      instance-name="join-dataset-bottom"
    ></join-data-preview-slot>

    <div class="row justify-content-center">
      <b-btn
        class="mt-3 join-modal-button"
        variant="outline-success"
        @click="commitJoin"
        :disabled="isPending"
      >
        <div class="row justify-content-center">
          <i class="fa fa-check-circle fa-2x mr-2"></i>
          <b>Commit join</b>
        </div>
      </b-btn>
      <b-btn
        class="mt-3 join-modal-button"
        variant="outline-danger"
        @click="onClose"
        :disabled="isPending"
      >
        <div class="row justify-content-center">
          <i class="fa fa-times-circle fa-2x mr-2"></i>
          <b>Cancel</b>
        </div>
      </b-btn>
    </div>

    <div class="join-progress">
      <b-progress
        v-if="isPending"
        :value="percentComplete"
        variant="outline-secondary"
        striped
        :animated="true"
      ></b-progress>
    </div>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import JoinDataPreviewSlot from "../components/JoinDataPreviewSlot";
import { createRouteEntry } from "../util/routes";
import { Dictionary } from "../util/dict";
import { getters as routeGetters } from "../store/route/module";
import {
  Dataset,
  TableData,
  TableColumn,
  TableRow,
} from "../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../store/dataset/module";
import { getTableDataItems, getTableDataFields } from "../util/data";

export default Vue.extend({
  name: "join-datasets-preview",

  components: {
    JoinDataPreviewSlot,
  },

  props: {
    datasetA: Object as () => Dataset,
    datasetB: Object as () => Dataset,
    joinedColumn: String as () => string,
    previewTableData: Object as () => TableData,
    searchResultIndex: Number as () => number,
    path: String as () => string,
  },

  data() {
    return {
      pending: false,
    };
  },

  computed: {
    terms(): string {
      return routeGetters.getRouteTerms(this.$store);
    },
    isPending(): boolean {
      return this.pending;
    },
    percentComplete(): number {
      return 100;
    },
    joinedDatasetID(): string {
      return `${this.datasetA.id}-${this.datasetB.id}`;
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
    emphasizedFields(): Dictionary<TableColumn> {
      const emphasized = {};
      _.forIn(this.joinDataPreviewFields, (field) => {
        const emph = {
          label: field.label,
          key: field.key,
          type: field.type,
          sortable: field.sortable,
          variant: null,
        };

        const isFieldSelected = field.key === this.joinedColumn;

        if (isFieldSelected) {
          emph.variant = "primary";
        }
        emphasized[field.key] = emph;
      });
      return emphasized;
    },
  },

  methods: {
    commitJoin() {
      this.pending = true;
      const importDatasetArgs = {
        datasetID: this.joinedDatasetID,
        terms: this.terms,
        source: "augmented",
        provenance: "local",
        originalDataset: this.datasetA,
        joinedDataset: this.datasetB,
        searchResultIndex: this.searchResultIndex,
        path: this.path,
      };
      datasetActions
        .importDataset(this.$store, importDatasetArgs)
        .then(() => {
          this.$emit("success", this.joinedDatasetID);
          this.pending = false;
        })
        .catch(() => {
          this.$emit("failure");
          this.pending = false;
        });
    },
    onClose() {
      this.$emit("close");
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

.join-modal-button {
  margin: 0 8px;
  width: 25% !important;
  line-height: 32px !important;
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
